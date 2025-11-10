// domain/entity/permission.go
package entity

import (
	"errors"
	"time"
	"unicode/utf8"
)

// Permission представляє дозвіл в системі RBAC.
// Дозволи визначають що саме може робити користувач з ресурсом.
type Permission struct {
	ID          int64
	Code        string // Унікальний код (напр. "owners.create")
	Name        string
	Description *string
	Resource    string // Ресурс (напр. "owners", "payments")
	Action      Action // Дія (create, read, update, delete, etc.)
	IsActive    bool
	CreatedAt   time.Time
}

// Action представляє тип дії над ресурсом.
type Action string

// Доступні типи дій
const (
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionApprove Action = "approve"
	ActionExport  Action = "export"
	ActionAll     Action = "all" // Для адміністраторів
)

// Validation errors
var (
	ErrPermissionCodeRequired     = errors.New("permission code is required")
	ErrPermissionCodeTooShort     = errors.New("permission code must be at least 3 characters")
	ErrPermissionCodeTooLong      = errors.New("permission code must not exceed 100 characters")
	ErrPermissionNameRequired     = errors.New("permission name is required")
	ErrPermissionResourceRequired = errors.New("permission resource is required")
	ErrPermissionResourceTooShort = errors.New("permission resource must be at least 2 characters")
	ErrPermissionActionInvalid    = errors.New("invalid permission action")
	ErrPermissionInactive         = errors.New("permission is inactive")
)

// NewPermission створює новий дозвіл з валідацією.
func NewPermission(code, name string, description *string, resource string, action Action) (*Permission, error) {
	perm := &Permission{
		Code:        code,
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := perm.Validate(); err != nil {
		return nil, err
	}

	return perm, nil
}

// Validate перевіряє коректність полів дозволу.
func (p *Permission) Validate() error {
	// Code validation
	if p.Code == "" {
		return ErrPermissionCodeRequired
	}
	codeLen := utf8.RuneCountInString(p.Code)
	if codeLen < 3 {
		return ErrPermissionCodeTooShort
	}
	if codeLen > 100 {
		return ErrPermissionCodeTooLong
	}

	// Name validation
	if p.Name == "" {
		return ErrPermissionNameRequired
	}

	// Resource validation
	if p.Resource == "" {
		return ErrPermissionResourceRequired
	}
	if utf8.RuneCountInString(p.Resource) < 2 {
		return ErrPermissionResourceTooShort
	}

	// Action validation
	if !p.Action.IsValid() {
		return ErrPermissionActionInvalid
	}

	return nil
}

// IsValid перевіряє чи є дія валідною.
func (a Action) IsValid() bool {
	switch a {
	case ActionCreate, ActionRead, ActionUpdate, ActionDelete,
		ActionApprove, ActionExport, ActionAll:
		return true
	default:
		return false
	}
}

// Deactivate деактивує дозвіл.
func (p *Permission) Deactivate() {
	p.IsActive = false
}

// Activate активує дозвіл.
func (p *Permission) Activate() {
	p.IsActive = true
}

// CanBeUsed перевіряє чи можна використовувати дозвіл.
func (p *Permission) CanBeUsed() error {
	if !p.IsActive {
		return ErrPermissionInactive
	}
	return nil
}

// Matches перевіряє чи відповідає дозвіл ресурсу та дії.
func (p *Permission) Matches(resource string, action Action) bool {
	// Дозвіл "system.all" дає доступ до всього
	if p.Resource == "system" && p.Action == ActionAll {
		return true
	}

	// Дозвіл "resource.all" дає доступ до всіх дій над ресурсом
	if p.Resource == resource && p.Action == ActionAll {
		return true
	}

	// Точна відповідність
	return p.Resource == resource && p.Action == action
}

// String повертає string representation дозволу.
func (p *Permission) String() string {
	return p.Code
}
