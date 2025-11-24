// domain/errors/errors.go
package errors

import (
	"errors"
	"fmt"
)

// Domain-level errors (Infrastructure agnostic)
var (
	// Generic errors
	ErrNotFound            = errors.New("record not found")
	ErrAlreadyExists       = errors.New("record already exists")
	ErrInvalidInput        = errors.New("invalid input")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrUnauthorized        = errors.New("unauthorized")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid username/email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUserDeleted        = errors.New("user is deleted")
	ErrSessionExpired     = errors.New("session has expired")
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidToken       = errors.New("invalid authentication token")

	// Authorization errors
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrRoleNotFound            = errors.New("role not found")
	ErrPermissionNotFound      = errors.New("permission not found")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	ErrDuplicateEntry   = errors.New("duplicate entry")

	// Concurrency errors
	ErrOptimisticLock = errors.New("record was modified by another user")
	ErrDeadlock       = errors.New("database deadlock detected")
)

// DomainError представляє доменну помилку з додатковим контекстом.
type DomainError struct {
	Code    string
	Message string
	Err     error
	Details map[string]interface{}
}

// Error реалізує інтерфейс error.
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap дозволяє використовувати errors.Is та errors.As.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// WithDetails додає деталі до помилки.
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewDomainError створює нову доменну помилку.
func NewDomainError(code, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

// Error codes для категоризації помилок
const (
	// Authentication
	CodeAuthenticationFailed = "AUTH_001"
	CodeInvalidCredentials   = "AUTH_002"
	CodeSessionExpired       = "AUTH_003"
	CodeInvalidToken         = "AUTH_004"

	// Authorization
	CodePermissionDenied        = "AUTHZ_001"
	CodeInsufficientPermissions = "AUTHZ_002"

	// Validation
	CodeValidationFailed = "VAL_001"
	CodeInvalidInput     = "VAL_002"
	CodeDuplicateEntry   = "VAL_003"

	// Data
	CodeNotFound            = "DATA_001"
	CodeAlreadyExists       = "DATA_002"
	CodeForeignKeyViolation = "DATA_003"

	// System
	CodeInternalError  = "SYS_001"
	CodeDatabaseError  = "SYS_002"
	CodeOptimisticLock = "SYS_003"

	// Business Logic
	CodeOperationNotAllowed = "BIZ_001"
)

// ValidationError представляє помилку валідації з деталями про поля.
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error реалізує інтерфейс error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError створює нову помилку валідації.
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ValidationErrors представляє множину помилок валідації.
type ValidationErrors struct {
	Errors []*ValidationError
}

// Error реалізує інтерфейс error.
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation errors"
	}
	return fmt.Sprintf("validation failed: %d error(s)", len(e.Errors))
}

// Add додає помилку валідації до списку.
func (e *ValidationErrors) Add(field, message string, value interface{}) {
	e.Errors = append(e.Errors, NewValidationError(field, message, value))
}

// HasErrors перевіряє чи є помилки.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// NewValidationErrors створює новий контейнер для помилок валідації.
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]*ValidationError, 0),
	}
}

// Is дозволяє використовувати errors.Is для доменних помилок.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As дозволяє використовувати errors.As для доменних помилок.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
