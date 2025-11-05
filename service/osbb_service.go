package service

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
)

// OSBBService надає бізнес-логіку для управління ОСББ.
// Відокремлює бізнес-правила від UI та забезпечує перевірку прав доступу.
type OSBBService struct {
	osbbRepo repository.OSBBRepository
	userRepo repository.UserRepository
}

// NewOSBBService створює новий екземпляр сервісу ОСББ.
//
// Параметри:
//   - osbbRepo: репозиторій для роботи з ОСББ
//   - userRepo: репозиторій користувачів (для перевірки прав)
//
// Повертає:
//   - *OSBBService: новий екземпляр сервісу
func NewOSBBService(
	osbbRepo repository.OSBBRepository,
	userRepo repository.UserRepository,
) *OSBBService {
	return &OSBBService{
		osbbRepo: osbbRepo,
		userRepo: userRepo,
	}
}

// CreateOSBB створює організацію ОСББ в системі.
// Доступно тільки адміністраторам.
// В системі може існувати тільки одна організація ОСББ.
//
// Параметри:
//   - userID: ID користувача, який створює ОСББ
//   - osbb: дані ОСББ для створення
//
// Повертає:
//   - error: nil при успіху, domain.ErrAccessDenied якщо немає прав
func (s *OSBBService) CreateOSBB(userID int, osbb *domain.OSBB) error {
	// Перевірка прав доступу (тільки адміністратори)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Бізнес-правило: перевіряємо, чи не існує вже ОСББ
	exists, err := s.osbbRepo.Exists()
	if err != nil {
		return fmt.Errorf("помилка перевірки існування ОСББ: %w", err)
	}
	if exists {
		return domain.ErrOSBBAlreadyExists
	}

	// Валідація даних
	if err := osbb.Validate(); err != nil {
		return err
	}

	// Бізнес-правило: базовий тариф має бути розумним (5-50 грн/м²)
	if osbb.BaseRate < 5 || osbb.BaseRate > 50 {
		return fmt.Errorf("базовий тариф має бути в межах 5-50 грн/м²")
	}

	// Збереження
	return s.osbbRepo.Save(osbb)
}

// GetOSBB отримує дані ОСББ.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - *domain.OSBB: дані ОСББ
//   - error: помилки доступу або пошуку
func (s *OSBBService) GetOSBB(userID int) (*domain.OSBB, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.osbbRepo.Get()
}

// UpdateOSBB оновлює дані ОСББ.
// Доступно тільки адміністраторам та голові правління.
//
// Параметри:
//   - userID: ID користувача
//   - osbb: оновлені дані ОСББ
//
// Повертає:
//   - error: nil при успіху
func (s *OSBBService) UpdateOSBB(userID int, osbb *domain.OSBB) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.HasRole(domain.RoleHead) {
		return domain.ErrAccessDenied
	}

	// Валідація даних
	if err := osbb.Validate(); err != nil {
		return err
	}

	// Бізнес-правило: базовий тариф має бути розумним
	if osbb.BaseRate < 5 || osbb.BaseRate > 50 {
		return fmt.Errorf("базовий тариф має бути в межах 5-50 грн/м²")
	}

	return s.osbbRepo.Update(osbb)
}

// UpdateBaseRate оновлює базовий тариф ОСББ.
// Окремий метод для відстеження змін тарифів.
// Доступно адміністраторам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - newRate: новий базовий тариф (грн/м²)
//
// Повертає:
//   - error: nil при успіху
func (s *OSBBService) UpdateBaseRate(userID int, newRate float64) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Бізнес-правило: валідація тарифу
	if newRate < 5 || newRate > 50 {
		return fmt.Errorf("базовий тариф має бути в межах 5-50 грн/м²")
	}

	// Отримуємо поточні дані ОСББ
	osbb, err := s.osbbRepo.Get()
	if err != nil {
		return err
	}

	// Бізнес-правило: логування зміни тарифу
	// TODO: В майбутньому додати audit log
	oldRate := osbb.BaseRate
	if oldRate != newRate {
		// Можна додати запис в історію змін
		fmt.Printf("Зміна тарифу: %.2f → %.2f грн/м² (користувач: %s)\n",
			oldRate, newRate, user.Username)
	}

	// Оновлюємо тариф
	osbb.BaseRate = newRate
	return s.osbbRepo.Update(osbb)
}

// GetSettings отримує налаштування та статистику ОСББ.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - *domain.OSBBSettings: налаштування з статистикою
//   - error: помилки доступу
func (s *OSBBService) GetSettings(userID int) (*domain.OSBBSettings, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.osbbRepo.GetSettings()
}

// IsConfigured перевіряє, чи налаштоване ОСББ в системі.
// Корисно для відображення майстра початкового налаштування.
//
// Повертає:
//   - bool: true якщо ОСББ існує
//   - error: nil при успіху
func (s *OSBBService) IsConfigured() (bool, error) {
	return s.osbbRepo.Exists()
}

// ValidateOSBBData виконує розширену валідацію даних ОСББ.
// Додаткові бізнес-правила понад базову валідацію.
//
// Параметри:
//   - osbb: дані для валідації
//
// Повертає:
//   - error: nil якщо дані валідні
func (s *OSBBService) ValidateOSBBData(osbb *domain.OSBB) error {
	// Базова валідація з domain
	if err := osbb.Validate(); err != nil {
		return err
	}

	// Бізнес-правило: перевірка ЄДРПОУ (має бути 8 цифр)
	if len(osbb.EDRPOU) != 8 {
		return domain.ErrInvalidEDRPOU
	}

	// Бізнес-правило: перевірка формату IBAN (має починатися з UA)
	if len(osbb.BankAccount) < 29 || osbb.BankAccount[:2] != "UA" {
		return fmt.Errorf("невалідний формат IBAN (має починатися з UA)")
	}

	// Бізнес-правило: МФО має бути 6 цифр
	if len(osbb.MFO) != 6 {
		return fmt.Errorf("невалідний МФО (має бути 6 цифр)")
	}

	// Бізнес-правило: тариф має бути розумним
	if osbb.BaseRate < 5 || osbb.BaseRate > 50 {
		return fmt.Errorf("базовий тариф має бути в межах 5-50 грн/м²")
	}

	return nil
}

// CalculateTotalMonthlyCharge розраховує загальне місячне нарахування
// на основі базового тарифу та загальної площі будинку.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - float64: загальна сума місячного нарахування
//   - error: помилки доступу або розрахунку
func (s *OSBBService) CalculateTotalMonthlyCharge(userID int) (float64, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	// Отримуємо налаштування ОСББ
	settings, err := s.osbbRepo.GetSettings()
	if err != nil {
		return 0, err
	}

	// Розрахунок: базовий тариф × загальна площа
	monthlyCharge := settings.OSBB.BaseRate * settings.TotalArea

	return monthlyCharge, nil
}

// GetDisplayInfo повертає інформацію ОСББ для відображення.
// Зручний метод для UI компонентів.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - string: форматована інформація про ОСББ
//   - error: помилки доступу
func (s *OSBBService) GetDisplayInfo(userID int) (string, error) {
	osbb, err := s.GetOSBB(userID)
	if err != nil {
		return "", err
	}

	info := fmt.Sprintf(
		"%s\n"+
			"Адреса: %s\n"+
			"ЄДРПОУ: %s\n"+
			"Базовий тариф: %.2f грн/м²\n"+
			"Голова: %s (%s)",
		osbb.GetDisplayName(),
		osbb.Address,
		osbb.EDRPOU,
		osbb.BaseRate,
		osbb.ChairmanName,
		osbb.ChairmanPhone,
	)

	return info, nil
}

// CanModifySettings перевіряє, чи може користувач змінювати налаштування ОСББ.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - bool: true якщо може змінювати
//   - error: помилки перевірки
func (s *OSBBService) CanModifySettings(userID int) (bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, err
	}

	// Тільки адміністратори та голова можуть змінювати налаштування
	return user.IsAdmin() || user.HasRole(domain.RoleHead), nil
}

// GetChairmanContact повертає контактну інформацію голови правління.
// Корисно для відображення на публічних сторінках.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - string: ПІБ та телефон голови
//   - error: помилки доступу
func (s *OSBBService) GetChairmanContact(userID int) (string, error) {
	osbb, err := s.GetOSBB(userID)
	if err != nil {
		return "", err
	}

	contact := fmt.Sprintf("%s\nТелефон: %s",
		osbb.ChairmanName,
		osbb.ChairmanPhone)

	if osbb.ChairmanEmail != "" {
		contact += fmt.Sprintf("\nEmail: %s", osbb.ChairmanEmail)
	}

	return contact, nil
}

// GetBankDetails повертає банківські реквізити для платежів.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - string: форматовані банківські реквізити
//   - error: помилки доступу
func (s *OSBBService) GetBankDetails(userID int) (string, error) {
	osbb, err := s.GetOSBB(userID)
	if err != nil {
		return "", err
	}

	details := fmt.Sprintf(
		"Отримувач: %s\n"+
			"ЄДРПОУ: %s\n"+
			"Банк: %s\n"+
			"IBAN: %s\n"+
			"МФО: %s",
		osbb.Name,
		osbb.EDRPOU,
		osbb.BankName,
		osbb.BankAccount,
		osbb.MFO,
	)

	return details, nil
}
