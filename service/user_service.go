package service

import (
	"osbb-accounting/domain"
	"osbb-accounting/repository"
)

// UserService надає бізнес-логіку для управління користувачами.
// Відокремлює логіку від UI та забезпечує перевірку прав доступу.
type UserService struct {
	userRepo    repository.UserRepository
	authService *AuthService
}

// NewUserService створює новий екземпляр сервісу користувачів.
//
// Параметри:
//   - userRepo: репозиторій для роботи з користувачами
//   - authService: сервіс аутентифікації (для роботи з паролями)
//
// Повертає:
//   - *UserService: новий екземпляр сервісу
func NewUserService(userRepo repository.UserRepository, authService *AuthService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

// CreateUser створює нового користувача.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора, який створює користувача
//   - username: ім'я користувача для входу
//   - password: пароль у відкритому вигляді (буде захешований)
//   - role: роль користувача (admin, head, accountant)
//   - fullName: повне ім'я користувача
//
// Повертає:
//   - *domain.User: створений користувач
//   - error: domain.ErrAccessDenied якщо не адмін, інші помилки
func (s *UserService) CreateUser(adminID int, username, password, role, fullName string) (*domain.User, error) {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return nil, err
	}
	if !admin.IsAdmin() {
		return nil, domain.ErrAccessDenied
	}

	// Валідація даних
	if username == "" || len(username) < 3 {
		return nil, domain.ErrUsernameTooShort
	}
	if password == "" || len(password) < 6 {
		return nil, domain.ErrPasswordTooShort
	}
	if fullName == "" {
		return nil, domain.ErrFullNameRequired
	}

	// Перевірка валідності ролі
	validRoles := map[string]bool{
		domain.RoleAdmin:      true,
		domain.RoleHead:       true,
		domain.RoleAccountant: true,
	}
	if !validRoles[role] {
		return nil, domain.ErrInvalidRole
	}

	// Хешування пароля
	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Створення користувача
	user := &domain.User{
		Username:       username,
		HashedPassword: hashedPassword,
		Role:           role,
		FullName:       fullName,
		IsActive:       true,
	}

	// Збереження в БД
	if err := s.userRepo.SaveUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers повертає список всіх користувачів.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора, який запитує список
//
// Повертає:
//   - []*domain.User: список користувачів
//   - error: domain.ErrAccessDenied якщо не адмін
func (s *UserService) GetAllUsers(adminID int) ([]*domain.User, error) {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return nil, err
	}
	if !admin.IsAdmin() {
		return nil, domain.ErrAccessDenied
	}

	return s.userRepo.FindAll()
}

// GetUserByID отримує користувача за ID.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - userID: ID користувача для отримання
//
// Повертає:
//   - *domain.User: знайдений користувач
//   - error: помилки доступу або пошуку
func (s *UserService) GetUserByID(adminID, userID int) (*domain.User, error) {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return nil, err
	}
	if !admin.IsAdmin() {
		return nil, domain.ErrAccessDenied
	}

	return s.userRepo.FindByID(userID)
}

// UpdateUser оновлює дані користувача.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - userID: ID користувача для оновлення
//   - fullName: нове повне ім'я (опціонально, "" = не змінювати)
//   - role: нова роль (опціонально, "" = не змінювати)
//
// Повертає:
//   - *domain.User: оновлений користувач
//   - error: помилки доступу або оновлення
func (s *UserService) UpdateUser(adminID, userID int, fullName, role string) (*domain.User, error) {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return nil, err
	}
	if !admin.IsAdmin() {
		return nil, domain.ErrAccessDenied
	}

	// Отримуємо користувача
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Оновлюємо поля, якщо вказано
	if fullName != "" {
		user.FullName = fullName
	}

	if role != "" {
		// Валідація ролі
		validRoles := map[string]bool{
			domain.RoleAdmin:      true,
			domain.RoleHead:       true,
			domain.RoleAccountant: true,
		}
		if !validRoles[role] {
			return nil, domain.ErrInvalidRole
		}
		user.Role = role
	}

	// Зберігаємо зміни
	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangeUserRole змінює роль користувача.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - userID: ID користувача
//   - newRole: нова роль
//
// Повертає:
//   - error: nil при успіху, помилки доступу або валідації
func (s *UserService) ChangeUserRole(adminID, userID int, newRole string) error {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return err
	}
	if !admin.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Валідація ролі
	validRoles := map[string]bool{
		domain.RoleAdmin:      true,
		domain.RoleHead:       true,
		domain.RoleAccountant: true,
	}
	if !validRoles[newRole] {
		return domain.ErrInvalidRole
	}

	// Оновлення ролі
	return s.userRepo.UpdateRole(userID, newRole)
}

// ResetUserPassword скидає пароль користувача.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - userID: ID користувача
//   - newPassword: новий пароль у відкритому вигляді
//
// Повертає:
//   - error: nil при успіху
func (s *UserService) ResetUserPassword(adminID, userID int, newPassword string) error {
	// Використовуємо метод з AuthService
	return s.authService.ResetPassword(adminID, userID, newPassword)
}

// DeactivateUser деактивує користувача (м'яке видалення).
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - userID: ID користувача для деактивації
//
// Повертає:
//   - error: nil при успіху, domain.ErrAccessDenied якщо не адмін
func (s *UserService) DeactivateUser(adminID, userID int) error {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return err
	}
	if !admin.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Заборона видалення самого себе
	if adminID == userID {
		return domain.ErrCannotDeleteSelf
	}

	return s.userRepo.DeactivateUser(userID)
}

// ActivateUser активує раніше деактивованого користувача.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - userID: ID користувача для активації
//
// Повертає:
//   - error: nil при успіху
func (s *UserService) ActivateUser(adminID, userID int) error {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return err
	}
	if !admin.IsAdmin() {
		return domain.ErrAccessDenied
	}

	return s.userRepo.ActivateUser(userID)
}

// GetUsersByRole отримує користувачів за роллю.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - adminID: ID адміністратора
//   - role: роль для фільтрації
//
// Повертає:
//   - []*domain.User: список користувачів з вказаною роллю
//   - error: помилки доступу або пошуку
func (s *UserService) GetUsersByRole(adminID int, role string) ([]*domain.User, error) {
	// Перевірка прав доступу
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return nil, err
	}
	if !admin.IsAdmin() {
		return nil, domain.ErrAccessDenied
	}

	return s.userRepo.FindByRole(role)
}
