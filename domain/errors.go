package domain

import "errors"

// Помилки валідації для сутності User.
// Ці помилки використовуються на рівні domain та service layers
// для забезпечення валідності даних перед збереженням у БД.
var (
	// ErrInvalidUsername повертається при порожньому або невалідному username
	ErrInvalidUsername = errors.New("невалідне ім'я користувача")

	// ErrUsernameTooShort повертається, якщо username менше 3 символів
	ErrUsernameTooShort = errors.New("ім'я користувача має бути не менше 3 символів")

	// ErrPasswordRequired повертається при спробі створити користувача без пароля
	ErrPasswordRequired = errors.New("пароль є обов'язковим")

	// ErrPasswordTooShort повертається, якщо plaintext пароль менше 6 символів
	ErrPasswordTooShort = errors.New("пароль має бути не менше 6 символів")

	// ErrInvalidRole повертається при невалідній ролі користувача
	ErrInvalidRole = errors.New("невалідна роль користувача")

	// ErrFullNameRequired повертається при порожньому повному імені
	ErrFullNameRequired = errors.New("повне ім'я є обов'язковим")

	// ErrInvalidCredentials повертається при невірному username або паролі
	ErrInvalidCredentials = errors.New("невірне ім'я користувача або пароль")

	// ErrUnauthorized повертається при спробі доступу без аутентифікації
	ErrUnauthorized = errors.New("необхідна аутентифікація")

	// ErrAccessDenied повертається при недостатніх правах доступу
	ErrAccessDenied = errors.New("доступ заборонено")
)
