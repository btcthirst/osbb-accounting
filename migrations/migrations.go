// migrations/migrations.go
package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// RunMigrations виконує всі pending міграції.
func RunMigrations(db *sql.DB) error {
	runner, err := NewMigrationRunner(db)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}

	return runner.Run()
}

// GetDatabaseVersion повертає поточну версію БД.
func GetDatabaseVersion(db *sql.DB) (string, error) {
	runner, err := NewMigrationRunner(db)
	if err != nil {
		return "", err
	}

	return runner.GetCurrentVersion()
}

// ShowMigrationStatus виводить статус всіх міграцій.
func ShowMigrationStatus(db *sql.DB) error {
	runner, err := NewMigrationRunner(db)
	if err != nil {
		return err
	}

	return runner.Status()
}

// RepairMigrations відновлює некоректні записи.
func RepairMigrations(db *sql.DB) error {
	runner, err := NewMigrationRunner(db)
	if err != nil {
		return err
	}

	return runner.RepairMigrations()
}

// ResetDatabase повністю очищує БД та перезапускає міграції.
// УВАГА: Видаляє всі дані!
func ResetDatabase(db *sql.DB) error {
	log.Println("⚠ WARNING: Resetting database - ALL DATA WILL BE LOST!")

	// Видаляємо таблицю schema_migrations
	if _, err := db.Exec("DROP TABLE IF EXISTS schema_migrations"); err != nil {
		return fmt.Errorf("failed to drop schema_migrations: %w", err)
	}

	// Отримуємо список всіх таблиць
	rows, err := db.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	// Видаляємо всі таблиці
	for _, table := range tables {
		log.Printf("Dropping table: %s", table)
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	log.Println("✓ Database reset complete")

	// Запускаємо міграції заново
	return RunMigrations(db)
}
