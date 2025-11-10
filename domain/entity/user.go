// domain/entity/user.go
package entity

import (
	"errors"
	"regexp"
	"time"
	"unicode/utf8"
)

// User представляє користувача системи.
// Це Domain Entity, яка містить бізнес-логіку валідації.
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string // BCrypt hash (завжди 60 символів)
	FirstName    string
	LastName     string
	MiddleName   *string // Опціональне поле
	Phone        *string // Опціональне поле
	IsActive     bool
	DeletedAt    *time.Time
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validation errors
var (
	ErrUsernameRequired      = errors.New("username is required")
	ErrUsernameTooShort      = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong       = errors.New("username must not exceed 50 characters")
	ErrUsernameInvalidFormat = errors.New("username can only contain letters, numbers, underscore and hyphen")

	ErrEmailRequired      = errors.New("email is required")
	ErrEmailInvalidFormat = errors.New("invalid email format")

	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooWeak  = errors.New("password must contain at least one uppercase, one lowercase, one number")

	ErrFirstNameRequired = errors.New("first name is required")
	ErrLastNameRequired  = errors.New("last name is required")
	ErrNameTooLong       = errors.New("name must not exceed 100 characters")

	ErrPhoneInvalidFormat = errors.New("phone must be at least 10 digits")

	ErrUserDeleted  = errors.New("user is deleted")
	ErrUserInactive = errors.New("user is inactive")
)

// Regular expressions for validation
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
)

// NewUser створює нового користувача з валідацією.
// Пароль приймається у відкритому вигляді і має бути захешований через PasswordHasher.
func NewUser(
	username, email, firstName, lastName string,
	middleName, phone *string,
) (*User, error) {
	user := &User{
		Username:   username,
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		Phone:      phone,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

// Validate перевіряє коректність всіх полів користувача.
func (u *User) Validate() error {
	// Username validation
	if u.Username == "" {
		return ErrUsernameRequired
	}
	if utf8.RuneCountInString(u.Username) < 3 {
		return ErrUsernameTooShort
	}
	if utf8.RuneCountInString(u.Username) > 50 {
		return ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(u.Username) {
		return ErrUsernameInvalidFormat
	}

	// Email validation
	if u.Email == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(u.Email) {
		return ErrEmailInvalidFormat
	}

	// FirstName validation
	if u.FirstName == "" {
		return ErrFirstNameRequired
	}
	if utf8.RuneCountInString(u.FirstName) > 100 {
		return ErrNameTooLong
	}

	// LastName validation
	if u.LastName == "" {
		return ErrLastNameRequired
	}
	if utf8.RuneCountInString(u.LastName) > 100 {
		return ErrNameTooLong
	}

	// Phone validation (optional)
	if u.Phone != nil && *u.Phone != "" {
		if !phoneRegex.MatchString(*u.Phone) {
			return ErrPhoneInvalidFormat
		}
	}

	return nil
}

// ValidatePassword перевіряє стійкість пароля (викликається перед хешуванням).
func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}

	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return ErrPasswordTooWeak
	}

	return nil
}

// SetPasswordHash встановлює захешований пароль.
// Викликається після хешування через PasswordHasher.
func (u *User) SetPasswordHash(hash string) error {
	if len(hash) != 60 { // BCrypt hash завжди 60 символів
		return errors.New("invalid password hash length")
	}
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
	return nil
}

// CanLogin перевіряє чи може користувач увійти в систему.
func (u *User) CanLogin() error {
	if u.DeletedAt != nil {
		return ErrUserDeleted
	}
	if !u.IsActive {
		return ErrUserInactive
	}
	return nil
}

// RecordLogin фіксує час останнього входу.
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// Deactivate деактивує користувача (не видаляє).
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Activate активує користувача.
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// SoftDelete виконує м'яке видалення користувача.
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.IsActive = false
	u.UpdatedAt = now
}

// IsDeleted перевіряє чи видалений користувач.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// FullName повертає повне ім'я користувача.
func (u *User) FullName() string {
	if u.MiddleName != nil && *u.MiddleName != "" {
		return u.LastName + " " + u.FirstName + " " + *u.MiddleName
	}
	return u.LastName + " " + u.FirstName
}

// Update оновлює дозволені поля користувача.
func (u *User) Update(firstName, lastName string, middleName, phone *string) error {
	u.FirstName = firstName
	u.LastName = lastName
	u.MiddleName = middleName
	u.Phone = phone
	u.UpdatedAt = time.Now()

	return u.Validate()
}

// ChangeEmail змінює email користувача з валідацією.
func (u *User) ChangeEmail(newEmail string) error {
	if newEmail == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(newEmail) {
		return ErrEmailInvalidFormat
	}

	u.Email = newEmail
	u.UpdatedAt = time.Now()
	return nil
}

// ChangeUsername змінює username користувача з валідацією.
func (u *User) ChangeUsername(newUsername string) error {
	if newUsername == "" {
		return ErrUsernameRequired
	}
	if utf8.RuneCountInString(newUsername) < 3 {
		return ErrUsernameTooShort
	}
	if utf8.RuneCountInString(newUsername) > 50 {
		return ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(newUsername) {
		return ErrUsernameInvalidFormat
	}

	u.Username = newUsername
	u.UpdatedAt = time.Now()
	return nil
}
