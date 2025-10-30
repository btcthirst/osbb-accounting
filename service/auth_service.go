package service

import (
	"osbb-accounting/domain"
	"osbb-accounting/repository"

	"golang.org/x/crypto/bcrypt"
)

// Константа для складності хешування BCrypt.
// Значення 12 забезпечує баланс між безпекою та швидкістю.
// Кожне збільшення на 1 подвоює час хешування.
const bcryptCost = 12

// AuthService надає функціонал для аутентифікації та управління паролями.
// Цей сервіс ізолює бізнес-логіку безпеки від UI та data access layers.
//
// ВАЖЛИВО: AuthService не має прямого доступу до БД. Весь доступ до даних
// здійснюється через інтерфейс repository.UserRepository, що забезпечує
// дотримання принципів Clean Architecture.
type AuthService struct {
	// userRepo - репозиторій для роботи з користувачами
	// Інтерфейс дозволяє легко підміняти реалізацію (наприклад, для тестів)
	userRepo repository.UserRepository
}

// NewAuthService створює новий екземпляр сервісу аутентифікації.
//
// Параметри:
//   - userRepo: реалізація інтерфейсу UserRepository
//
// Повертає:
//   - *AuthService: новий екземпляр сервісу
func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// HashPassword створює BCrypt хеш з plaintext пароля.
// Використовує сіль (salt), що автоматично генерується BCrypt.
//
// Параметри:
//   - password: пароль у відкритому вигляді
//
// Повертає:
//   - string: хешований пароль (safe для зберігання в БД)
//   - error: помилка при хешуванні (наприклад, пароль занадто довгий)
//
// Приклад:
//
//	hash, err := authService.HashPassword("mySecretPass123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// hash тепер містить безпечний хеш для зберігання
func (s *AuthService) HashPassword(password string) (string, error) {
	// Валідація мінімальної довжини пароля перед хешуванням
	if len(password) < 6 {
		return "", domain.ErrPasswordTooShort
	}

	// GenerateFromPassword створює хеш з автоматичною сіллю
	// bcryptCost визначає складність обчислень (захист від brute-force)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// CheckPasswordHash перевіряє, чи відповідає plaintext пароль хешу.
// Використовується при аутентифікації користувача.
//
// Параметри:
//   - password: пароль у відкритому вигляді (введений користувачем)
//   - hash: збережений BCrypt хеш з БД
//
// Повертає:
//   - bool: true якщо пароль правильний, false в іншому випадку
//
// Приклад:
//
//	isValid := authService.CheckPasswordHash("userInput123", user.HashedPassword)
//	if !isValid {
//	    return errors.New("невірний пароль")
//	}
func (s *AuthService) CheckPasswordHash(password, hash string) bool {
	// CompareHashAndPassword безпечно порівнює пароль з хешем
	// Повертає nil якщо паролі співпадають, error в іншому випадку
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Authenticate виконує повну аутентифікацію користувача.
// Перевіряє наявність користувача, валідність пароля та статус активності.
//
// Параметри:
//   - username: ім'я користувача для входу
//   - password: пароль у відкритому вигляді
//
// Повертає:
//   - *domain.User: дані аутентифікованого користувача (без пароля в JSON)
//   - error: domain.ErrInvalidCredentials при невдалій аутентифікації,
//     інші помилки при проблемах з БД
//
// Приклад використання:
//
//	user, err := authService.Authenticate("admin", "password123")
//	if err != nil {
//	    // Показати помилку входу в UI
//	    return err
//	}
//	// Користувач успішно аутентифікований, можна відкрити головне вікно
func (s *AuthService) Authenticate(username, password string) (*domain.User, error) {
	// Крок 1: Шукаємо користувача за username через репозиторій
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// Якщо користувача не знайдено, повертаємо загальну помилку
		// (не вказуємо точно, що саме не так - безпека)
		if err == repository.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		// Інші помилки БД пробрасуємо вище
		return nil, err
	}

	// Крок 2: Перевіряємо, чи активний користувач
	if !user.IsActive {
		return nil, domain.ErrInvalidCredentials
	}

	// Крок 3: Перевіряємо правильність пароля
	if !s.CheckPasswordHash(password, user.HashedPassword) {
		return nil, domain.ErrInvalidCredentials
	}

	// Крок 4: Аутентифікація успішна, повертаємо дані користувача
	// HashedPassword не буде в JSON завдяки тегу `json:"-"` в структурі
	return user, nil
}

// ValidateCredentials виконує попередню валідацію введених даних
// перед спробою аутентифікації. Це дозволяє уникнути зайвих звернень до БД.
//
// Параметри:
//   - username: ім'я користувача для перевірки
//   - password: пароль для перевірки
//
// Повертає:
//   - error: опис проблеми валідації або nil якщо все ок
func (s *AuthService) ValidateCredentials(username, password string) error {
	if username == "" {
		return domain.ErrInvalidUsername
	}
	if len(username) < 3 {
		return domain.ErrUsernameTooShort
	}
	if password == "" {
		return domain.ErrPasswordRequired
	}
	if len(password) < 6 {
		return domain.ErrPasswordTooShort
	}
	return nil
}

// ChangePassword дозволяє користувачу змінити свій пароль.
// Перевіряє старий пароль перед встановленням нового.
//
// Параметри:
//   - userID: ідентифікатор користувача
//   - oldPassword: поточний пароль (для підтвердження)
//   - newPassword: новий пароль
//
// Повертає:
//   - error: nil при успіху, помилки валідації або аутентифікації
func (s *AuthService) ChangePassword(userID int, oldPassword, newPassword string) error {
	// Крок 1: Отримуємо користувача з БД
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Крок 2: Перевіряємо старий пароль
	if !s.CheckPasswordHash(oldPassword, user.HashedPassword) {
		return domain.ErrInvalidCredentials
	}

	// Крок 3: Валідуємо новий пароль
	if len(newPassword) < 6 {
		return domain.ErrPasswordTooShort
	}

	// Крок 4: Хешуємо новий пароль
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Крок 5: Зберігаємо новий хеш в БД
	return s.userRepo.UpdatePassword(userID, newHash)
}

// ResetPassword дозволяє адміністратору скинути пароль користувача.
// Використовується коли користувач забув пароль.
//
// Параметри:
//   - adminID: ID адміністратора, який виконує операцію
//   - targetUserID: ID користувача, якому скидається пароль
//   - newPassword: новий пароль для встановлення
//
// Повертає:
//   - error: nil при успіху, domain.ErrAccessDenied якщо не адмін
func (s *AuthService) ResetPassword(adminID, targetUserID int, newPassword string) error {
	// Крок 1: Перевіряємо, чи є користувач адміністратором
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return err
	}
	if !admin.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Крок 2: Валідуємо новий пароль
	if len(newPassword) < 6 {
		return domain.ErrPasswordTooShort
	}

	// Крок 3: Хешуємо новий пароль
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Крок 4: Оновлюємо пароль цільового користувача
	return s.userRepo.UpdatePassword(targetUserID, newHash)
}
