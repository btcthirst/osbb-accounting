package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// DBConfig містить конфігурацію для підключення до SQLite БД.
type DBConfig struct {
	// DBPath - шлях до файлу БД SQLite
	DBPath string

	// MigrationsPath - шлях до директорії з SQL міграціями
	MigrationsPath string
}

// DefaultDBConfig повертає конфігурацію БД за замовчуванням.
// Створює БД у домашній директорії користувача.
func DefaultDBConfig() *DBConfig {
	// Отримуємо домашню директорію користувача
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Якщо не вдалося, використовуємо поточну директорію
		homeDir = "."
	}

	// Створюємо директорію для даних ОСББ
	dataDir := filepath.Join(homeDir, ".osbb-accounting")

	return &DBConfig{
		DBPath:         filepath.Join(dataDir, "osbb.db"),
		MigrationsPath: "migrations",
	}
}

// InitDatabase ініціалізує з'єднання з SQLite БД та виконує міграції.
//
// Параметри:
//   - config: конфігурація БД
//
// Повертає:
//   - *sql.DB: активне з'єднання з БД
//   - error: помилка при підключенні або міграції
func InitDatabase(config *DBConfig) (*sql.DB, error) {
	// Крок 1: Створюємо директорію для БД, якщо вона не існує
	dbDir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("не вдалося створити директорію для БД: %w", err)
	}

	// Крок 2: Відкриваємо з'єднання з SQLite БД
	// Параметри:
	// - _foreign_keys=ON: увімкнення підтримки foreign keys
	// - _journal_mode=WAL: Write-Ahead Logging для кращої продуктивності
	dsn := fmt.Sprintf("%s?_foreign_keys=ON&_journal_mode=WAL", config.DBPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("не вдалося відкрити БД: %w", err)
	}

	// Крок 3: Перевіряємо з'єднання
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("не вдалося підключитися до БД: %w", err)
	}

	// Крок 4: Налаштовуємо пул з'єднань
	// SQLite працює краще з одним з'єднанням для запису
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Крок 5: Виконуємо міграції
	if err := runMigrations(db, config.MigrationsPath); err != nil {
		db.Close()
		return nil, fmt.Errorf("помилка виконання міграцій: %w", err)
	}

	// Крок 6: Створюємо дефолтного адміністратора, якщо користувачів немає
	if err := ensureDefaultAdmin(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("помилка створення адміністратора: %w", err)
	}

	return db, nil
}

// runMigrations виконує SQL міграції з вказаної директорії.
//
// Параметри:
//   - db: з'єднання з БД
//   - migrationsPath: шлях до директорії з .sql файлами
//
// Повертає:
//   - error: помилка при виконанні міграцій
func runMigrations(db *sql.DB, migrationsPath string) error {
	// Крок 1: Створюємо таблицю для відстеження виконаних міграцій
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("не вдалося створити таблицю міграцій: %w", err)
	}

	// Крок 2: Читаємо файли міграцій
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		// Якщо директорії не існує, це не критична помилка
		// (міграції можуть бути вбудовані в executable)
		return nil
	}

	// Крок 3: Виконуємо кожну міграцію по черзі
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		version := file.Name()

		// Перевіряємо, чи вже виконана ця міграція
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("помилка перевірки міграції %s: %w", version, err)
		}

		if count > 0 {
			// Міграція вже виконана, пропускаємо
			continue
		}

		// Читаємо SQL з файлу
		migrationPath := filepath.Join(migrationsPath, file.Name())
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("не вдалося прочитати міграцію %s: %w", version, err)
		}

		// Виконуємо міграцію в транзакції
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("не вдалося почати транзакцію для міграції %s: %w", version, err)
		}

		// Виконуємо SQL
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("помилка виконання міграції %s: %w", version, err)
		}

		// Записуємо, що міграція виконана
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("помилка запису міграції %s: %w", version, err)
		}

		// Коммітимо транзакцію
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("помилка комміту міграції %s: %w", version, err)
		}

		fmt.Printf("✓ Міграція %s виконана успішно\n", version)
	}

	return nil
}

// CloseDatabase коректно закриває з'єднання з БД.
//
// Параметри:
//   - db: з'єднання для закриття
//
// Повертає:
//   - error: помилка при закритті
func CloseDatabase(db *sql.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}

// ensureDefaultAdmin створює дефолтного адміністратора, якщо користувачів немає.
func ensureDefaultAdmin(db *sql.DB) error {
	// Перевіряємо, чи є користувачі в БД
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("помилка перевірки користувачів: %w", err)
	}

	// Якщо користувачі вже є, нічого не робимо
	if count > 0 {
		return nil
	}

	fmt.Println("📝 Створення дефолтного адміністратора...")

	// Імпортуємо bcrypt для хешування
	password := "admin123"
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("помилка хешування пароля: %w", err)
	}

	// Створюємо адміністратора
	query := `
		INSERT INTO users (username, hashed_password, role, full_name, is_active)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = db.Exec(query, "admin", hashedPassword, "admin", "Системний Адміністратор", true)
	if err != nil {
		return fmt.Errorf("помилка створення адміністратора: %w", err)
	}

	fmt.Println("✓ Дефолтний адміністратор створений")
	fmt.Println("  Ім'я користувача: admin")
	fmt.Println("  Пароль: admin123")
	fmt.Println("  ⚠️ ВАЖЛИВО: Змініть пароль після першого входу!")

	return nil
}

// hashPassword створює BCrypt хеш з пароля.
// Винесено в окрему функцію для уникнення циклічних імпортів.
func hashPassword(password string) (string, error) {
	// Використовуємо той самий cost, що і в AuthService
	const bcryptCost = 12

	// Імпортуємо bcrypt локально
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}
