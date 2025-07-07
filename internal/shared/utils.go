package shared

import (
	"archive/zip"
	"crypto/md5"
	"embedup-go/internal/cstmerr"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ResetNTPService() error {
	log.Println("Attempting to reset NTP service...")
	cmd := exec.Command("/usr/bin/sudo", "/usr/bin/systemctl", "restart", "ntp")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to restart ntp service: %v, Output: %s", err, string(output))
		return cstmerr.NewScriptError("Failed to restart ntp service", err)
	}
	log.Println("NTP service reset successfully.")
	return nil
}

func UpdateNTPService() {
	for {
		if err := ResetNTPService(); err != nil {
			log.Printf("NTP reset error (continuing): %v", err)
		} else {
			break
		}
		time.Sleep(time.Duration(300) * time.Second)
	}
}

func CheckAndCreateDir(dir string) error {
	// Ensure download_base_dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil { // 0755 gives rwx for owner, rx for group/other
			log.Printf("failed to create download base directory %s: %v \n", dir, err)
			return err
		}
	} else if err != nil {
		log.Printf("failed to check download base directory %s: %v \n", dir, err)
		return err
	}
	return nil
}

func CalculateStringMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func CalculateMD5(filePath string, n int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := md5.New()

	limitedReader := io.LimitReader(file, int64(n))

	if _, err := io.Copy(hash, limitedReader); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// UnzipFile extracts a zip file to a specified directory, skipping already extracted files.
func UnzipFile(zipFilePath string, outputDir string) error {
	log.Printf("Unzipping update from %s to %s", zipFilePath, outputDir)

	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return cstmerr.NewArchiveError(fmt.Sprintf("Failed to open zip file %s", zipFilePath), err)
	}
	defer r.Close()

	log.Printf("Archive contains %d files", len(r.File))

	for _, f := range r.File {
		outPath := filepath.Join(outputDir, f.Name)

		// Security check: Prevent "zip slip" vulnerability
		if !strings.HasPrefix(outPath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return cstmerr.NewArchiveError(fmt.Sprintf("Illegal file path in archive: %s", f.Name), nil)
		}

		// Check if the file/directory already exists and is complete
		if fileInfo, err := os.Stat(outPath); err == nil {
			// File or directory exists, now check if it's complete/correct
			if f.FileInfo().IsDir() {
				if fileInfo.IsDir() {
					log.Printf("Skipping existing directory: %s", outPath)
					continue // Directory already exists
				} else {
					// Path exists but is a file, not a directory. This is an issue.
					return cstmerr.NewFileSystemError(fmt.Sprintf("Path %s exists but is a file, expected a directory.", outPath))
				}
			} else {
				// It's a file, check its size and modification time
				if !fileInfo.IsDir() && fileInfo.Size() == f.FileInfo().Size() && fileInfo.ModTime().Truncate(time.Second).Equal(f.FileInfo().ModTime().Truncate(time.Second)) {
					log.Printf("Skipping existing and matching file: %s", outPath)
					continue // File exists and matches size and modification time
				}
				// If it exists but doesn't match, we'll overwrite it by proceeding
				log.Printf("Existing file %s found but does not match (size or mod time). Overwriting.", outPath)
			}
		} else if !os.IsNotExist(err) {
			// An error other than "does not exist" occurred while checking the file.
			return cstmerr.NewFileSystemError(fmt.Sprintf("Error checking existence of %s: %v", outPath, err))
		}

		// Proceed with extraction if the file doesn't exist or doesn't match
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
				return cstmerr.NewFileSystemError(fmt.Sprintf("Failed to create directory %s: %v", outPath, err))
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
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

		_, err = io.Copy(outFile, rc)

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

		// Set file permissions and modification time
		if f.Mode()&os.ModeSymlink == 0 {
			if err := os.Chmod(outPath, f.Mode()); err != nil {
				log.Printf("Warning: Failed to set permissions on %s: %v", outPath, err)
			}
			// Set the modification time to match the original file in the archive
			if err := os.Chtimes(outPath, time.Now(), f.FileInfo().ModTime()); err != nil {
				log.Printf("Warning: Failed to set modification time on %s: %v", outPath, err)
			}
		}
	}
	log.Println("Unzipping done.")
	return nil
}
