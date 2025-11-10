// domain/valueobject/auth.go
package valueobject

import (
	"errors"
	"time"
)

// Credentials представляє облікові дані користувача для входу.
// Це Value Object, який валідується при створенні.
type Credentials struct {
	UsernameOrEmail string
	Password        string
}

var (
	ErrCredentialsEmpty         = errors.New("username/email and password are required")
	ErrCredentialsUsernameEmpty = errors.New("username or email is required")
	ErrCredentialsPasswordEmpty = errors.New("password is required")
)

// NewCredentials створює новий об'єкт облікових даних з валідацією.
func NewCredentials(usernameOrEmail, password string) (*Credentials, error) {
	if usernameOrEmail == "" {
		return nil, ErrCredentialsUsernameEmpty
	}
	if password == "" {
		return nil, ErrCredentialsPasswordEmpty
	}

	return &Credentials{
		UsernameOrEmail: usernameOrEmail,
		Password:        password,
	}, nil
}

// AuthContext представляє контекст аутентифікації (IP, User-Agent, тощо).
type AuthContext struct {
	IPAddress string
	UserAgent string
	Timestamp time.Time
}

// NewAuthContext створює новий контекст аутентифікації.
func NewAuthContext(ipAddress, userAgent string) *AuthContext {
	return &AuthContext{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}
}

// AuthResult представляє результат аутентифікації.
type AuthResult struct {
	UserID       int64
	SessionToken string
	ExpiresAt    time.Time
	Roles        []string
	Permissions  []string
}

// NewAuthResult створює результат успішної аутентифікації.
func NewAuthResult(userID int64, sessionToken string, expiresAt time.Time) *AuthResult {
	return &AuthResult{
		UserID:       userID,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
		Roles:        make([]string, 0),
		Permissions:  make([]string, 0),
	}
}

// WithRoles додає ролі до результату аутентифікації.
func (ar *AuthResult) WithRoles(roles []string) *AuthResult {
	ar.Roles = roles
	return ar
}

// WithPermissions додає дозволи до результату аутентифікації.
func (ar *AuthResult) WithPermissions(permissions []string) *AuthResult {
	ar.Permissions = permissions
	return ar
}

// IsExpired перевіряє чи закінчився токен.
func (ar *AuthResult) IsExpired() bool {
	return time.Now().After(ar.ExpiresAt)
}

// PasswordChangeRequest представляє запит на зміну пароля.
type PasswordChangeRequest struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

var (
	ErrPasswordChangeSame = errors.New("new password must be different from old password")
)

// NewPasswordChangeRequest створює запит на зміну пароля з валідацією.
func NewPasswordChangeRequest(userID int64, oldPassword, newPassword string) (*PasswordChangeRequest, error) {
	if oldPassword == "" || newPassword == "" {
		return nil, errors.New("old and new passwords are required")
	}

	if oldPassword == newPassword {
		return nil, ErrPasswordChangeSame
	}

	return &PasswordChangeRequest{
		UserID:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}, nil
}

// RegistrationRequest представляє запит на реєстрацію нового користувача.
type RegistrationRequest struct {
	Username   string
	Email      string
	Password   string
	FirstName  string
	LastName   string
	MiddleName *string
	Phone      *string
}

// NewRegistrationRequest створює запит на реєстрацію з базовою валідацією.
func NewRegistrationRequest(
	username, email, password, firstName, lastName string,
	middleName, phone *string,
) (*RegistrationRequest, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email and password are required")
	}
	if firstName == "" || lastName == "" {
		return nil, errors.New("first name and last name are required")
	}

	return &RegistrationRequest{
		Username:   username,
		Email:      email,
		Password:   password,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		Phone:      phone,
	}, nil
}
