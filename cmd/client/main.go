package main

import (
	"archive/zip"
	"embedup-go/configs/config"
	apiClient "embedup-go/internal/apic"
	"embedup-go/internal/cstmerr"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func resetNTPService() error {
	log.Println("Attempting to reset NTP service...")
	// In the Rust code, sudo is used directly. Ensure your Go application
	// has the necessary permissions or is run by a user who can execute this.
	cmd := exec.Command("/usr/bin/sudo", "/usr/bin/systemctl", "restart", "ntp") //
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to restart ntp service: %v, Output: %s", err, string(output))
		return cstmerr.NewScriptError("Failed to restart ntp service", err)
	}
	log.Println("NTP service reset successfully.")
	return nil
}

func initLogging() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile) // Basic logging setup
	log.Println("Logging initialized")
}

// unzipUpdate extracts a ZIP archive to a destination directory.
// It mirrors the functionality of `unzip_update` in the Rust code.
func unzipUpdate(zipFilePath string, outputDir string) error {
	log.Printf("Unzipping update from %s to %s", zipFilePath, outputDir)

	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return cstmerr.NewArchiveError(fmt.Sprintf("Failed to open zip file %s", zipFilePath), err)
	}
	defer r.Close()

	log.Printf("Archive contains %d files", len(r.File))

	for _, f := range r.File {
		outPath := filepath.Join(outputDir, f.Name)

		// Sanitize file path to prevent directory traversal vulnerabilities
		if !strings.HasPrefix(outPath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return cstmerr.NewArchiveError(fmt.Sprintf("Illegal file path in archive: %s", f.Name), nil)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil { //
				return cstmerr.NewFileSystemError(fmt.Sprintf("Failed to create directory %s: %v", outPath, err))
			}
			continue
		}

		// Create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil { //
			return cstmerr.NewFileSystemError(fmt.Sprintf("Failed to create parent directory for %s: %v", outPath, err))
		}

		outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return cstmerr.NewFileIOError(fmt.Sprintf("Failed to create output file %s", outPath), err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return cstmerr.NewArchiveError(fmt.Sprintf("Failed to open file in archive %s", f.Name), err)
		}

		_, err = io.Copy(outFile, rc) //

		// Close files explicitly to handle errors
		closeErr1 := rc.Close()
		closeErr2 := outFile.Close()

		if err != nil {
			return cstmerr.NewFileIOError(fmt.Sprintf("Failed to copy content to %s", outPath), err)
		}
		if closeErr1 != nil {
			return cstmerr.NewArchiveError(fmt.Sprintf("Failed to close archive file entry %s", f.Name), closeErr1)
		}
		if closeErr2 != nil {
			return cstmerr.NewFileIOError(fmt.Sprintf("Failed to close output file %s", outPath), closeErr2)
		}

		// Set permissions (Unix specific, from Rust code)
		// The Rust code uses `file.unix_mode()`. Here, we use the mode from f.Mode()
		// which should be similar.
		if f.Mode()&os.ModeSymlink == 0 { // Don't chmod symlinks directly
			if err := os.Chmod(outPath, f.Mode()); err != nil { //
				log.Printf("Warning: Failed to set permissions on %s: %v", outPath, err)
				// Depending on requirements, this might be a non-fatal error.
			}
		}
	}
	log.Println("Unzipping done.")
	return nil
}

// runUpdateScript executes the provided update script.
func runUpdateScript(cfg *config.Config, scriptPath string, workingDir string) error {
	log.Printf("Running update script %s in working directory %s", scriptPath, workingDir)

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return cstmerr.NewScriptError(fmt.Sprintf("Update script not found at %s", scriptPath), err)
	}

	err := os.Chmod(scriptPath, 0755)
	if err != nil {
		return cstmerr.NewFileSystemError(fmt.Sprintf("Failed to set executable permission on script %s: %v", scriptPath, err))
	}
	log.Printf("Set executable permission on %s", scriptPath)

	cmd := exec.Command(scriptPath)
	cmd.Dir = workingDir
	// Set environment variables, specifically DB_PASSWORD as in the Rust code
	cmd.Env = append(os.Environ(), fmt.Sprintf("DB_PASSWORD=%s", cfg.DBPassword))

	output, err := cmd.CombinedOutput() // Gets both stdout and stderr

	if err != nil {
		log.Printf("Update script failed.\nStatus: %s\nSTDOUT:\n%s\nSTDERR:\n%s",
			cmd.ProcessState.String(),
			string(output),
			"")
		return cstmerr.NewScriptError(fmt.Sprintf("Update script failed.\nStatus: %s\nSTDOUT:\n%s\nSTDERR:\n%s",
			cmd.ProcessState.String(),
			string(output),
			""), err)
	}

	log.Printf("Update script executed successfully. Output:\n%s", string(output))
	return nil
}

// runUpdateCycle performs one cycle of checking for, downloading, and applying an update.
func runUpdateCycle(cfg *config.Config, apiClient *apiClient.APIClient, currentVersion int) error {
	log.Println("Starting update check cycle...")

	updateInfo, err := apiClient.CheckForUpdates()
	if err != nil {
		if apiErr, ok := err.(*cstmerr.APIRequestFailedError); ok {
			log.Printf("API request failed during update check: Status %d, Message: %s", apiErr.StatusCode, apiErr.Message)
		} else {
			log.Printf("Error checking for updates: %v", err)
		}
		// In Rust, specific errors like NoUpdateAvailable are handled.
		// Here, we might need to inspect the error type more closely if different behaviors are needed.
		// For now, any error from CheckForUpdates logs and returns.
		return fmt.Errorf("update check failed: %w", err)
	}

	log.Printf("New version available: %d, URL: %s. Current version: %d",
		updateInfo.VersionCode, updateInfo.FileURL, currentVersion) //

	if updateInfo.VersionCode > currentVersion {
		fileNameParts := strings.Split(updateInfo.FileURL, "/")
		fileNameWithExt := fileNameParts[len(fileNameParts)-1]
		// Assuming the file name from URL does not have .zip, but if it does, this needs adjustment
		// 	// The Rust code `format!("{}.zip", file_name)` suggests the URL gives a base name.
		// 	// If file_url already ends with .zip, then `fileNameWithExt` is fine.
		// 	// Let's assume file_url gives the full name like "update_package_v2.zip"
		// 	// If it's "update_package_v2", then we need to add ".zip"
		// 	// The Rust code does: `download_path.push(format!("{}.zip", file_name));`
		// 	// where file_name is `update_info.file_url.split('/').last().unwrap();`
		// 	// This implies file_name *might not* have .zip. Let's stick to the Rust logic.
		baseFileName := fileNameWithExt
		if strings.HasSuffix(strings.ToLower(baseFileName), ".zip") {
			baseFileName = baseFileName[:len(baseFileName)-4]
		}

		downloadFileName := fmt.Sprintf("%s.zip", baseFileName)
		downloadPath := filepath.Join(cfg.DownloadBaseDir, downloadFileName)

		log.Printf("Downloading update %s to %s", updateInfo.FileURL, downloadPath)
		err = apiClient.DownloadUpdate(updateInfo.FileURL, downloadPath)
		if err != nil {
			log.Printf("Error downloading update: %v", err)
			if _, ok := err.(*cstmerr.TimeoutError); ok { //
				log.Println("Download timed out, will try again sooner.")
				cfg.PollIntervalSeconds = 1 // Adjust a copy, or make cfg a pointer if it needs to be modified globally
			} else {
				cfg.PollIntervalSeconds = 300 //
			}
			// Report status on download failure
			statusMsg := fmt.Sprintf("version %d download failed: %v", updateInfo.VersionCode, err)
			if reportErr := apiClient.ReportStatus(currentVersion, statusMsg); reportErr != nil { //
				log.Printf("Failed to report download failure status: %v", reportErr)
			}
			return fmt.Errorf("download failed: %w", err)
		}
		log.Println("File downloaded successfully.")
		statusMsg := fmt.Sprintf("version %d downloaded successfully", updateInfo.VersionCode)
		if reportErr := apiClient.ReportStatus(currentVersion, statusMsg); reportErr != nil {
			log.Printf("Failed to report download success status: %v", reportErr)
		}

		extractedDirName := baseFileName
		outExtractedPath := filepath.Join(cfg.DownloadBaseDir, extractedDirName)

		log.Printf("Extracting update to %s", outExtractedPath)
		// Clean up previous extraction if it exists, or handle this in unzipUpdate
		if _, err := os.Stat(outExtractedPath); err == nil {
			log.Printf("Removing existing extraction directory: %s", outExtractedPath)
			if err := os.RemoveAll(outExtractedPath); err != nil {
				log.Printf("Failed to remove existing extraction directory %s: %v", outExtractedPath, err)
				// TODO:This could be a critical error, decide if to proceed or return
			}
		}

		if err := unzipUpdate(downloadPath, outExtractedPath); err != nil {
			log.Printf("Error unzipping file: %v", err)
			// Cleanup on unzip error as in Rust code
			if removeErr := os.Remove(downloadPath); removeErr != nil {
				log.Printf("Failed to remove downloaded zip file %s after unzip error: %v", downloadPath, removeErr)
			}
			if removeErr := os.RemoveAll(outExtractedPath); removeErr != nil {
				log.Printf("Failed to remove extraction directory %s after unzip error: %v", outExtractedPath, removeErr)
			}
			statusMsg := fmt.Sprintf("file extraction for version %d failed: %v", updateInfo.VersionCode, err)
			if reportErr := apiClient.ReportStatus(currentVersion, statusMsg); reportErr != nil {
				log.Printf("Failed to report extraction failure status: %v", reportErr)
			}
			return fmt.Errorf("unzip failed: %w", err)
		}
		log.Println("File extracted successfully.")
		statusMsg = fmt.Sprintf("file for version %d extracted successfully", updateInfo.VersionCode)
		if reportErr := apiClient.ReportStatus(currentVersion, statusMsg); reportErr != nil { //
			log.Printf("Failed to report extraction success status: %v", reportErr)
		}

		scriptPath := filepath.Join(outExtractedPath, cfg.UpdateScriptName) //
		log.Printf("Attempting to run update script: %s", scriptPath)
		if err := runUpdateScript(cfg, scriptPath, outExtractedPath); err != nil { //
			log.Printf("Update script execution failed: %v", err)
			// The Rust code calls ReportStatus here.
			if msg, ok := err.(*cstmerr.ScriptError); ok {
				statusMsg := fmt.Sprintf("update to version %d failed during script execution: %s", updateInfo.VersionCode, msg)
				if reportErr := apiClient.ReportStatus(currentVersion, statusMsg); reportErr != nil { //
					log.Printf("Failed to report script failure status: %v", reportErr)
				}
			}
			//TODO: handle role back
			return fmt.Errorf("update script failed: %w", err)
		}

		log.Printf("Update script executed successfully. System should be updated to version %d.", updateInfo.VersionCode)

		checkCurrentVersion, err := config.GetCurrentVersion(cfg)
		if err != nil {
			log.Printf("Failed to get current version (assuming 0 and continuing): %v", err)
			checkCurrentVersion = 0 // Default to 0
		}
		log.Printf("Current service version: %d", checkCurrentVersion)

		if checkCurrentVersion != updateInfo.VersionCode {
			statusMsg = fmt.Sprintf("updated successfully from %d to %d but checking the current version is %d",
				currentVersion, updateInfo.VersionCode, checkCurrentVersion)
			if reportErr := apiClient.ReportStatus(checkCurrentVersion, statusMsg); reportErr != nil {
				log.Printf("Failed to report successful update status: %v", reportErr)
			}
		} else {
			statusMsg = fmt.Sprintf("updated successfully from %d to %d", currentVersion, updateInfo.VersionCode)
			if reportErr := apiClient.ReportStatus(checkCurrentVersion, statusMsg); reportErr != nil {
				log.Printf("Failed to report successful update status: %v", reportErr)
			}
		}

		cfg.PollIntervalSeconds = 300 // Reset poll interval on successful update path
	} else {
		log.Println("No new update available or service is up-to-date.")
	}

	return nil
}

func main() {
	initLogging()
	log.Println("Embedded Updater starting...")

	configPath := os.Getenv("PODBOX_UPDATE_CONF")
	if configPath == "" {
		configPath = "/etc/podbox_update/config.toml" // Default path
	}

	appConfig, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
		return // Redundant due to Fatalf
	}
	log.Printf("Configuration loaded for service: %s", appConfig.ServiceName)

	// Ensure download_base_dir exists
	if _, err := os.Stat(appConfig.DownloadBaseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appConfig.DownloadBaseDir, 0755); err != nil { // 0755 gives rwx for owner, rx for group/other
			log.Fatalf("failed to create download base directory %s: %v", appConfig.DownloadBaseDir, err)
			return
		}
	} else if err != nil {
		log.Fatalf("failed to check download base directory %s: %v", appConfig.DownloadBaseDir, err)
		return
	}

	// Create API client
	// The Rust version gets the token from config.DeviceToken
	apiClientInstance := apiClient.New(appConfig, appConfig.DeviceToken)
	// Main update loop
	for {
		if err := resetNTPService(); err != nil {
			log.Printf("NTP reset error (continuing): %v", err)
			// Decide if this is fatal or if the loop should continue
		}

		currentVersion, err := config.GetCurrentVersion(appConfig)
		if err != nil {
			log.Printf("Failed to get current version (assuming 0 and continuing): %v", err)
			currentVersion = 0 // Default to 0
		}
		log.Printf("Current service version: %d", currentVersion)

		// 	// Create a mutable copy of config for this cycle if poll interval needs dynamic adjustment
		// 	// Or, make appConfig a pointer and modify it directly.
		// 	// For simplicity here, we'll assume cfg.PollIntervalSeconds modification in runUpdateCycle
		// 	// affects the appConfig instance if appConfig is passed by reference (as a pointer) or
		// 	// if runUpdateCycle modifies a global or shared config object.
		// 	// The Rust code passes `&mut config` to `run_update_cycle`.
		// 	// So, appConfig should be a pointer if we want to modify its fields within functions.
		// 	// Let's adjust `runUpdateCycle` to accept `*config.Config`
		// 	// and `config.Load` to return `*config.Config`. (This was done in the example config.go)

		cycleErr := runUpdateCycle(appConfig, apiClientInstance, currentVersion)
		if cycleErr != nil {
			log.Printf("Update cycle ended with error: %v", cycleErr)
			// Error recovery strategy: Rust code logs and continues.
		}

		log.Printf("Update check cycle finished. Sleeping for %d seconds.", appConfig.PollIntervalSeconds)
		time.Sleep(time.Duration(appConfig.PollIntervalSeconds) * time.Second) //
	}
}
