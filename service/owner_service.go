package service

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"regexp"
)

// OwnerService надає бізнес-логіку для управління власниками квартир.
// Відокремлює бізнес-правила від UI та забезпечує перевірку прав доступу.
type OwnerService struct {
	ownerRepo   repository.OwnerRepository
	accountRepo repository.PersonalAccountRepository
	userRepo    repository.UserRepository
}

// NewOwnerService створює новий екземпляр сервісу власників.
//
// Параметри:
//   - ownerRepo: репозиторій власників
//   - accountRepo: репозиторій особистих рахунків (для статистики)
//   - userRepo: репозиторій користувачів (для перевірки прав)
//
// Повертає:
//   - *OwnerService: новий екземпляр сервісу
func NewOwnerService(
	ownerRepo repository.OwnerRepository,
	accountRepo repository.PersonalAccountRepository,
	userRepo repository.UserRepository,
) *OwnerService {
	return &OwnerService{
		ownerRepo:   ownerRepo,
		accountRepo: accountRepo,
		userRepo:    userRepo,
	}
}

// CreateOwner створює нового власника.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача, який створює власника
//   - owner: дані власника для створення
//
// Повертає:
//   - error: nil при успіху, domain.ErrAccessDenied якщо немає прав
func (s *OwnerService) CreateOwner(userID int, owner *domain.Owner) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Розширена валідація
	if err := s.ValidateOwnerData(owner); err != nil {
		return err
	}

	// Бізнес-правило: перевірка унікальності ІПН
	existingOwner, err := s.ownerRepo.FindByTaxID(owner.TaxID)
	if err == nil && existingOwner != nil {
		return domain.ErrOwnerAlreadyExists
	}

	// Збереження
	return s.ownerRepo.Save(owner)
}

// GetOwnerByID отримує власника за ID.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - ownerID: ID власника
//
// Повертає:
//   - *domain.Owner: знайдений власник
//   - error: помилки доступу або пошуку
func (s *OwnerService) GetOwnerByID(userID, ownerID int) (*domain.Owner, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.ownerRepo.FindByID(ownerID)
}

// GetOwnerByTaxID отримує власника за ІПН.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - taxID: індивідуальний податковий номер
//
// Повертає:
//   - *domain.Owner: знайдений власник
//   - error: помилки доступу або пошуку
func (s *OwnerService) GetOwnerByTaxID(userID int, taxID string) (*domain.Owner, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.ownerRepo.FindByTaxID(taxID)
}

// GetAllOwners повертає список всіх активних власників.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - []*domain.Owner: список власників
//   - error: помилки доступу
func (s *OwnerService) GetAllOwners(userID int) ([]*domain.Owner, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.ownerRepo.FindAll()
}

// SearchOwners шукає власників за ПІБ або ІПН.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - query: пошуковий запит
//
// Повертає:
//   - []*domain.Owner: знайдені власники
//   - error: помилки доступу
func (s *OwnerService) SearchOwners(userID int, query string) ([]*domain.Owner, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Бізнес-правило: мінімальна довжина запиту
	if len(query) < 2 {
		return nil, fmt.Errorf("пошуковий запит має бути не менше 2 символів")
	}

	return s.ownerRepo.Search(query)
}

// UpdateOwner оновлює дані власника.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - owner: оновлені дані власника
//
// Повертає:
//   - error: nil при успіху
func (s *OwnerService) UpdateOwner(userID int, owner *domain.Owner) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Розширена валідація
	if err := s.ValidateOwnerData(owner); err != nil {
		return err
	}

	// Бізнес-правило: перевірка існування власника
	existing, err := s.ownerRepo.FindByID(owner.ID)
	if err != nil {
		return err
	}

	// Бізнес-правило: якщо змінюється ІПН, перевіряємо унікальність
	if existing.TaxID != owner.TaxID {
		duplicate, err := s.ownerRepo.FindByTaxID(owner.TaxID)
		if err == nil && duplicate != nil && duplicate.ID != owner.ID {
			return domain.ErrOwnerAlreadyExists
		}
	}

	return s.ownerRepo.Update(owner)
}

// DeactivateOwner деактивує власника (м'яке видалення).
// Доступно тільки адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//   - ownerID: ID власника для деактивації
//
// Повертає:
//   - error: nil при успіху
func (s *OwnerService) DeactivateOwner(userID, ownerID int) error {
	// Перевірка прав доступу (тільки адміни)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Бізнес-правило: перевірка існування власника
	owner, err := s.ownerRepo.FindByID(ownerID)
	if err != nil {
		return err
	}

	// Бізнес-правило: перевірка, чи немає активних рахунків з боргом
	accounts, err := s.accountRepo.FindByOwnerID(ownerID)
	if err != nil {
		return fmt.Errorf("помилка перевірки рахунків: %w", err)
	}

	for _, account := range accounts {
		if account.IsActive && account.HasDebt() {
			return fmt.Errorf(
				"не можна деактивувати власника з непогашеним боргом (рахунок: %s, борг: %.2f грн)",
				account.AccountNumber,
				account.GetDebtAmount(),
			)
		}
	}

	// Попередження користувача про кількість рахунків
	if len(accounts) > 0 {
		fmt.Printf("⚠️ Увага: власник %s має %d активних рахунків\n",
			owner.FullName, len(accounts))
	}

	return s.ownerRepo.Deactivate(ownerID)
}

// ActivateOwner активує раніше деактивованого власника.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - ownerID: ID власника для активації
//
// Повертає:
//   - error: nil при успіху
func (s *OwnerService) ActivateOwner(userID, ownerID int) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	return s.ownerRepo.Activate(ownerID)
}

// GetOwnerWithAccounts отримує власника з його особистими рахунками.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - ownerID: ID власника
//
// Повертає:
//   - *domain.OwnerWithAccounts: власник з рахунками
//   - error: помилки доступу
func (s *OwnerService) GetOwnerWithAccounts(userID, ownerID int) (*domain.OwnerWithAccounts, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.ownerRepo.GetWithAccounts(ownerID)
}

// GetOwnerSummary отримує зведену інформацію про власника.
// Включає кількість квартир, загальний баланс, борги.
//
// Параметри:
//   - userID: ID користувача
//   - ownerID: ID власника
//
// Повертає:
//   - string: форматована зведена інформація
//   - error: помилки доступу або розрахунку
func (s *OwnerService) GetOwnerSummary(userID, ownerID int) (string, error) {
	// Отримуємо власника з рахунками
	ownerData, err := s.GetOwnerWithAccounts(userID, ownerID)
	if err != nil {
		return "", err
	}

	// Підраховуємо загальний баланс
	var totalBalance float64
	var activeAccounts int
	var totalDebt float64
	var totalOverpayment float64

	for _, account := range ownerData.Accounts {
		if !account.IsActive {
			continue
		}
		activeAccounts++
		totalBalance += account.CurrentBalance
		if account.HasDebt() {
			totalDebt += account.GetDebtAmount()
		} else {
			totalOverpayment += account.GetOverpaymentAmount()
		}
	}

	summary := fmt.Sprintf(
		"Власник: %s\n"+
			"ІПН: %s\n"+
			"Контакт: %s\n"+
			"Кількість квартир: %d\n"+
			"Загальний баланс: %.2f грн\n",
		ownerData.Owner.FullName,
		ownerData.Owner.TaxID,
		ownerData.Owner.GetContactInfo(),
		activeAccounts,
		totalBalance,
	)

	if totalDebt > 0 {
		summary += fmt.Sprintf("⚠️ Заборгованість: %.2f грн\n", totalDebt)
	}
	if totalOverpayment > 0 {
		summary += fmt.Sprintf("✓ Переплата: %.2f грн\n", totalOverpayment)
	}

	return summary, nil
}

// ValidateOwnerData виконує розширену валідацію даних власника.
// Додаткові бізнес-правила понад базову валідацію.
//
// Параметри:
//   - owner: дані для валідації
//
// Повертає:
//   - error: nil якщо дані валідні
func (s *OwnerService) ValidateOwnerData(owner *domain.Owner) error {
	// Базова валідація з domain
	if err := owner.Validate(); err != nil {
		return err
	}

	// Бізнес-правило: ІПН має містити тільки цифри
	if !isNumeric(owner.TaxID) {
		return fmt.Errorf("ІПН має містити тільки цифри")
	}

	// Бізнес-правило: валідація формату телефону (українські номери)
	if !isValidPhoneNumber(owner.Phone) {
		return fmt.Errorf("невалідний формат телефону (очікується +380XXXXXXXXX)")
	}

	// Бізнес-правило: валідація email (якщо вказаний)
	if owner.Email != "" && !isValidEmail(owner.Email) {
		return fmt.Errorf("невалідний формат email")
	}

	// Бізнес-правило: альтернативний телефон (якщо вказаний)
	if owner.AlternativePhone != "" && !isValidPhoneNumber(owner.AlternativePhone) {
		return fmt.Errorf("невалідний формат альтернативного телефону")
	}

	// Бізнес-правило: перевірка паспортних даних (якщо вказані)
	if owner.PassportSeries != "" && len(owner.PassportSeries) != 2 {
		return fmt.Errorf("серія паспорта має містити 2 літери")
	}
	if owner.PassportNumber != "" && len(owner.PassportNumber) != 6 {
		return fmt.Errorf("номер паспорта має містити 6 цифр")
	}

	return nil
}

// CountOwners підраховує кількість активних власників.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - int: кількість власників
//   - error: помилки доступу
func (s *OwnerService) CountOwners(userID int) (int, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	return s.ownerRepo.Count()
}

// GetOwnersWithDebt отримує список власників з непогашеною заборгованістю.
// Доступно бухгалтерам та адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - map[int]float64: ключ - ID власника, значення - сума боргу
//   - error: помилки доступу
func (s *OwnerService) GetOwnersWithDebt(userID int) (map[int]float64, error) {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return nil, domain.ErrAccessDenied
	}

	// Отримуємо всіх власників
	owners, err := s.ownerRepo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make(map[int]float64)

	// Для кожного власника перевіряємо наявність боргу
	for _, owner := range owners {
		accounts, err := s.accountRepo.FindByOwnerID(owner.ID)
		if err != nil {
			continue
		}

		var totalDebt float64
		for _, account := range accounts {
			if account.IsActive && account.HasDebt() {
				totalDebt += account.GetDebtAmount()
			}
		}

		if totalDebt > 0 {
			result[owner.ID] = totalDebt
		}
	}

	return result, nil
}

// Допоміжні функції валідації

// isNumeric перевіряє, чи рядок містить тільки цифри.
func isNumeric(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}

// isValidPhoneNumber перевіряє валідність українського номера телефону.
func isValidPhoneNumber(phone string) bool {
	// Формат: +380XXXXXXXXX (12-13 символів)
	matched, _ := regexp.MatchString(`^\+380\d{9}$`, phone)
	return matched
}

// isValidEmail перевіряє валідність email адреси.
func isValidEmail(email string) bool {
	// Проста валідація email
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}
