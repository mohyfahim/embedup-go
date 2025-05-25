package config

import (
	"bytes"
	"embedup-go/internal/cstmerr"
	"fmt"
	"log"
	"os"

	// Still useful for GetCurrentVersion
	"github.com/spf13/viper" // Import Viper
)

// Config matches the structure of your config file and environment variables.
// Viper uses mapstructure tags by default, but you can customize them.
type Config struct {
	ServiceName         string `mapstructure:"service_name"`
	CurrentVersionFile  string `mapstructure:"current_version_file"`
	UpdateCheckAPIURL   string `mapstructure:"update_check_api_url"`
	StatusReportAPIURL  string `mapstructure:"status_report_api_url"`
	PollIntervalSeconds uint64 `mapstructure:"poll_interval_seconds"`
	DownloadBaseDir     string `mapstructure:"download_base_dir"`
	DecryptionKeyHex    string `mapstructure:"decryption_key_hex"` // Kept for completeness
	UpdateScriptName    string `mapstructure:"update_script_name"`
	DBPassword          string `mapstructure:"db_password"`
	DeviceToken         string `mapstructure:"device_token"`
}

// Load reads the configuration using Viper.
// It will look for a config file (e.g., config.toml) in specified paths
// and can also read from environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values (optional, but good practice)
	v.SetDefault("service_name", "PodboxUpdateService")
	v.SetDefault("current_version_file", "/etc/podbox_update/version.txt")
	v.SetDefault("update_check_api_url", "https://localhost:8080/check_update")
	v.SetDefault("status_report_api_url", "https://localhost:8080/report_status")
	v.SetDefault("poll_interval_seconds", 300)
	v.SetDefault("download_base_dir", "/opt/updater_downloads")
	v.SetDefault("update_script_name", "update.sh")

	if configPath != "" {
		// If a specific config file path is provided, use it directly.
		v.SetConfigFile(configPath)
		v.SetConfigType("toml") // Or "json", "yaml", etc. based on your file type
	} else {
		// Otherwise, search for a config file.
		v.SetConfigName("config")               // Name of config file (without extension)
		v.SetConfigType("toml")                 // REQUIRED if the config file does not have the extension in the name
		v.AddConfigPath("/etc/podbox_update/")  // Path to look for the config file in
		v.AddConfigPath("$HOME/.podbox_update") // Call multiple times to add many search paths
		v.AddConfigPath(".")                    // Optionally look for config in the working directory
	}

	// Environment variable integration (optional but very useful)
	// Viper can automatically override config file values with environment variables.
	// E.g., PODBOX_SERVICE_NAME will override ServiceName.
	v.SetEnvPrefix("PODBOX_UPDATE") // Will be uppercased automatically
	v.AutomaticEnv()
	// You can also bind specific env vars:
	// v.BindEnv("service_name", "PODBOX_SERVICE_NAME")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if not required and rely on defaults/env
			log.Println("Config file not found, using defaults and environment variables.")
		} else {
			// Config file was found but another error was produced
			return nil, cstmerr.NewFileIOError("failed to read config file", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, cstmerr.NewConfigError("failed to unmarshal config", err)
	}

	log.Printf("Configuration loaded. Service Name: %s, Update URL: %s", config.ServiceName, config.UpdateCheckAPIURL)
	return &config, nil
}

// GetCurrentVersion reads the current version from the file specified in the config.
// This function remains largely the same as it's reading a dynamic version file,
// not a static config value typically handled by Viper at startup.
func GetCurrentVersion(cfg *Config) (int, error) {
	if _, err := os.Stat(cfg.CurrentVersionFile); os.IsNotExist(err) {
		log.Printf("Version file %s not found, assuming version 0.", cfg.CurrentVersionFile)
		return 0, nil // Default to 0 if file doesn't exist
	}

	versionData, err := os.ReadFile(cfg.CurrentVersionFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read current version file %s: %w", cfg.CurrentVersionFile, err)
	}

	var version int
	// Trim whitespace and parse
	trimmedVersionData := bytes.TrimSpace(versionData)
	if len(trimmedVersionData) == 0 {
		log.Printf("Version file %s is empty, assuming version 0.", cfg.CurrentVersionFile)
		return 0, nil
	}

	_, err = fmt.Sscanf(string(trimmedVersionData), "%d", &version)
	if err != nil {
		// The original Rust code uses ParseIntError, Sscanf gives a generic error.
		return 0, fmt.Errorf("invalid version format in version file %s ('%s'): %w", cfg.CurrentVersionFile, string(trimmedVersionData), err)
	}
	return version, nil
}

// GetDecryptionKey (if needed) would decode the hex string.
// func (c *Config) GetDecryptionKey() ([]byte, error) {
// 	return hex.DecodeString(c.DecryptionKeyHex)
// }
