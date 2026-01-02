package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `json:"path"`
}

// DefaultDatabasePath is the default database location
const DefaultDatabasePath = "~/Library/Application Support/MacWhisper/Database/main.sqlite"

// ConfigPaths returns configuration file paths in priority order
func ConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "MacWhisperTool", "config.json"),
		filepath.Join(home, "Library", "Application Support", "MacWhisperTool", "config.json"),
	}
}

// Load loads configuration from files
func Load() (*Config, error) {
	for _, path := range ConfigPaths() {
		if cfg, err := loadFromFile(path); err == nil {
			return cfg, nil
		}
	}
	return nil, nil // No config file found, use defaults
}

// loadFromFile loads config from a specific file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetDatabasePath returns the database path with priority:
// 1. --db flag (handled in cmd/root.go)
// 2. MAC_WHISPER_DB environment variable
// 3. Config file (~/.config/MacWhisperTool/config.json)
// 4. Config file (~/Library/Application Support/MacWhisperTool/config.json)
// 5. Default path
func GetDatabasePath() string {
	// Check environment variable
	if envPath := os.Getenv("MAC_WHISPER_DB"); envPath != "" {
		return envPath
	}

	// Check config files
	if cfg, err := Load(); err == nil && cfg != nil && cfg.Database.Path != "" {
		return cfg.Database.Path
	}

	// Return default
	return DefaultDatabasePath
}
