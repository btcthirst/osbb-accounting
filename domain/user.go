package domain

import "time"

// Константи для ролей користувачів у системі ОСББ.
// Використовуються для реалізації Role-Based Access Control (RBAC).
const (
	// RoleAdmin - повний доступ до всіх функцій системи, включаючи
	// управління користувачами та критичні налаштування
	RoleAdmin = "admin"

	// RoleHead - роль голови ОСББ, доступ до перегляду звітів,
	// управління квартирами та затвердження рішень
	RoleHead = "head"

	// RoleAccountant - роль бухгалтера, доступ до фінансових операцій,
	// введення платежів, генерації рахунків та звітів
	RoleAccountant = "accountant"
)

// User представляє користувача системи обліку ОСББ.
// Це центральна сутність (Entity) для системи аутентифікації та авторизації.
//
// Поля:
//   - ID: унікальний ідентифікатор користувача в БД (auto-increment)
//   - Username: унікальне ім'я користувача для входу в систему
//   - HashedPassword: хешований пароль за допомогою BCrypt (ніколи не зберігається в plaintext)
//   - Role: роль користувача (одна з констант Role*)
//   - FullName: повне ім'я користувача для відображення в UI
//   - CreatedAt: час створення облікового запису
//   - UpdatedAt: час останнього оновлення запису
//   - IsActive: прапорець активності (для м'якого видалення користувачів)
type User struct {
	ID             int       `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	HashedPassword string    `json:"-" db:"hashed_password"` // "-" приховує пароль у JSON
	Role           string    `json:"role" db:"role"`
	FullName       string    `json:"full_name" db:"full_name"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	IsActive       bool      `json:"is_active" db:"is_active"`
}

// HasRole перевіряє, чи має користувач вказану роль.
// Використовується для RBAC перевірок перед виконанням операцій.
//
// Параметри:
//   - role: назва ролі для перевірки (одна з констант Role*)
//
// Повертає:
//   - true, якщо роль користувача відповідає вказаній
//   - false в іншому випадку
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsAdmin перевіряє, чи має користувач права адміністратора.
// Зручний метод для швидкої перевірки адміністративних прав.
//
// Повертає:
//   - true, якщо користувач є адміністратором
//   - false в іншому випадку
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// CanManageUsers перевіряє, чи може користувач управляти іншими користувачами.
// На даний момент тільки адміністратори мають це право.
//
// Повертає:
//   - true, якщо користувач може управляти іншими користувачами
//   - false в іншому випадку
func (u *User) CanManageUsers() bool {
	return u.IsAdmin()
}

// CanManageFinances перевіряє, чи має користувач доступ до фінансових операцій.
// Бухгалтери та адміністратори мають цей доступ.
//
// Повертає:
//   - true, якщо користувач може працювати з фінансами
//   - false в іншому випадку
func (u *User) CanManageFinances() bool {
	return u.HasRole(RoleAccountant) || u.IsAdmin()
}

// Validate виконує базову валідацію полів користувача.
// Використовується перед збереженням у БД для забезпечення цілісності даних.
//
// Повертає:
//   - error з описом проблеми, якщо валідація не пройшла
//   - nil, якщо всі поля валідні
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if len(u.Username) < 3 {
		return ErrUsernameTooShort
	}
	if u.HashedPassword == "" {
		return ErrPasswordRequired
	}
	if !isValidRole(u.Role) {
		return ErrInvalidRole
	}
	if u.FullName == "" {
		return ErrFullNameRequired
	}
	return nil
}

// isValidRole перевіряє, чи є роль однією з дозволених.
//
// Параметри:
//   - role: назва ролі для перевірки
//
// Повертає:
//   - true, якщо роль валідна
//   - false в іншому випадку
func isValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleHead, RoleAccountant:
		return true
	default:
		return false
	}
}
