package config

import (
	"os"
	"path/filepath"
)

// AppConfig містить глобальну конфігурацію додатку.
type AppConfig struct {
	// AppName - назва додатку для відображення в UI
	AppName string

	// AppVersion - версія додатку
	AppVersion string

	// AppID - унікальний ідентифікатор додатку (для Fyne)
	AppID string

	// DataDir - директорія для зберігання даних
	DataDir string

	// DBPath - шлях до файлу бази даних
	DBPath string

	// MigrationsPath - шлях до SQL міграцій
	MigrationsPath string

	// WindowWidth - ширина вікна за замовчуванням
	WindowWidth float32

	// WindowHeight - висота вікна за замовчуванням
	WindowHeight float32

	// MinWindowWidth - мінімальна ширина вікна
	MinWindowWidth float32

	// MinWindowHeight - мінімальна висота вікна
	MinWindowHeight float32
}

// DefaultConfig повертає конфігурацію за замовчуванням.
func DefaultConfig() *AppConfig {
	// Отримуємо домашню директорію користувача
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Створюємо директорію для даних додатку
	dataDir := filepath.Join(homeDir, ".osbb-accounting")

	return &AppConfig{
		AppName:         "ОСББ Облік",
		AppVersion:      "1.0.0-MVP",
		AppID:           "com.osbb.accounting",
		DataDir:         dataDir,
		DBPath:          filepath.Join(dataDir, "osbb.db"),
		MigrationsPath:  "migrations",
		WindowWidth:     1200,
		WindowHeight:    800,
		MinWindowWidth:  1000,
		MinWindowHeight: 600,
	}
}

// EnsureDataDir створює директорію для даних, якщо вона не існує.
func (c *AppConfig) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0755)
}
