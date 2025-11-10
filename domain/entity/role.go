// domain/entity/role.go
package entity

import (
	"errors"
	"time"
	"unicode/utf8"
)

// Role представляє роль користувача в системі.
// Ролі використовуються для групування дозволів.
type Role struct {
	ID          int64
	Name        string
	Description *string
	IsSystem    bool // Системні ролі не можна видаляти
	IsActive    bool
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Системні ролі (константи для типізації)
const (
	RoleAdmin      = "admin"
	RoleAccountant = "accountant"
	RoleManager    = "manager"
	RoleViewer     = "viewer"
)

// Validation errors
var (
	ErrRoleNameRequired = errors.New("role name is required")
	ErrRoleNameTooShort = errors.New("role name must be at least 2 characters")
	ErrRoleNameTooLong  = errors.New("role name must not exceed 50 characters")
	ErrRoleIsSystem     = errors.New("cannot modify system role")
	ErrRoleDeleted      = errors.New("role is deleted")
	ErrRoleInactive     = errors.New("role is inactive")
)

// NewRole створює нову роль з валідацією.
func NewRole(name string, description *string, isSystem bool) (*Role, error) {
	role := &Role{
		Name:        name,
		Description: description,
		IsSystem:    isSystem,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	return role, nil
}

// Validate перевіряє коректність полів ролі.
func (r *Role) Validate() error {
	if r.Name == "" {
		return ErrRoleNameRequired
	}

	length := utf8.RuneCountInString(r.Name)
	if length < 2 {
		return ErrRoleNameTooShort
	}
	if length > 50 {
		return ErrRoleNameTooLong
	}

	return nil
}

// Update оновлює поля ролі (тільки для не-системних).
func (r *Role) Update(name string, description *string) error {
	if r.IsSystem {
		return ErrRoleIsSystem
	}

	r.Name = name
	r.Description = description
	r.UpdatedAt = time.Now()

	return r.Validate()
}

// Deactivate деактивує роль (тільки для не-системних).
func (r *Role) Deactivate() error {
	if r.IsSystem {
		return ErrRoleIsSystem
	}

	r.IsActive = false
	r.UpdatedAt = time.Now()
	return nil
}

// Activate активує роль.
func (r *Role) Activate() {
	r.IsActive = true
	r.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення ролі (тільки для не-системних).
func (r *Role) SoftDelete() error {
	if r.IsSystem {
		return ErrRoleIsSystem
	}

	now := time.Now()
	r.DeletedAt = &now
	r.IsActive = false
	r.UpdatedAt = now
	return nil
}

// IsDeleted перевіряє чи видалена роль.
func (r *Role) IsDeleted() bool {
	return r.DeletedAt != nil
}

// CanBeAssigned перевіряє чи можна призначити роль користувачу.
func (r *Role) CanBeAssigned() error {
	if r.DeletedAt != nil {
		return ErrRoleDeleted
	}
	if !r.IsActive {
		return ErrRoleInactive
	}
	return nil
}
