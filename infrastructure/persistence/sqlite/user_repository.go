// infrastructure/persistence/sqlite/user_repository.go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UserRepository реалізує repository.UserRepository для SQLite.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository створює новий UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create створює нового користувача.
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			username, email, password_hash, 
			first_name, last_name, middle_name, phone,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Phone,
		boolToInt(user.IsActive),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "username") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"username already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "username")
			}
			if strings.Contains(err.Error(), "email") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"email already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "email")
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = id
	user.CreatedAt = time.Unix(now, 0)
	user.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує користувача за ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash,
			first_name, last_name, middle_name, phone,
			is_active, deleted_at, last_login_at,
			created_at, updated_at
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`

	user := &entity.User{}
	var deletedAt, lastLoginAt, createdAt, updatedAt sql.NullInt64
	var middleName, phone sql.NullString
	var isActive int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&middleName,
		&phone,
		&isActive,
		&deletedAt,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	// Конвертація nullable полів
	user.IsActive = intToBool(isActive)
	user.MiddleName = nullStringToPtr(middleName)
	user.Phone = nullStringToPtr(phone)
	user.DeletedAt = nullInt64ToTimePtr(deletedAt)
	user.LastLoginAt = nullInt64ToTimePtr(lastLoginAt)
	user.CreatedAt = time.Unix(createdAt.Int64, 0)
	user.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return user, nil
}

// GetByUsername отримує користувача за username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash,
			first_name, last_name, middle_name, phone,
			is_active, deleted_at, last_login_at,
			created_at, updated_at
		FROM users
		WHERE username = ? AND deleted_at IS NULL
	`

	return r.scanUser(ctx, query, username)
}

// GetByEmail отримує користувача за email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash,
			first_name, last_name, middle_name, phone,
			is_active, deleted_at, last_login_at,
			created_at, updated_at
		FROM users
		WHERE email = ? AND deleted_at IS NULL
	`

	return r.scanUser(ctx, query, email)
}

// GetByUsernameOrEmail отримує користувача за username або email.
func (r *UserRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*entity.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash,
			first_name, last_name, middle_name, phone,
			is_active, deleted_at, last_login_at,
			created_at, updated_at
		FROM users
		WHERE (username = ? OR email = ?) AND deleted_at IS NULL
	`

	return r.scanUser(ctx, query, usernameOrEmail, usernameOrEmail)
}

// Update оновлює дані користувача.
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET username = ?,
		    email = ?,
		    first_name = ?,
		    last_name = ?,
		    middle_name = ?,
		    phone = ?,
		    is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Phone,
		boolToInt(user.IsActive),
		now,
		user.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "username") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"username already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "username")
			}
			if strings.Contains(err.Error(), "email") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"email already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "email")
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	user.UpdatedAt = time.Unix(now, 0)

	return nil
}

// UpdatePasswordHash оновлює хеш пароля.
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, passwordHash, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update password hash: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// UpdateLastLogin оновлює час останнього входу.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET last_login_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// SoftDelete виконує м'яке видалення користувача.
func (r *UserRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE users
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// ExistsByUsername перевіряє чи існує користувач з таким username.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = ? AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByEmail перевіряє чи існує користувач з таким email.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ? AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// List отримує список користувачів з фільтрацією.
func (r *UserRepository) List(ctx context.Context, filter repository.UserFilter) ([]*entity.User, error) {
	query := `
		SELECT 
			id, username, email, password_hash,
			first_name, last_name, middle_name, phone,
			is_active, deleted_at, last_login_at,
			created_at, updated_at
		FROM users
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	// Фільтр по видаленим
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	// Фільтр по активності
	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	// Пошук по username, email, імені
	if filter.SearchQuery != "" {
		query += ` AND (
			username LIKE ? OR 
			email LIKE ? OR 
			first_name LIKE ? OR 
			last_name LIKE ?
		)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Фільтр по ролі
	if filter.RoleID != nil {
		query += ` AND id IN (
			SELECT user_id FROM user_roles WHERE role_id = ?
		)`
		args = append(args, *filter.RoleID)
	}

	// Сортування
	query += " ORDER BY created_at DESC"

	// Pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		var deletedAt, lastLoginAt, createdAt, updatedAt sql.NullInt64
		var middleName, phone sql.NullString
		var isActive int

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.FirstName,
			&user.LastName,
			&middleName,
			&phone,
			&isActive,
			&deletedAt,
			&lastLoginAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.IsActive = intToBool(isActive)
		user.MiddleName = nullStringToPtr(middleName)
		user.Phone = nullStringToPtr(phone)
		user.DeletedAt = nullInt64ToTimePtr(deletedAt)
		user.LastLoginAt = nullInt64ToTimePtr(lastLoginAt)
		user.CreatedAt = time.Unix(createdAt.Int64, 0)
		user.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Count отримує загальну кількість користувачів.
func (r *UserRepository) Count(ctx context.Context, filter repository.UserFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.SearchQuery != "" {
		query += ` AND (
			username LIKE ? OR 
			email LIKE ? OR 
			first_name LIKE ? OR 
			last_name LIKE ?
		)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if filter.RoleID != nil {
		query += ` AND id IN (
			SELECT user_id FROM user_roles WHERE role_id = ?
		)`
		args = append(args, *filter.RoleID)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// scanUser - helper для сканування одного користувача.
func (r *UserRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*entity.User, error) {
	user := &entity.User{}
	var deletedAt, lastLoginAt, createdAt, updatedAt sql.NullInt64
	var middleName, phone sql.NullString
	var isActive int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&middleName,
		&phone,
		&isActive,
		&deletedAt,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	user.IsActive = intToBool(isActive)
	user.MiddleName = nullStringToPtr(middleName)
	user.Phone = nullStringToPtr(phone)
	user.DeletedAt = nullInt64ToTimePtr(deletedAt)
	user.LastLoginAt = nullInt64ToTimePtr(lastLoginAt)
	user.CreatedAt = time.Unix(createdAt.Int64, 0)
	user.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return user, nil
}

// Helper functions для конвертації типів
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i == 1
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullInt64ToTimePtr(ni sql.NullInt64) *time.Time {
	if ni.Valid {
		t := time.Unix(ni.Int64, 0)
		return &t
	}
	return nil
}
