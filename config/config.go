package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration structure.
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Security SecurityConfig `yaml:"security"`
}

// AppConfig holds application-specific settings.
type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
	WindowTitle string `yaml:"window_title"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Path            string        `yaml:"path"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	BCryptCost int `yaml:"bcrypt_cost"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		App: AppConfig{
			Name:        "OSBB Accounting",
			Environment: "dev",
			Version:     "0.1.0",
			WindowTitle: "ОСББ Управління",
		},
		Database: DatabaseConfig{
			Path:            "./osbb.db",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Security: SecurityConfig{
			BCryptCost: 10, // Default cost
		},
	}
}

// Load reads the configuration from the specified path.
// If the file does not exist, it returns the default configuration and a nil error (optional behavior, or we can error).
// Here we'll return error if file exists but fails to parse, or if file doesn't exist we can return default?
// Let's stick to: Load tries to load. If file missing, return error.
// Wrapper in main can decide to use Default if Load fails due to missing file.
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default() // Start with defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}
