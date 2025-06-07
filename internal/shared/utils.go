package shared

import (
	"crypto/md5"
	"embedup-go/internal/cstmerr"
	"io"
	"log"
	"os"
	"os/exec"
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
