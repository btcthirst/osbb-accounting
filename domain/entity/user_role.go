// domain/entity/user_role.go
package entity

import (
	"errors"
	"time"
)

// UserRole представляє зв'язок між користувачем та роллю.
// Один користувач може мати декілька ролей.
type UserRole struct {
	ID        int64
	UserID    int64
	RoleID    int64
	GrantedAt time.Time
	GrantedBy *int64 // ID користувача, який призначив роль
}

var (
	ErrUserRoleUserIDRequired = errors.New("user ID is required")
	ErrUserRoleRoleIDRequired = errors.New("role ID is required")
)

// NewUserRole створює новий зв'язок користувач-роль.
func NewUserRole(userID, roleID int64, grantedBy *int64) (*UserRole, error) {
	if userID <= 0 {
		return nil, ErrUserRoleUserIDRequired
	}
	if roleID <= 0 {
		return nil, ErrUserRoleRoleIDRequired
	}

	return &UserRole{
		UserID:    userID,
		RoleID:    roleID,
		GrantedAt: time.Now(),
		GrantedBy: grantedBy,
	}, nil
}

// RolePermission представляє зв'язок між роллю та дозволом.
type RolePermission struct {
	ID           int64
	RoleID       int64
	PermissionID int64
	GrantedAt    time.Time
}

var (
	ErrRolePermissionRoleIDRequired       = errors.New("role ID is required")
	ErrRolePermissionPermissionIDRequired = errors.New("permission ID is required")
)

// NewRolePermission створює новий зв'язок роль-дозвіл.
func NewRolePermission(roleID, permissionID int64) (*RolePermission, error) {
	if roleID <= 0 {
		return nil, ErrRolePermissionRoleIDRequired
	}
	if permissionID <= 0 {
		return nil, ErrRolePermissionPermissionIDRequired
	}

	return &RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		GrantedAt:    time.Now(),
	}, nil
}
