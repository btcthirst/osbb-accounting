package sqlite

import (
	"database/sql"
	"errors"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite драйвер
)

// SQLiteUserRepository - конкретна реалізація UserRepository для SQLite.
// Ця імплементація ІЗОЛЬОВАНА від бізнес-логіки і може бути легко замінена
// на PostgreSQL реалізацію без зміни service layer.
//
// ВАЖЛИВО: Тільки ця структура має доступ до database/sql.
// Весь інший код працює через інтерфейс repository.UserRepository.
type SQLiteUserRepository struct {
	db *sql.DB // З'єднання з БД SQLite
}

// NewSQLiteUserRepository створює новий екземпляр SQLite репозиторію.
//
// Параметри:
//   - db: активне з'єднання з SQLite БД
//
// Повертає:
//   - repository.UserRepository: інтерфейс репозиторію
func NewSQLiteUserRepository(db *sql.DB) repository.UserRepository {
	return &SQLiteUserRepository{db: db}
}

// SaveUser реалізує збереження нового користувача в SQLite БД.
func (r *SQLiteUserRepository) SaveUser(user *domain.User) error {
	// Валідація даних перед збереженням
	if err := user.Validate(); err != nil {
		return err
	}

	// SQL запит для вставки нового користувача
	query := `
		INSERT INTO users (username, hashed_password, role, full_name, created_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true

	// Виконуємо INSERT запит
	result, err := r.db.Exec(
		query,
		user.Username,
		user.HashedPassword,
		user.Role,
		user.FullName,
		now,
		now,
		user.IsActive,
	)

	if err != nil {
		// Перевірка на порушення унікальності username
		if isUniqueConstraintError(err) {
			return repository.ErrUserAlreadyExists
		}
		return err
	}

	// Отримуємо згенерований ID та присвоюємо його user
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)

	return nil
}

// FindByID реалізує пошук користувача за ID в SQLite БД.
func (r *SQLiteUserRepository) FindByID(id int) (*domain.User, error) {
	if id <= 0 {
		return nil, repository.ErrInvalidUserID
	}

	query := `
		SELECT id, username, hashed_password, role, full_name, created_at, updated_at, is_active
		FROM users
		WHERE id = ? AND is_active = 1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.HashedPassword,
		&user.Role,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// FindByUsername реалізує пошук користувача за username в SQLite БД.
func (r *SQLiteUserRepository) FindByUsername(username string) (*domain.User, error) {
	if username == "" {
		return nil, domain.ErrInvalidUsername
	}

	query := `
		SELECT id, username, hashed_password, role, full_name, created_at, updated_at, is_active
		FROM users
		WHERE username = ? AND is_active = 1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.HashedPassword,
		&user.Role,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// FindAll реалізує отримання всіх активних користувачів з SQLite БД.
func (r *SQLiteUserRepository) FindAll() ([]*domain.User, error) {
	query := `
		SELECT id, username, hashed_password, role, full_name, created_at, updated_at, is_active
		FROM users
		WHERE is_active = 1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.HashedPassword,
			&user.Role,
			&user.FullName,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IsActive,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FindByRole реалізує пошук користувачів за роллю в SQLite БД.
func (r *SQLiteUserRepository) FindByRole(role string) ([]*domain.User, error) {
	query := `
		SELECT id, username, hashed_password, role, full_name, created_at, updated_at, is_active
		FROM users
		WHERE role = ? AND is_active = 1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.HashedPassword,
			&user.Role,
			&user.FullName,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IsActive,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser реалізує оновлення користувача в SQLite БД.
func (r *SQLiteUserRepository) UpdateUser(user *domain.User) error {
	if user.ID <= 0 {
		return repository.ErrInvalidUserID
	}

	if err := user.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE users
		SET username = ?, role = ?, full_name = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		user.Username,
		user.Role,
		user.FullName,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return repository.ErrUserAlreadyExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UpdateRole реалізує оновлення ролі користувача в SQLite БД.
func (r *SQLiteUserRepository) UpdateRole(id int, role string) error {
	if id <= 0 {
		return repository.ErrInvalidUserID
	}

	// Валідація ролі
	tempUser := &domain.User{Role: role}
	if err := tempUser.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE users
		SET role = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.Exec(query, role, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UpdatePassword реалізує оновлення пароля користувача в SQLite БД.
func (r *SQLiteUserRepository) UpdatePassword(id int, hashedPassword string) error {
	if id <= 0 {
		return repository.ErrInvalidUserID
	}

	if hashedPassword == "" {
		return domain.ErrPasswordRequired
	}

	query := `
		UPDATE users
		SET hashed_password = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.Exec(query, hashedPassword, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// DeactivateUser реалізує м'яке видалення користувача в SQLite БД.
func (r *SQLiteUserRepository) DeactivateUser(id int) error {
	if id <= 0 {
		return repository.ErrInvalidUserID
	}

	query := `
		UPDATE users
		SET is_active = 0, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// ActivateUser реалізує активацію користувача в SQLite БД.
func (r *SQLiteUserRepository) ActivateUser(id int) error {
	if id <= 0 {
		return repository.ErrInvalidUserID
	}

	query := `
		UPDATE users
		SET is_active = 1, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// DeletePermanently реалізує фізичне видалення користувача з SQLite БД.
func (r *SQLiteUserRepository) DeletePermanently(id int) error {
	if id <= 0 {
		return repository.ErrInvalidUserID
	}

	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// isUniqueConstraintError перевіряє, чи є помилка порушенням унікальності.
// Для SQLite це перевірка на "UNIQUE constraint failed".
func isUniqueConstraintError(err error) bool {
	return err != nil && (err.Error() == "UNIQUE constraint failed: users.username" ||
		contains(err.Error(), "UNIQUE"))
}

// contains - допоміжна функція для перевірки підстроки.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsInside(s, substr)))
}

func containsInside(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
