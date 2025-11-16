// infrastructure/persistence/sqlite/database.go
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"osbb-accounting/infrastructure/persistence/sqlite/migrations"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Config містить конфігурацію для підключення до SQLite.
type Config struct {
	Path            string        // Шлях до файлу БД (напр. "./data/osbb.db")
	MaxOpenConns    int           // Максимальна кількість відкритих з'єднань
	MaxIdleConns    int           // Максимальна кількість idle з'єднань
	ConnMaxLifetime time.Duration // Максимальний час життя з'єднання
	ConnMaxIdleTime time.Duration // Максимальний час простою з'єднання
}

// DefaultConfig повертає конфігурацію за замовчуванням.
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	dataDir := filepath.Join(homeDir, "osbb-accounting")

	return &Config{
		Path:            filepath.Join(dataDir, "osbb.db"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// Connect створює підключення до SQLite БД та виконує міграції.
func Connect(config *Config) (*sql.DB, error) {
	// Крок 1: Створюємо директорію для БД
	dbDir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// DSN для SQLite з оптимізаціями
	dsn := fmt.Sprintf(
		"file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000&_foreign_keys=ON",
		config.Path,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Налаштування connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Перевірка підключення
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Виконання початкових PRAGMA команд
	if err := initializePragmas(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize pragmas: %w", err)
	}

	// КРИТИЧНО: Виконуємо міграції
	if err := migrations.RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// initializePragmas встановлює оптимальні PRAGMA для SQLite.
func initializePragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA cache_size = -64000;", // 64MB cache
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}

// Close закриває з'єднання з БД.
func Close(db *sql.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}

// Vacuum виконує VACUUM для дефрагментації БД.
func Vacuum(db *sql.DB) error {
	_, err := db.Exec("VACUUM;")
	return err
}

// Analyze виконує ANALYZE для оптимізації query planner.
func Analyze(db *sql.DB) error {
	_, err := db.Exec("ANALYZE;")
	return err
}

// CheckIntegrity перевіряє цілісність БД.
func CheckIntegrity(db *sql.DB) error {
	var result string
	err := db.QueryRow("PRAGMA integrity_check;").Scan(&result)
	if err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}
	return nil
}

// GetVersion отримує версію SQLite.
func GetVersion(db *sql.DB) (string, error) {
	var version string
	err := db.QueryRow("SELECT sqlite_version();").Scan(&version)
	return version, err
}

// Stats повертає статистику connection pool.
type Stats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// GetStats отримує статистику connection pool.
func GetStats(db *sql.DB) Stats {
	stats := db.Stats()
	return Stats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}
