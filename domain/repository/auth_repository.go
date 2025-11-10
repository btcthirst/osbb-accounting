// domain/repository/auth_repository.go
package repository

import (
	"context"

	"osbb-accounting/domain/entity"
)

// UserRepository визначає контракт для роботи з користувачами.
// Цей інтерфейс належить Domain Layer і не має залежностей від інфраструктури.
type UserRepository interface {
	// Create створює нового користувача
	Create(ctx context.Context, user *entity.User) error

	// GetByID отримує користувача за ID
	GetByID(ctx context.Context, id int64) (*entity.User, error)

	// GetByUsername отримує користувача за username
	GetByUsername(ctx context.Context, username string) (*entity.User, error)

	// GetByEmail отримує користувача за email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetByUsernameOrEmail отримує користувача за username або email
	GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*entity.User, error)

	// Update оновлює дані користувача
	Update(ctx context.Context, user *entity.User) error

	// UpdatePasswordHash оновлює хеш пароля
	UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error

	// UpdateLastLogin оновлює час останнього входу
	UpdateLastLogin(ctx context.Context, userID int64) error

	// SoftDelete м'яке видалення користувача
	SoftDelete(ctx context.Context, id int64) error

	// ExistsByUsername перевіряє чи існує користувач з таким username
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail перевіряє чи існує користувач з таким email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// List отримує список користувачів з фільтрацією
	List(ctx context.Context, filter UserFilter) ([]*entity.User, error)

	// Count отримує загальну кількість користувачів
	Count(ctx context.Context, filter UserFilter) (int64, error)
}

// UserFilter - критерії пошуку користувачів
type UserFilter struct {
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string // Пошук по username, email, імені
	RoleID         *int64 // Фільтр по ролі
	Limit          int
	Offset         int
}

// SessionRepository визначає контракт для роботи з сесіями.
type SessionRepository interface {
	// Create створює нову сесію
	Create(ctx context.Context, session *entity.Session) error

	// GetByToken отримує сесію за токеном
	GetByToken(ctx context.Context, token string) (*entity.Session, error)

	// GetByID отримує сесію за ID
	GetByID(ctx context.Context, id int64) (*entity.Session, error)

	// GetActiveByUserID отримує всі активні сесії користувача
	GetActiveByUserID(ctx context.Context, userID int64) ([]*entity.Session, error)

	// Update оновлює сесію (наприклад, продовжує час життя)
	Update(ctx context.Context, session *entity.Session) error

	// Delete видаляє сесію (logout)
	Delete(ctx context.Context, id int64) error

	// DeleteByToken видаляє сесію за токеном
	DeleteByToken(ctx context.Context, token string) error

	// DeleteAllByUserID видаляє всі сесії користувача
	DeleteAllByUserID(ctx context.Context, userID int64) error

	// DeleteExpired видаляє всі закінчені сесії
	DeleteExpired(ctx context.Context) (int64, error)

	// CleanupOldSessions видаляє старі закінчені сесії (для cleanup job)
	CleanupOldSessions(ctx context.Context, olderThan int) (int64, error)
}

// RoleRepository визначає контракт для роботи з ролями.
type RoleRepository interface {
	// Create створює нову роль
	Create(ctx context.Context, role *entity.Role) error

	// GetByID отримує роль за ID
	GetByID(ctx context.Context, id int64) (*entity.Role, error)

	// GetByName отримує роль за назвою
	GetByName(ctx context.Context, name string) (*entity.Role, error)

	// GetAll отримує всі ролі
	GetAll(ctx context.Context) ([]*entity.Role, error)

	// GetByUserID отримує всі ролі користувача
	GetByUserID(ctx context.Context, userID int64) ([]*entity.Role, error)

	// Update оновлює роль
	Update(ctx context.Context, role *entity.Role) error

	// Delete видаляє роль
	Delete(ctx context.Context, id int64) error

	// AssignToUser призначає роль користувачу
	AssignToUser(ctx context.Context, userRole *entity.UserRole) error

	// RevokeFromUser відкликає роль у користувача
	RevokeFromUser(ctx context.Context, userID, roleID int64) error

	// HasRole перевіряє чи має користувач роль
	HasRole(ctx context.Context, userID, roleID int64) (bool, error)
}

// PermissionRepository визначає контракт для роботи з дозволами.
type PermissionRepository interface {
	// Create створює новий дозвіл
	Create(ctx context.Context, permission *entity.Permission) error

	// GetByID отримує дозвіл за ID
	GetByID(ctx context.Context, id int64) (*entity.Permission, error)

	// GetByCode отримує дозвіл за кодом
	GetByCode(ctx context.Context, code string) (*entity.Permission, error)

	// GetAll отримує всі дозволи
	GetAll(ctx context.Context) ([]*entity.Permission, error)

	// GetByRoleID отримує всі дозволи ролі
	GetByRoleID(ctx context.Context, roleID int64) ([]*entity.Permission, error)

	// GetByUserID отримує всі дозволи користувача (через ролі)
	GetByUserID(ctx context.Context, userID int64) ([]*entity.Permission, error)

	// Update оновлює дозвіл
	Update(ctx context.Context, permission *entity.Permission) error

	// Delete видаляє дозвіл
	Delete(ctx context.Context, id int64) error

	// AssignToRole призначає дозвіл ролі
	AssignToRole(ctx context.Context, rolePermission *entity.RolePermission) error

	// RevokeFromRole відкликає дозвіл у ролі
	RevokeFromRole(ctx context.Context, roleID, permissionID int64) error

	// HasPermission перевіряє чи має користувач конкретний дозвіл
	HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)

	// HasPermissionForResource перевіряє чи має користувач дозвіл на ресурс/дію
	HasPermissionForResource(ctx context.Context, userID int64, resource string, action entity.Action) (bool, error)
}
