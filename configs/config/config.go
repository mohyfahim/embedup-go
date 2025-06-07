package config

import (
	"bytes"
	"embedup-go/internal/cstmerr"
	"fmt"
	"log"
	"os"
	"time"

	// Still useful for GetCurrentVersion
	"github.com/spf13/viper" // Import Viper
)

// DatabaseConfig holds all database connection parameters.
type DatabaseConfig struct {
	Host         string        `mapstructure:"db_host"`
	Port         int           `mapstructure:"db_port"`
	User         string        `mapstructure:"db_user"`
	Password     string        `mapstructure:"db_password_conf"` // Renamed to avoid conflict with existing DBPassword
	DBName       string        `mapstructure:"db_name"`
	SSLMode      string        `mapstructure:"db_sslmode"`
	ReadTimeout  time.Duration `mapstructure:"db_read_timeout"`  // Example advanced option
	WriteTimeout time.Duration `mapstructure:"db_write_timeout"` // Example advanced option
}

// Config matches the structure of your config file and environment variables.
// Viper uses mapstructure tags by default, but you can customize them.
type Config struct {
	ServiceName         string         `mapstructure:"service_name"`
	CurrentVersionFile  string         `mapstructure:"current_version_file"`
	ContentUpdateAPIURL string         `mapstructure:"content_update_api_url"`
	UpdateCheckAPIURL   string         `mapstructure:"update_check_api_url"`
	StatusReportAPIURL  string         `mapstructure:"status_report_api_url"`
	PollIntervalSeconds uint64         `mapstructure:"poll_interval_seconds"`
	DownloadBaseDir     string         `mapstructure:"download_base_dir"`
	DecryptionKeyHex    string         `mapstructure:"decryption_key_hex"`
	UpdateScriptName    string         `mapstructure:"update_script_name"`
	DBPassword          string         `mapstructure:"db_password"`
	DeviceToken         string         `mapstructure:"device_token"`
	Database            DatabaseConfig `mapstructure:"database"`
}

// Load reads the configuration using Viper.
// It will look for a config file (e.g., config.toml) in specified paths
// and can also read from environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values for database config
	v.SetDefault("database.db_host", "localhost")
	v.SetDefault("database.db_port", 5432)
	v.SetDefault("database.db_user", "postgres")
	v.SetDefault("database.db_name", "podbox")
	v.SetDefault("database.db_sslmode", "disable") // Common default for local dev
	v.SetDefault("database.db_read_timeout", "5s")
	v.SetDefault("database.db_write_timeout", "5s")

	// Set default values (optional, but good practice)
	v.SetDefault("service_name", "PodboxUpdateService")
	v.SetDefault("current_version_file", "/etc/podbox_update/version.txt")
	v.SetDefault("update_check_api_url", "https://localhost:8080/check_update")
	v.SetDefault("status_report_api_url", "https://localhost:8080/report_status")
	v.SetDefault("poll_interval_seconds", 300)
	v.SetDefault("download_base_dir", "/opt/updater_downloads")
	v.SetDefault("update_script_name", "update.sh")

	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("toml")
	} else {
		v.SetConfigName("config")
		v.SetConfigType("toml")
		v.AddConfigPath("/etc/podbox_update/")
		v.AddConfigPath("$HOME/.podbox_update")
		v.AddConfigPath(".")
	}
	v.BindEnv("database.db_password_conf",
		"PODBOX_UPDATE_DB_PASSWORD_CONF")

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
