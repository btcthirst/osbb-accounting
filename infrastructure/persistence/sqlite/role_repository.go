// infrastructure/persistence/sqlite/role_repository.go
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
)

// RoleRepository реалізує repository.RoleRepository для SQLite.
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository створює новий RoleRepository.
func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create створює нову роль.
func (r *RoleRepository) Create(ctx context.Context, role *entity.Role) error {
	query := `
		INSERT INTO roles (
			name, description, is_system, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		role.Name,
		role.Description,
		boolToInt(role.IsSystem),
		boolToInt(role.IsActive),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"role name already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "name")
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	role.ID = id
	role.CreatedAt = time.Unix(now, 0)
	role.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує роль за ID.
func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*entity.Role, error) {
	query := `
		SELECT id, name, description, is_system, is_active, 
		       deleted_at, created_at, updated_at
		FROM roles
		WHERE id = ? AND deleted_at IS NULL
	`

	role := &entity.Role{}
	var description sql.NullString
	var deletedAt, createdAt, updatedAt sql.NullInt64
	var isSystem, isActive int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&description,
		&isSystem,
		&isActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	role.Description = nullStringToPtr(description)
	role.IsSystem = intToBool(isSystem)
	role.IsActive = intToBool(isActive)
	role.DeletedAt = nullInt64ToTimePtr(deletedAt)
	role.CreatedAt = time.Unix(createdAt.Int64, 0)
	role.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return role, nil
}

// GetByName отримує роль за назвою.
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `
		SELECT id, name, description, is_system, is_active, 
		       deleted_at, created_at, updated_at
		FROM roles
		WHERE name = ? AND deleted_at IS NULL
	`

	role := &entity.Role{}
	var description sql.NullString
	var deletedAt, createdAt, updatedAt sql.NullInt64
	var isSystem, isActive int

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&description,
		&isSystem,
		&isActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	role.Description = nullStringToPtr(description)
	role.IsSystem = intToBool(isSystem)
	role.IsActive = intToBool(isActive)
	role.DeletedAt = nullInt64ToTimePtr(deletedAt)
	role.CreatedAt = time.Unix(createdAt.Int64, 0)
	role.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return role, nil
}

// GetAll отримує всі ролі.
func (r *RoleRepository) GetAll(ctx context.Context) ([]*entity.Role, error) {
	query := `
		SELECT id, name, description, is_system, is_active, 
		       deleted_at, created_at, updated_at
		FROM roles
		WHERE deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*entity.Role, 0)
	for rows.Next() {
		role := &entity.Role{}
		var description sql.NullString
		var deletedAt, createdAt, updatedAt sql.NullInt64
		var isSystem, isActive int

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&description,
			&isSystem,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role.Description = nullStringToPtr(description)
		role.IsSystem = intToBool(isSystem)
		role.IsActive = intToBool(isActive)
		role.DeletedAt = nullInt64ToTimePtr(deletedAt)
		role.CreatedAt = time.Unix(createdAt.Int64, 0)
		role.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// GetByUserID отримує всі ролі користувача.
func (r *RoleRepository) GetByUserID(ctx context.Context, userID int64) ([]*entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_system, r.is_active, 
		       r.deleted_at, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.deleted_at IS NULL AND r.is_active = 1
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*entity.Role, 0)
	for rows.Next() {
		role := &entity.Role{}
		var description sql.NullString
		var deletedAt, createdAt, updatedAt sql.NullInt64
		var isSystem, isActive int

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&description,
			&isSystem,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role.Description = nullStringToPtr(description)
		role.IsSystem = intToBool(isSystem)
		role.IsActive = intToBool(isActive)
		role.DeletedAt = nullInt64ToTimePtr(deletedAt)
		role.CreatedAt = time.Unix(createdAt.Int64, 0)
		role.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// Update оновлює роль.
func (r *RoleRepository) Update(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles
		SET name = ?, description = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL AND is_system = 0
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		role.Name,
		role.Description,
		boolToInt(role.IsActive),
		now,
		role.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"role name already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "name")
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrRoleNotFound
	}

	role.UpdatedAt = time.Unix(now, 0)

	return nil
}

// Delete видаляє роль.
func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE roles
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL AND is_system = 0
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrRoleNotFound
	}

	return nil
}

// AssignToUser призначає роль користувачу.
func (r *RoleRepository) AssignToUser(ctx context.Context, userRole *entity.UserRole) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, granted_at, granted_by)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		userRole.UserID,
		userRole.RoleID,
		userRole.GrantedAt.Unix(),
		userRole.GrantedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"user already has this role",
				domainErrors.ErrAlreadyExists,
			)
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return domainErrors.ErrForeignKeyViolation
		}
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	userRole.ID = id

	return nil
}

// RevokeFromUser відкликає роль у користувача.
func (r *RoleRepository) RevokeFromUser(ctx context.Context, userID, roleID int64) error {
	query := `DELETE FROM user_roles WHERE user_id = ? AND role_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to revoke role from user: %w", err)
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

// HasRole перевіряє чи має користувач роль.
func (r *RoleRepository) HasRole(ctx context.Context, userID, roleID int64) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM user_roles 
		WHERE user_id = ? AND role_id = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, roleID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check role: %w", err)
	}

	return count > 0, nil
}
