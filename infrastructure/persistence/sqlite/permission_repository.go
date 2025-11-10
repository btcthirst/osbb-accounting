// infrastructure/persistence/sqlite/permission_repository.go
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

// PermissionRepository реалізує repository.PermissionRepository для SQLite.
type PermissionRepository struct {
	db *sql.DB
}

// NewPermissionRepository створює новий PermissionRepository.
func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create створює новий дозвіл.
func (r *PermissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	query := `
		INSERT INTO permissions (
			code, name, description, resource, action, is_active, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		permission.Code,
		permission.Name,
		permission.Description,
		permission.Resource,
		string(permission.Action),
		boolToInt(permission.IsActive),
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"permission code already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "code")
		}
		return fmt.Errorf("failed to create permission: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	permission.ID = id
	permission.CreatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує дозвіл за ID.
func (r *PermissionRepository) GetByID(ctx context.Context, id int64) (*entity.Permission, error) {
	query := `
		SELECT id, code, name, description, resource, action, is_active, created_at
		FROM permissions
		WHERE id = ?
	`

	permission := &entity.Permission{}
	var description sql.NullString
	var action string
	var isActive int
	var createdAt int64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID,
		&permission.Code,
		&permission.Name,
		&description,
		&permission.Resource,
		&action,
		&isActive,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission by id: %w", err)
	}

	permission.Description = nullStringToPtr(description)
	permission.Action = entity.Action(action)
	permission.IsActive = intToBool(isActive)
	permission.CreatedAt = time.Unix(createdAt, 0)

	return permission, nil
}

// GetByCode отримує дозвіл за кодом.
func (r *PermissionRepository) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	query := `
		SELECT id, code, name, description, resource, action, is_active, created_at
		FROM permissions
		WHERE code = ?
	`

	permission := &entity.Permission{}
	var description sql.NullString
	var action string
	var isActive int
	var createdAt int64

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&permission.ID,
		&permission.Code,
		&permission.Name,
		&description,
		&permission.Resource,
		&action,
		&isActive,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission by code: %w", err)
	}

	permission.Description = nullStringToPtr(description)
	permission.Action = entity.Action(action)
	permission.IsActive = intToBool(isActive)
	permission.CreatedAt = time.Unix(createdAt, 0)

	return permission, nil
}

// GetAll отримує всі дозволи.
func (r *PermissionRepository) GetAll(ctx context.Context) ([]*entity.Permission, error) {
	query := `
		SELECT id, code, name, description, resource, action, is_active, created_at
		FROM permissions
		ORDER BY resource, action
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %w", err)
	}
	defer rows.Close()

	return r.scanPermissions(rows)
}

// GetByRoleID отримує всі дозволи ролі.
func (r *PermissionRepository) GetByRoleID(ctx context.Context, roleID int64) ([]*entity.Permission, error) {
	query := `
		SELECT p.id, p.code, p.name, p.description, p.resource, p.action, p.is_active, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ? AND p.is_active = 1
		ORDER BY p.resource, p.action
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	return r.scanPermissions(rows)
}

// GetByUserID отримує всі дозволи користувача (через ролі).
func (r *PermissionRepository) GetByUserID(ctx context.Context, userID int64) ([]*entity.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.code, p.name, p.description, p.resource, p.action, p.is_active, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND p.is_active = 1
		ORDER BY p.resource, p.action
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	return r.scanPermissions(rows)
}

// Update оновлює дозвіл.
func (r *PermissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	query := `
		UPDATE permissions
		SET name = ?, description = ?, is_active = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		permission.Name,
		permission.Description,
		boolToInt(permission.IsActive),
		permission.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrPermissionNotFound
	}

	return nil
}

// Delete видаляє дозвіл.
func (r *PermissionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM permissions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrPermissionNotFound
	}

	return nil
}

// AssignToRole призначає дозвіл ролі.
func (r *PermissionRepository) AssignToRole(ctx context.Context, rolePermission *entity.RolePermission) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, granted_at)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		rolePermission.RoleID,
		rolePermission.PermissionID,
		rolePermission.GrantedAt.Unix(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"role already has this permission",
				domainErrors.ErrAlreadyExists,
			)
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return domainErrors.ErrForeignKeyViolation
		}
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	rolePermission.ID = id

	return nil
}

// RevokeFromRole відкликає дозвіл у ролі.
func (r *PermissionRepository) RevokeFromRole(ctx context.Context, roleID, permissionID int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?`

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke permission from role: %w", err)
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

// HasPermission перевіряє чи має користувач конкретний дозвіл.
func (r *PermissionRepository) HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? 
		  AND (p.code = ? OR p.code = 'system.all')
		  AND p.is_active = 1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, permissionCode).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return count > 0, nil
}

// HasPermissionForResource перевіряє чи має користувач дозвіл на ресурс/дію.
func (r *PermissionRepository) HasPermissionForResource(
	ctx context.Context,
	userID int64,
	resource string,
	action entity.Action,
) (bool, error) {
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? 
		  AND p.is_active = 1
		  AND (
		      (p.resource = 'system' AND p.action = 'all') OR
		      (p.resource = ? AND p.action = 'all') OR
		      (p.resource = ? AND p.action = ?)
		  )
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, resource, resource, string(action)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check resource permission: %w", err)
	}

	return count > 0, nil
}

// scanPermissions - helper для сканування списку дозволів.
func (r *PermissionRepository) scanPermissions(rows *sql.Rows) ([]*entity.Permission, error) {
	permissions := make([]*entity.Permission, 0)

	for rows.Next() {
		permission := &entity.Permission{}
		var description sql.NullString
		var action string
		var isActive int
		var createdAt int64

		err := rows.Scan(
			&permission.ID,
			&permission.Code,
			&permission.Name,
			&description,
			&permission.Resource,
			&action,
			&isActive,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		permission.Description = nullStringToPtr(description)
		permission.Action = entity.Action(action)
		permission.IsActive = intToBool(isActive)
		permission.CreatedAt = time.Unix(createdAt, 0)

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}
