package apiclient

import (
	"embedup-go/configs/config"
	"embedup-go/internal/cstmerr"
	sharedModels "embedup-go/internal/shared"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type UpdateInfo = sharedModels.UpdateInfo
type UpdateErr = sharedModels.UpdateErr
type StatusReportPayload = sharedModels.StatusReportPayload

// APIClient holds the HTTP client and configuration.
type APIClient struct {
	client HTTPClient
	config *config.Config
	token  string
}

// New creates a new APIClient.
func New(cfg *config.Config, token string) *APIClient {
	client := NewRestyAdapter()
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
	headers := map[string]string{
		"device-token": ac.token,
	}

	// Prepare request options for the httpclient
	opts := &RequestOptions{
		Headers:       headers,
		SuccessResult: &updateInfo, // Tell the adapter to unmarshal success response here
		ErrorResult:   &apiErr,     // Tell the adapter to unmarshal error response here
	}

	// Use the httpClient interface to make the GET request
	resp, err := ac.client.Get(ac.config.UpdateCheckAPIURL, opts)
	if err != nil {
		// This 'err' is from the HTTP client adapter itself (e.g., network issue, DNS failure).
		// The adapter (e.g., RestyAdapter) should already wrap this in a cstmerr type.
		log.Printf("Error during HTTP GET for update check: %v", err)
		return nil, err // Return the error from the adapter directly
	}

	if resp.IsError() { // Check for HTTP status codes >= 400
		log.Printf("Update check API request failed with status %d: %s", resp.StatusCode, apiErr.Message)
		// If apiErr.Message is empty, use raw body
		errMsg := apiErr.Message
		if errMsg == "" {
			errMsg = string(resp.Body)
		}
		return nil, cstmerr.NewAPIRequestFailedError(resp.StatusCode, errMsg)
	}

	// If the status code is not an "error" (>=400), ensure it's a "success" (2xx).
	if !resp.IsSuccess() {
		// This catches cases like 3xx or other non-2xx codes not already caught by IsError().
		errMsg := fmt.Sprintf("API request returned an unexpected non-success status code %d. Body: %s", resp.StatusCode, string(resp.Body))
		log.Println(errMsg)
		return nil, cstmerr.NewAPIRequestFailedError(resp.StatusCode, errMsg)
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

	// Step 1: HEAD Request to get file info (size, range support)
	headOpts := &RequestOptions{} // No special options needed for this HEAD
	headResp, err := ac.client.Head(url, headOpts)
	if err != nil {
		log.Printf("HEAD request for download failed: %v", err)
		return err
	}

	if headResp.StatusCode != http.StatusOK && headResp.StatusCode != http.StatusPartialContent { // Allow 206 for potential prior partial
		// Servers might not support HEAD for ranged requests or return non-200 for other reasons
		// For simplicity here, we proceed, but in a robust client, you might handle this differently
		return cstmerr.NewHeadError(fmt.Sprintf("HEAD request failed with status: %d", headResp.StatusCode))
	}

	totalSizeStr := headResp.Headers.Get("X-Content-Length") // Or "Content-Length"
	if totalSizeStr == "" {
		totalSizeStr = headResp.Headers.Get("Content-Length")
	}
	totalSize, _ := strconv.ParseInt(totalSizeStr, 10, 64) // Error ignored for now, handle robustly

	supportsRange := headResp.Headers.Get("Accept-Ranges") == "bytes"

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
	getStreamOpts := &RequestOptions{
		Headers: make(map[string]string),
	}
	openMode := os.O_CREATE | os.O_WRONLY
	if currentOffset > 0 && supportsRange {
		log.Printf("Resuming download from offset %d", currentOffset)
		getStreamOpts.Headers["Range"] = fmt.Sprintf("bytes=%d-", currentOffset)
		openMode = os.O_APPEND | os.O_WRONLY | os.O_CREATE // Append if resuming
	} else {
		// If not resuming, or server doesn't support range, download from start and truncate
		openMode = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
		currentOffset = 0 // Reset offset as we are starting fresh or server dictates it
	}

	streamResp, err := ac.client.GetStream(url, getStreamOpts)

	if err != nil {
		return cstmerr.NewDownloadError(fmt.Sprintf("download GET request failed: %v", err))
	}
	defer streamResp.Body.Close()

	if streamResp.StatusCode != http.StatusOK && streamResp.StatusCode != http.StatusPartialContent {
		return cstmerr.NewDownloadError(fmt.Sprintf("download request failed with status: %d", streamResp.StatusCode))
	}

	// // If server sends 200 OK even when we asked for a range, it means it doesn't support/honor range for this request
	// // or it's sending the full file. We should truncate and write from beginning.
	if streamResp.StatusCode == http.StatusOK && currentOffset > 0 {
		log.Println("Server responded with 200 OK despite a Range request, assuming full file. Restarting download.")
		openMode = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
		currentOffset = 0 // Our effective offset is now 0
	}
	destFile, err := os.OpenFile(destinationPath, openMode, 0644) // 0644 is rw for owner, r for group/other
	if err != nil {
		return cstmerr.NewFileIOError(fmt.Sprintf("failed to open/create destination file %s", destinationPath), err)
	}
	defer destFile.Close()

	log.Printf("Downloading from %s to %s (offset: %d, server status: %d)", url, destinationPath, currentOffset, streamResp.StatusCode)

	bytesWritten, err := io.Copy(destFile, streamResp.Body)
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
	headers := map[string]string{
		"device-token": ac.token,
		"Content-Type": "application/json", // Explicitly set Content-Type for JSON payload
	}
	opts := &RequestOptions{
		Headers: headers,
		Body:    payload, // The adapter (RestyAdapter) will marshal this to JSON
		// No SuccessResult or ErrorResult needed if we primarily check status code
		// and use raw body for error messages, as in the original code.
	}
	resp, err := ac.client.Put(ac.config.StatusReportAPIURL, opts)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() { // Check for non-success status codes
		errorMessage := string(resp.Body)
		if errorMessage == "" {
			errorMessage = "Unknown error from API"
		}
		log.Printf("Status report API request failed with status %d: %s", resp.StatusCode, errorMessage)
		return cstmerr.NewAPIRequestFailedError(resp.StatusCode, errorMessage)
	}

	log.Println("Status report successful")
	return nil
}
