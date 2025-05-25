package apic

import (
	"embedup-go/configs/config"
	"embedup-go/internal/cstmerr"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"resty.dev/v3"
)

// UpdateInfo matches the JSON structure for update information.
type UpdateInfo struct {
	VersionCode int    `json:"versionCode"`
	FileURL     string `json:"fileUrl"`
}

// UpdateErr matches the JSON structure for API error messages.
type UpdateErr struct {
	Message string `json:"message"`
}

// StatusReportPayload matches the JSON structure for reporting status.
type StatusReportPayload struct {
	VersionCode   int    `json:"versionCode"`
	StatusMessage string `json:"statusMessage"`
}

// APIClient holds the HTTP client and configuration.
type APIClient struct {
	client *resty.Client
	config *config.Config
	token  string
}

// New creates a new APIClient.
func New(cfg *config.Config, token string) *APIClient {
	transportSettings := &resty.TransportSettings{
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 60 * time.Second,
	}
	client := resty.NewWithTransportSettings(transportSettings)
	return &APIClient{
		client: client,
		config: cfg,
		token:  token,
	}
}

// CheckForUpdates fetches update information from the API.
func (ac *APIClient) CheckForUpdates() (*UpdateInfo, error) {
	log.Printf("Checking for updates at: %s", ac.config.UpdateCheckAPIURL)
	var updateInfo UpdateInfo
	var apiErr UpdateErr // To capture error structure from API

	resp, err := ac.client.R(). // Create a new request
					SetHeader("device-token", ac.token).
					SetResult(&updateInfo).          // Tell resty to unmarshal success response into updateInfo
					SetError(&apiErr).               // Tell resty to unmarshal error response into apiErr
					Get(ac.config.UpdateCheckAPIURL) // Perform GET request

	if err != nil { // This is for network errors, DNS errors, timeouts set by client.SetTimeout(), etc.
		// Check if it's a cstmerr.TimeoutError (though SetTimeout doesn't directly return that type)
		// You might need to check err.Error() string for "context deadline exceeded" if SetTimeout is hit.
		return nil, cstmerr.NewAPIClientError(err)
	}

	if resp.IsError() { // Check for HTTP status codes >= 400
		log.Printf("Update check API request failed with status %s: %s", resp.Status(), apiErr.Message)
		// If apiErr.Message is empty, use raw body
		errMsg := apiErr.Message
		if errMsg == "" {
			errMsg = resp.String()
		}
		return nil, cstmerr.NewAPIRequestFailedError(resp.StatusCode(), errMsg)
	}

	log.Printf("Received update info: %+v", updateInfo)
	return &updateInfo, nil
}

// DownloadUpdate downloads a file from the given URL to the destination path.
// It supports resuming downloads.
func (ac *APIClient) DownloadUpdate(url string, destinationPath string) error {
	log.Printf("Attempting to download from %s to %s", url, destinationPath)

	// Ensure parent directory exists
	parentDir := filepath.Dir(destinationPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return cstmerr.NewFileSystemError(fmt.Sprintf("failed to create parent directory %s for download: %v", parentDir, err))
		}
	}

	// Step 1: Head Request (optional but good for getting size and range support early)
	headResp, err := ac.client.R().Head(url)
	if err != nil {
		return cstmerr.NewHeadError(fmt.Sprintf("HEAD request failed: %v", err))
	}
	defer headResp.Body.Close()

	if headResp.StatusCode() != http.StatusOK && headResp.StatusCode() != http.StatusPartialContent { // Allow 206 for potential prior partial
		// Servers might not support HEAD for ranged requests or return non-200 for other reasons
		// For simplicity here, we proceed, but in a robust client, you might handle this differently
		return cstmerr.NewHeadError(fmt.Sprintf("HEAD request failed with status: %d", headResp.StatusCode))
	}

	totalSizeStr := headResp.Header().Get("X-Content-Length") // Or "Content-Length"
	if totalSizeStr == "" {
		totalSizeStr = headResp.Header().Get("Content-Length")
	}
	totalSize, _ := strconv.ParseInt(totalSizeStr, 10, 64) // Error ignored for now, handle robustly

	supportsRange := false
	if acceptRanges := headResp.Header().Get("Accept-Ranges"); acceptRanges == "bytes" {
		supportsRange = true
	}
	log.Printf("File size: %d, Supports range: %t", totalSize, supportsRange)

	// STEP 2: Determine current downloaded size
	var currentOffset int64 = 0
	fileInfo, err := os.Stat(destinationPath)
	if err == nil { // File exists
		currentOffset = fileInfo.Size()
	} else if !os.IsNotExist(err) { // Some other error accessing the file
		return cstmerr.NewFileSystemError(fmt.Sprintf("failed to get metadata for existing file %s: %v", destinationPath, err))
	}
	log.Printf("Current downloaded size for file %s is %d", destinationPath, currentOffset)

	// Step 3: Compare downloaded size
	if totalSize > 0 && currentOffset >= totalSize {
		log.Printf("File %s already fully downloaded (%d bytes).", destinationPath, currentOffset)
		return nil
	}

	// Step 4: Make GET request (potentially ranged)
	req := ac.client.R()

	openMode := os.O_CREATE | os.O_WRONLY
	if currentOffset > 0 && supportsRange {
		log.Printf("Resuming download from offset %d", currentOffset)
		req.SetHeader("Range", fmt.Sprintf("bytes=%d-", currentOffset))
		openMode = os.O_APPEND | os.O_WRONLY | os.O_CREATE // Append if resuming
	} else {
		// If not resuming, or server doesn't support range, download from start and truncate
		openMode = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
		currentOffset = 0 // Reset offset as we are starting fresh or server dictates it
	}

	resp, err := req.SetDoNotParseResponse(true).Get(url)

	if err != nil {
		return cstmerr.NewDownloadError(fmt.Sprintf("download GET request failed: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusPartialContent {
		return cstmerr.NewDownloadError(fmt.Sprintf("download request failed with status: %d", resp.StatusCode()))
	}

	// // If server sends 200 OK even when we asked for a range, it means it doesn't support/honor range for this request
	// // or it's sending the full file. We should truncate and write from beginning.
	if resp.StatusCode() == http.StatusOK && currentOffset > 0 {
		log.Println("Server responded with 200 OK despite a Range request, assuming full file. Restarting download.")
		openMode = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
		currentOffset = 0 // Our effective offset is now 0
	}
	destFile, err := os.OpenFile(destinationPath, openMode, 0644) // 0644 is rw for owner, r for group/other
	if err != nil {
		return cstmerr.NewFileIOError(fmt.Sprintf("failed to open/create destination file %s", destinationPath), err)
	}
	defer destFile.Close()

	// If we are appending, ensure the file pointer is at the end.
	// This is usually the default for O_APPEND, but explicit seek can be used if needed.
	// if openMode&os.O_APPEND != 0 && currentOffset > 0 {
	// 	_, err = destFile.Seek(currentOffset, io.SeekStart) // Not strictly necessary with O_APPEND but good for clarity
	// 	if err != nil {
	// 		return cstmerr.NewFileIOError(fmt.Sprintf("failed to seek in destination file %s: %v", destinationPath, err))
	// 	}
	// }

	log.Printf("Downloading from %s to %s (offset: %d, server status: %d)", url, destinationPath, currentOffset, resp.StatusCode())

	bytesWritten, err := io.Copy(destFile, resp.RawResponse.Body)
	if err != nil {
		// Check for specific I/O errors or network interruptions during copy
		// For example, "context deadline exceeded" can indicate a timeout during the copy operation
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return cstmerr.NewTimeoutError(err)
		}
		return cstmerr.NewDownloadError(fmt.Sprintf("error reading download stream or writing to file: %v", err))
	}

	log.Printf("Downloaded %d bytes to %s. Total size on disk now: %d", bytesWritten, destinationPath, currentOffset+bytesWritten)
	log.Printf("Download complete: %s", destinationPath)
	return nil
}

// ReportStatus sends a status update to the API.
func (ac *APIClient) ReportStatus(versionCode int, statusMessage string) error {
	payload := StatusReportPayload{
		VersionCode:   versionCode,
		StatusMessage: statusMessage,
	}

	log.Printf("Reporting status: %+v to %s", payload, ac.config.StatusReportAPIURL)

	resp, err := ac.client.R().SetHeader("device-token", ac.token).SetBody(payload).Put(ac.config.StatusReportAPIURL)
	if err != nil {
		return cstmerr.NewAPIClientError(fmt.Errorf("status report request failed: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 { // Check for non-success status codes
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMessage := string(bodyBytes)
		if errorMessage == "" {
			errorMessage = "Unknown error from API"
		}
		log.Printf("Status report API request failed with status %d: %s", resp.StatusCode(), errorMessage)
		return cstmerr.NewAPIRequestFailedError(resp.StatusCode(), errorMessage)
	}

	log.Println("Status report successful")
	return nil
}
