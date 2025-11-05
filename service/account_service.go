package service

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"time"
)

// AccountService надає бізнес-логіку для управління особистими рахунками.
// Відокремлює бізнес-правила від UI та забезпечує перевірку прав доступу.
type AccountService struct {
	accountRepo   repository.PersonalAccountRepository
	apartmentRepo repository.ApartmentRepository
	ownerRepo     repository.OwnerRepository
	paymentRepo   repository.PaymentRepository
	userRepo      repository.UserRepository
}

// NewAccountService створює новий екземпляр сервісу особистих рахунків.
//
// Параметри:
//   - accountRepo: репозиторій особистих рахунків
//   - apartmentRepo: репозиторій квартир (для валідації)
//   - ownerRepo: репозиторій власників (для валідації)
//   - paymentRepo: репозиторій платежів (для розрахунку балансу)
//   - userRepo: репозиторій користувачів (для перевірки прав)
//
// Повертає:
//   - *AccountService: новий екземпляр сервісу
func NewAccountService(
	accountRepo repository.PersonalAccountRepository,
	apartmentRepo repository.ApartmentRepository,
	ownerRepo repository.OwnerRepository,
	paymentRepo repository.PaymentRepository,
	userRepo repository.UserRepository,
) *AccountService {
	return &AccountService{
		accountRepo:   accountRepo,
		apartmentRepo: apartmentRepo,
		ownerRepo:     ownerRepo,
		paymentRepo:   paymentRepo,
		userRepo:      userRepo,
	}
}

// CreateAccount створює новий особистий рахунок.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача, який створює рахунок
//   - apartmentID: ID квартири
//   - ownerID: ID власника
//
// Повертає:
//   - *domain.PersonalAccount: створений рахунок
//   - error: nil при успіху
func (s *AccountService) CreateAccount(userID, apartmentID, ownerID int) (*domain.PersonalAccount, error) {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return nil, domain.ErrAccessDenied
	}

	// Бізнес-правило: перевірка існування квартири
	apartment, err := s.apartmentRepo.FindByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не знайдена: %w", err)
	}

	// Бізнес-правило: перевірка існування власника
	owner, err := s.ownerRepo.FindByID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("власника не знайдено: %w", err)
	}

	// Бізнес-правило: перевірка, чи квартира вже не має рахунку
	existingAccount, err := s.accountRepo.FindByApartmentID(apartmentID)
	if err == nil && existingAccount != nil {
		return nil, domain.ErrAccountAlreadyAssigned
	}

	// Генеруємо унікальний номер рахунку
	accountNumber, err := s.accountRepo.GenerateAccountNumber()
	if err != nil {
		return nil, fmt.Errorf("помилка генерації номера рахунку: %w", err)
	}

	// Створюємо рахунок
	account := &domain.PersonalAccount{
		AccountNumber:  accountNumber,
		ApartmentID:    apartmentID,
		OwnerID:        ownerID,
		OpenedAt:       time.Now(),
		CurrentBalance: 0, // Початковий баланс = 0
		IsActive:       true,
		Notes:          fmt.Sprintf("Рахунок відкрито для кв. %s (власник: %s)", apartment.ApartmentNumber, owner.FullName),
	}

	// Збереження
	if err := s.accountRepo.Save(account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountByID отримує рахунок за ID.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - accountID: ID рахунку
//
// Повертає:
//   - *domain.PersonalAccount: знайдений рахунок
//   - error: помилки доступу або пошуку
func (s *AccountService) GetAccountByID(userID, accountID int) (*domain.PersonalAccount, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.accountRepo.FindByID(accountID)
}

// GetAccountByNumber отримує рахунок за номером.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - accountNumber: номер особистого рахунку
//
// Повертає:
//   - *domain.PersonalAccount: знайдений рахунок
//   - error: помилки доступу або пошуку
func (s *AccountService) GetAccountByNumber(userID int, accountNumber string) (*domain.PersonalAccount, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.accountRepo.FindByAccountNumber(accountNumber)
}

// GetAccountByApartment отримує рахунок квартири.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//
// Повертає:
//   - *domain.PersonalAccount: рахунок квартири
//   - error: помилки доступу або пошуку
func (s *AccountService) GetAccountByApartment(userID, apartmentID int) (*domain.PersonalAccount, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.accountRepo.FindByApartmentID(apartmentID)
}

// GetAllAccounts повертає список всіх активних рахунків.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - []*domain.PersonalAccount: список рахунків
//   - error: помилки доступу
func (s *AccountService) GetAllAccounts(userID int) ([]*domain.PersonalAccount, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.accountRepo.FindAll()
}

// GetDebtorAccounts отримує список рахунків з заборгованістю.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - []*domain.PersonalAccount: рахунки з боргом
//   - error: помилки доступу
func (s *AccountService) GetDebtorAccounts(userID int) ([]*domain.PersonalAccount, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.accountRepo.FindWithDebt()
}

// UpdateAccount оновлює дані рахунку.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - account: оновлені дані рахунку
//
// Повертає:
//   - error: nil при успіху
func (s *AccountService) UpdateAccount(userID int, account *domain.PersonalAccount) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Валідація
	if err := account.Validate(); err != nil {
		return err
	}

	return s.accountRepo.Update(account)
}

// RecalculateBalance пересраховує баланс рахунку на основі платежів.
// Синхронізує баланс з реальними даними з таблиці payments.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - accountID: ID рахунку для перерахунку
//
// Повертає:
//   - float64: новий баланс
//   - error: nil при успіху
func (s *AccountService) RecalculateBalance(userID, accountID int) (float64, error) {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return 0, domain.ErrAccessDenied
	}

	// Отримуємо рахунок
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return 0, err
	}

	// Розраховуємо баланс через PaymentRepository
	newBalance, err := s.paymentRepo.CalculateBalance(account.ApartmentID)
	if err != nil {
		return 0, fmt.Errorf("помилка розрахунку балансу: %w", err)
	}

	// Оновлюємо баланс рахунку
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		return 0, err
	}

	return newBalance, nil
}

// CloseAccount закриває особистий рахунок.
// Можливо тільки якщо немає непогашеного боргу.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//   - accountID: ID рахунку для закриття
//
// Повертає:
//   - error: nil при успіху
func (s *AccountService) CloseAccount(userID, accountID int) error {
	// Перевірка прав доступу (тільки адміни)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Бізнес-правило: перерахунок балансу перед закриттям
	newBalance, err := s.RecalculateBalance(userID, accountID)
	if err != nil {
		return fmt.Errorf("помилка перерахунку балансу: %w", err)
	}

	// Бізнес-правило: неможливо закрити рахунок з боргом
	if newBalance < 0 {
		return fmt.Errorf("%w (борг: %.2f грн)", domain.ErrCannotCloseAccountWithDebt, -newBalance)
	}

	return s.accountRepo.Close(accountID)
}

// ReopenAccount знову відкриває закритий рахунок.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - accountID: ID рахунку для відкриття
//
// Повертає:
//   - error: nil при успіху
func (s *AccountService) ReopenAccount(userID, accountID int) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() && !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	return s.accountRepo.Reopen(accountID)
}

// GetTotalDebt отримує загальну суму боргу по всіх рахунках.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - float64: загальний борг
//   - error: помилки доступу
func (s *AccountService) GetTotalDebt(userID int) (float64, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	return s.accountRepo.GetTotalDebt()
}

// GetTotalOverpayment отримує загальну суму переплат.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - float64: загальна переплата
//   - error: помилки доступу
func (s *AccountService) GetTotalOverpayment(userID int) (float64, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	return s.accountRepo.GetTotalOverpayment()
}

// GetFinancialSummary отримує загальну фінансову статистику.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - string: форматована статистика
//   - error: помилки доступу або розрахунку
func (s *AccountService) GetFinancialSummary(userID int) (string, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	// Збираємо статистику
	totalAccounts, err := s.accountRepo.Count()
	if err != nil {
		return "", err
	}

	totalDebt, err := s.accountRepo.GetTotalDebt()
	if err != nil {
		return "", err
	}

	totalOverpayment, err := s.accountRepo.GetTotalOverpayment()
	if err != nil {
		return "", err
	}

	debtorAccounts, err := s.accountRepo.FindWithDebt()
	if err != nil {
		return "", err
	}

	netBalance := totalOverpayment - totalDebt

	summary := fmt.Sprintf(
		"=== Фінансова Статистика ===\n"+
			"Всього рахунків: %d\n"+
			"Рахунків з боргом: %d\n"+
			"Загальний борг: %.2f грн\n"+
			"Загальна переплата: %.2f грн\n"+
			"Чистий баланс: %.2f грн\n",
		totalAccounts,
		len(debtorAccounts),
		totalDebt,
		totalOverpayment,
		netBalance,
	)

	if netBalance < 0 {
		summary += fmt.Sprintf("\n⚠️ Увага: загальна заборгованість перевищує переплати на %.2f грн\n", -netBalance)
	}

	return summary, nil
}

// GetAccountStatement формує виписку по рахунку.
// Включає баланс, історію операцій, статус.
//
// Параметри:
//   - userID: ID користувача
//   - accountID: ID рахунку
//
// Повертає:
//   - string: форматована виписка
//   - error: помилки доступу
func (s *AccountService) GetAccountStatement(userID, accountID int) (string, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	// Отримуємо рахунок
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return "", err
	}

	// Отримуємо квартиру
	apartment, err := s.apartmentRepo.FindByID(account.ApartmentID)
	if err != nil {
		return "", err
	}

	// Отримуємо власника
	owner, err := s.ownerRepo.FindByID(account.OwnerID)
	if err != nil {
		return "", err
	}

	// Формуємо виписку
	statement := fmt.Sprintf(
		"=== Виписка по Особистому Рахунку ===\n"+
			"Номер рахунку: %s\n"+
			"Квартира: №%s (поверх %d)\n"+
			"Власник: %s\n"+
			"ІПН: %s\n"+
			"Контакт: %s\n"+
			"Дата відкриття: %s\n"+
			"Статус: %s\n\n"+
			"Поточний баланс: %.2f грн\n",
		account.AccountNumber,
		apartment.ApartmentNumber,
		apartment.Floor,
		owner.FullName,
		owner.TaxID,
		owner.GetContactInfo(),
		account.OpenedAt.Format("02.01.2006"),
		getAccountStatus(account),
		account.CurrentBalance,
	)

	if account.HasDebt() {
		statement += fmt.Sprintf("⚠️ Заборгованість: %.2f грн\n", account.GetDebtAmount())
	} else if account.GetOverpaymentAmount() > 0 {
		statement += fmt.Sprintf("✓ Переплата: %.2f грн\n", account.GetOverpaymentAmount())
	}

	if account.Notes != "" {
		statement += fmt.Sprintf("\nПримітки: %s\n", account.Notes)
	}

	return statement, nil
}

// RecalculateAllBalances пересраховує баланси всіх рахунків.
// Корисно після масових операцій або виправлення помилок.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - int: кількість оновлених рахунків
//   - error: nil при успіху
func (s *AccountService) RecalculateAllBalances(userID int) (int, error) {
	// Перевірка прав доступу (тільки адміни)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, err
	}
	if !user.IsAdmin() {
		return 0, domain.ErrAccessDenied
	}

	// Отримуємо всі рахунки
	accounts, err := s.accountRepo.FindAll()
	if err != nil {
		return 0, err
	}

	successCount := 0
	for _, account := range accounts {
		_, err := s.RecalculateBalance(userID, account.ID)
		if err != nil {
			fmt.Printf("⚠️ Помилка перерахунку рахунку %s: %v\n", account.AccountNumber, err)
			continue
		}
		successCount++
	}

	return successCount, nil
}

func (s *AccountService) CountOwners(userID int) (int, error) {
	// Перевірка прав доступу (тільки адміни)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, err
	}
	if !user.IsAdmin() {
		return 0, domain.ErrAccessDenied
	}
	return s.accountRepo.Count()
}

// getAccountStatus повертає статус рахунку для відображення.
func getAccountStatus(account *domain.PersonalAccount) string {
	if account.IsClosed() {
		return fmt.Sprintf("Закритий (%s)", account.ClosedAt.Format("02.01.2006"))
	}
	if !account.IsActive {
		return "Неактивний"
	}
	return "Активний"
}
