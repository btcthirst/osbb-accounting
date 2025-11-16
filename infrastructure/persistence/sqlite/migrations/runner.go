// infrastructure/persistence/sqlite/migrations/runner.go
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
)

//go:embed *.sql
var migrationsFS embed.FS

// Migration представляє одну міграцію.
type Migration struct {
	Version     string
	Description string
	SQL         string
	Checksum    string
}

// MigrationRunner виконує міграції бази даних.
type MigrationRunner struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrationRunner створює новий runner.
func NewMigrationRunner(db *sql.DB) (*MigrationRunner, error) {
	runner := &MigrationRunner{
		db:         db,
		migrations: make([]Migration, 0),
	}

	if err := runner.loadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return runner, nil
}

// loadMigrations завантажує всі міграції з embedded файлів.
func (r *MigrationRunner) loadMigrations() error {
	entries, err := migrationsFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Ігноруємо rollback файли
		if strings.Contains(entry.Name(), "_rollback") {
			continue
		}

		// Парсимо версію з імені файлу (напр. "001_init_schema.sql" -> "001")
		version := strings.TrimSuffix(entry.Name(), ".sql")
		parts := strings.SplitN(version, "_", 2)
		if len(parts) != 2 {
			log.Printf("Warning: skipping migration with invalid name format: %s", entry.Name())
			continue
		}

		migrationVersion := parts[0]
		description := strings.ReplaceAll(parts[1], "_", " ")

		// Читаємо SQL
		sqlBytes, err := migrationsFS.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		migration := Migration{
			Version:     migrationVersion,
			Description: description,
			SQL:         string(sqlBytes),
			Checksum:    calculateChecksum(sqlBytes),
		}

		r.migrations = append(r.migrations, migration)
	}

	// Сортуємо міграції за версією
	sort.Slice(r.migrations, func(i, j int) bool {
		return r.migrations[i].Version < r.migrations[j].Version
	})

	return nil
}

// Run виконує всі pending міграції.
func (r *MigrationRunner) Run() error {
	log.Println("=== Starting database migrations ===")

	// Перевіряємо/створюємо таблицю schema_migrations
	if err := r.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// Отримуємо список застосованих міграцій
	appliedMigrations, err := r.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Перевіряємо цілісність
	if err := r.verifyMigrations(appliedMigrations); err != nil {
		return fmt.Errorf("migration integrity check failed: %w", err)
	}

	// Виконуємо pending міграції
	executed := 0
	for _, migration := range r.migrations {
		if _, applied := appliedMigrations[migration.Version]; applied {
			log.Printf("✓ Migration %s already applied: %s", migration.Version, migration.Description)
			continue
		}

		log.Printf("⚙ Applying migration %s: %s", migration.Version, migration.Description)

		if err := r.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		log.Printf("✓ Migration %s applied successfully", migration.Version)
		executed++
	}

	if executed == 0 {
		log.Println("✓ Database is up to date, no migrations needed")
	} else {
		log.Printf("✓ Successfully applied %d migration(s)", executed)
	}

	log.Println("=== Migration completed ===")
	return nil
}

// ensureMigrationsTable створює таблицю schema_migrations якщо її немає.
func (r *MigrationRunner) ensureMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL UNIQUE,
			description TEXT,
			applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			checksum TEXT
		)
	`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// getAppliedMigrations повертає map застосованих міграцій.
func (r *MigrationRunner) getAppliedMigrations() (map[string]Migration, error) {
	query := `SELECT version, description, checksum FROM schema_migrations ORDER BY version`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]Migration)
	for rows.Next() {
		var m Migration
		var checksum sql.NullString
		if err := rows.Scan(&m.Version, &m.Description, &checksum); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		if checksum.Valid {
			m.Checksum = checksum.String
		}
		applied[m.Version] = m
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migrations: %w", err)
	}

	return applied, nil
}

// verifyMigrations перевіряє цілісність застосованих міграцій.
func (r *MigrationRunner) verifyMigrations(appliedMigrations map[string]Migration) error {
	for _, migration := range r.migrations {
		applied, exists := appliedMigrations[migration.Version]
		if !exists {
			continue
		}

		// Перевіряємо checksum
		if applied.Checksum != "" && applied.Checksum != migration.Checksum {
			return fmt.Errorf(
				"migration %s has been modified after being applied (checksum mismatch: expected %s, got %s)",
				migration.Version,
				applied.Checksum,
				migration.Checksum,
			)
		}
	}

	return nil
}

// applyMigration виконує одну міграцію в транзакції.
func (r *MigrationRunner) applyMigration(migration Migration) error {
	// Починаємо транзакцію
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Виконуємо SQL міграції
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Записуємо в schema_migrations (використовуємо INSERT OR IGNORE для безпеки)
	insertQuery := `
		INSERT OR IGNORE INTO schema_migrations (version, description, checksum)
		VALUES (?, ?, ?)
	`
	result, err := tx.Exec(insertQuery, migration.Version, migration.Description, migration.Checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Перевіряємо чи був вставлений запис
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Запис вже існує - це означає що міграція була частково застосована
		// Оновлюємо checksum
		updateQuery := `UPDATE schema_migrations SET checksum = ? WHERE version = ?`
		if _, err := tx.Exec(updateQuery, migration.Checksum, migration.Version); err != nil {
			return fmt.Errorf("failed to update migration checksum: %w", err)
		}
		log.Printf("⚠ Migration %s record updated (was partially applied)", migration.Version)
	}

	// Коммітимо транзакцію
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCurrentVersion повертає поточну версію схеми БД.
func (r *MigrationRunner) GetCurrentVersion() (string, error) {
	query := `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`

	var version string
	err := r.db.QueryRow(query).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "000", nil // База ще не мігрована
		}
		return "", fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// Status виводить статус всіх міграцій.
func (r *MigrationRunner) Status() error {
	appliedMigrations, err := r.getAppliedMigrations()
	if err != nil {
		return err
	}

	log.Println("=== Migration Status ===")
	log.Printf("Total migrations: %d", len(r.migrations))
	log.Printf("Applied: %d", len(appliedMigrations))
	log.Printf("Pending: %d", len(r.migrations)-len(appliedMigrations))
	log.Println()

	for _, migration := range r.migrations {
		applied, exists := appliedMigrations[migration.Version]
		status := "[ ]"

		if exists {
			status = "[✓]"
			if applied.Checksum != "" && applied.Checksum != migration.Checksum {
				status = "[⚠]"
			}
		}

		log.Printf("%s %s: %s", status, migration.Version, migration.Description)
	}

	log.Println("=== End of Status ===")
	return nil
}

// calculateChecksum обчислює checksum для SQL.
func calculateChecksum(data []byte) string {
	if len(data) == 0 {
		return "empty"
	}
	return fmt.Sprintf("%d-%x", len(data), data[:min(32, len(data))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RepairMigrations видаляє некоректні записи з schema_migrations.
// УВАГА: Використовувати тільки для відновлення після помилок!
func (r *MigrationRunner) RepairMigrations() error {
	log.Println("⚠ WARNING: Repairing migration records...")

	appliedMigrations, err := r.getAppliedMigrations()
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Видаляємо записи для міграцій, які не існують у файлах
	for version := range appliedMigrations {
		found := false
		for _, migration := range r.migrations {
			if migration.Version == version {
				found = true
				break
			}
		}

		if !found {
			log.Printf("⚠ Removing orphaned migration record: %s", version)
			if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version); err != nil {
				return fmt.Errorf("failed to delete orphaned record: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("✓ Migration records repaired")
	return nil
}
