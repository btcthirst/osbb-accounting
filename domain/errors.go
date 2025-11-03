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

	// ErrCannotDeleteSelf повертається при спробі видалити самого себе
	ErrCannotDeleteSelf = errors.New("не можна видалити власний обліковий запис")
)

// Помилки валідації для сутності Payment.
var (
	// ErrInvalidApartmentID повертається при невалідному ID квартири
	ErrInvalidApartmentID = errors.New("невалідний ID квартири")

	// ErrInvalidPaymentType повертається при невалідному типі платежу
	ErrInvalidPaymentType = errors.New("невалідний тип платежу")

	// ErrInvalidAmount повертається при невалідній сумі
	ErrInvalidAmount = errors.New("невалідна сума платежу")

	// ErrDescriptionRequired повертається при порожньому описі
	ErrDescriptionRequired = errors.New("опис платежу є обов'язковим")

	// ErrPeriodRequired повертається при порожньому періоді
	ErrPeriodRequired = errors.New("період є обов'язковим")

	// ErrPaymentNotFound повертається, коли платіж не знайдено
	ErrPaymentNotFound = errors.New("платіж не знайдено")
)

// Помилки валідації для сутності Apartment.
var (
	// ErrApartmentNumberRequired повертається при порожньому номері квартири
	ErrApartmentNumberRequired = errors.New("номер квартири є обов'язковим")

	// ErrInvalidFloor повертається при невалідному номері поверху
	ErrInvalidFloor = errors.New("номер поверху не може бути від'ємним")

	// ErrInvalidArea повертається при невалідній площі
	ErrInvalidArea = errors.New("площа квартири має бути більше 0")

	// ErrInvalidRooms повертається при невалідній кількості кімнат
	ErrInvalidRooms = errors.New("кількість кімнат не може бути від'ємною")

	// ErrOwnerNameRequired повертається при порожньому імені власника
	ErrOwnerNameRequired = errors.New("ім'я власника є обов'язковим")

	// ErrInvalidResidentsCount повертається при невалідній кількості мешканців
	ErrInvalidResidentsCount = errors.New("кількість мешканців не може бути від'ємною")

	// ErrApartmentNotFound повертається, коли квартиру не знайдено
	ErrApartmentNotFound = errors.New("квартиру не знайдено")

	// ErrApartmentAlreadyExists повертається при спробі створити квартиру з існуючим номером
	ErrApartmentAlreadyExists = errors.New("квартира з таким номером вже існує")
)
