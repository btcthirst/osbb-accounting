package service

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"time"
)

// PaymentService надає бізнес-логіку для управління платежами.
// Відокремлює бізнес-правила від UI та забезпечує перевірку прав доступу.
type PaymentService struct {
	paymentRepo   repository.PaymentRepository
	apartmentRepo repository.ApartmentRepository
	userRepo      repository.UserRepository
}

// NewPaymentService створює новий екземпляр сервісу платежів.
//
// Параметри:
//   - paymentRepo: репозиторій платежів
//   - apartmentRepo: репозиторій квартир (для валідації)
//   - userRepo: репозиторій користувачів (для перевірки прав)
//
// Повертає:
//   - *PaymentService: новий екземпляр сервісу
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	apartmentRepo repository.ApartmentRepository,
	userRepo repository.UserRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo:   paymentRepo,
		apartmentRepo: apartmentRepo,
		userRepo:      userRepo,
	}
}

// CreatePayment створює новий платіж (вхідний).
// Доступно бухгалтерам та адміністраторам.
//
// Параметри:
//   - userID: ID користувача, який створює платіж
//   - apartmentID: ID квартири
//   - amount: сума платежу
//   - category: категорія платежу
//   - description: опис
//   - paymentDate: дата платежу
//   - period: період (YYYY-MM)
//   - notes: примітки (опціонально)
//
// Повертає:
//   - *domain.Payment: створений платіж
//   - error: помилки доступу або валідації
func (s *PaymentService) CreatePayment(
	userID, apartmentID int,
	amount float64,
	category domain.PaymentCategory,
	description string,
	paymentDate time.Time,
	period string,
	notes string,
) (*domain.Payment, error) {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.CanManageFinances() {
		return nil, domain.ErrAccessDenied
	}

	// Бізнес-правило: перевірка існування квартири
	_, err = s.apartmentRepo.FindByID(apartmentID)
	if err != nil {
		return nil, err
	}

	// Бізнес-правило: сума має бути додатною
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Створення платежу
	payment := &domain.Payment{
		ApartmentID: apartmentID,
		Type:        domain.PaymentTypeIncoming,
		Category:    category,
		Amount:      amount,
		Description: description,
		PaymentDate: paymentDate,
		Period:      period,
		CreatedBy:   userID,
		Notes:       notes,
	}

	// Валідація
	if err := payment.Validate(); err != nil {
		return nil, err
	}

	// Збереження
	if err := s.paymentRepo.Save(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// CreateCharge створює нарахування для квартири.
// Доступно бухгалтерам та адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//   - amount: сума нарахування
//   - category: категорія
//   - description: опис
//   - period: період (YYYY-MM)
//   - notes: примітки
//
// Повертає:
//   - *domain.Payment: створене нарахування
//   - error: помилки
func (s *PaymentService) CreateCharge(
	userID, apartmentID int,
	amount float64,
	category domain.PaymentCategory,
	description string,
	period string,
	notes string,
) (*domain.Payment, error) {
	// Перевірка прав
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.CanManageFinances() {
		return nil, domain.ErrAccessDenied
	}

	// Перевірка квартири
	_, err = s.apartmentRepo.FindByID(apartmentID)
	if err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	charge := &domain.Payment{
		ApartmentID: apartmentID,
		Type:        domain.PaymentTypeCharge,
		Category:    category,
		Amount:      amount,
		Description: description,
		PaymentDate: time.Now(),
		Period:      period,
		CreatedBy:   userID,
		Notes:       notes,
	}

	if err := charge.Validate(); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Save(charge); err != nil {
		return nil, err
	}

	return charge, nil
}

// CreateMonthlyCharges створює стандартні нарахування на всі квартири за місяць.
// Доступно тільки адміністраторам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - period: період (YYYY-MM)
//   - maintenanceRate: тариф на утримання за м²
//   - utilitiesRate: тариф на комунальні за м²
//
// Повертає:
//   - int: кількість створених нарахувань
//   - error: помилки
func (s *PaymentService) CreateMonthlyCharges(
	userID int,
	period string,
	maintenanceRate, utilitiesRate float64,
) (int, error) {
	// Перевірка прав
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, err
	}
	if !user.CanManageFinances() {
		return 0, domain.ErrAccessDenied
	}

	// Отримуємо всі активні квартири
	apartments, err := s.apartmentRepo.FindAll()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, apt := range apartments {
		// Нарахування на утримання
		maintenanceAmount := apt.Area * maintenanceRate
		maintenance := &domain.Payment{
			ApartmentID: apt.ID,
			Type:        domain.PaymentTypeCharge,
			Category:    domain.CategoryMaintenance,
			Amount:      maintenanceAmount,
			Description: fmt.Sprintf("Утримання будинку за %s", period),
			PaymentDate: time.Now(),
			Period:      period,
			CreatedBy:   userID,
		}

		if err := s.paymentRepo.Save(maintenance); err != nil {
			continue // Пропускаємо помилки і продовжуємо
		}
		count++

		// Нарахування на комунальні
		utilitiesAmount := apt.Area * utilitiesRate
		utilities := &domain.Payment{
			ApartmentID: apt.ID,
			Type:        domain.PaymentTypeCharge,
			Category:    domain.CategoryUtilities,
			Amount:      utilitiesAmount,
			Description: fmt.Sprintf("Комунальні послуги за %s", period),
			PaymentDate: time.Now(),
			Period:      period,
			CreatedBy:   userID,
		}

		if err := s.paymentRepo.Save(utilities); err != nil {
			continue
		}
		count++
	}

	return count, nil
}

// GetPaymentByID отримує платіж за ID.
//
// Параметри:
//   - userID: ID користувача
//   - paymentID: ID платежу
//
// Повертає:
//   - *domain.Payment: знайдений платіж
//   - error: помилки
func (s *PaymentService) GetPaymentByID(userID, paymentID int) (*domain.Payment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.FindByID(paymentID)
}

// GetPaymentsByApartment отримує всі платежі квартири.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//
// Повертає:
//   - []*domain.Payment: список платежів
//   - error: помилки
func (s *PaymentService) GetPaymentsByApartment(userID, apartmentID int) ([]*domain.Payment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.FindByApartmentID(apartmentID)
}

// SearchPayments шукає платежі за фільтром.
//
// Параметри:
//   - userID: ID користувача
//   - filter: параметри пошуку
//
// Повертає:
//   - []*domain.Payment: знайдені платежі
//   - error: помилки
func (s *PaymentService) SearchPayments(userID int, filter *domain.PaymentFilter) ([]*domain.Payment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.FindByFilter(filter)
}

// UpdatePayment оновлює платіж.
// Доступно бухгалтерам та адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//   - payment: оновлені дані платежу
//
// Повертає:
//   - error: nil при успіху
func (s *PaymentService) UpdatePayment(userID int, payment *domain.Payment) error {
	// Перевірка прав
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Валідація
	if err := payment.Validate(); err != nil {
		return err
	}

	return s.paymentRepo.Update(payment)
}

// DeletePayment видаляє платіж.
// Доступно тільки адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//   - paymentID: ID платежу
//
// Повертає:
//   - error: nil при успіху
func (s *PaymentService) DeletePayment(userID, paymentID int) error {
	// Перевірка прав (тільки адміни можуть видаляти)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() {
		return domain.ErrAccessDenied
	}

	return s.paymentRepo.Delete(paymentID)
}

// GetApartmentBalance отримує поточний баланс квартири.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//
// Повертає:
//   - float64: баланс (+ переплата, - борг)
//   - error: помилки
func (s *PaymentService) GetApartmentBalance(userID, apartmentID int) (float64, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	return s.paymentRepo.CalculateBalance(apartmentID)
}

// GetApartmentSummary отримує зведення по квартирі.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//
// Повертає:
//   - *domain.PaymentSummary: зведення
//   - error: помилки
func (s *PaymentService) GetApartmentSummary(userID, apartmentID int) (*domain.PaymentSummary, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.GetSummaryByApartment(apartmentID)
}

// GetMonthlyReport генерує звіт за місяць для квартири.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//   - period: період (YYYY-MM)
//
// Повертає:
//   - *domain.MonthlyReport: звіт
//   - error: помилки
func (s *PaymentService) GetMonthlyReport(userID, apartmentID int, period string) (*domain.MonthlyReport, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.GetMonthlyReport(apartmentID, period)
}

// GetDebtorsList отримує список квартир з боргом.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - []int: ID квартир з боргом
//   - error: помилки
func (s *PaymentService) GetDebtorsList(userID int) ([]int, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.GetDebtorApartments()
}

// GetDebtorsDetails отримує детальну інформацію про боржників.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - map[int]*domain.PaymentSummary: зведення по боржникам (ключ - apartment_id)
//   - error: помилки
func (s *PaymentService) GetDebtorsDetails(userID int) (map[int]*domain.PaymentSummary, error) {
	// Отримуємо список боржників
	debtorIDs, err := s.GetDebtorsList(userID)
	if err != nil {
		return nil, err
	}

	result := make(map[int]*domain.PaymentSummary)

	for _, apartmentID := range debtorIDs {
		summary, err := s.paymentRepo.GetSummaryByApartment(apartmentID)
		if err != nil {
			continue
		}
		result[apartmentID] = summary
	}

	return result, nil
}

// GetPaymentsByCategory отримує розбивку платежів по категоріях.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири (0 = всі квартири)
//   - period: період (порожній = всі періоди)
//
// Повертає:
//   - map[domain.PaymentCategory]float64: суми по категоріях
//   - error: помилки
func (s *PaymentService) GetPaymentsByCategory(userID, apartmentID int, period string) (map[domain.PaymentCategory]float64, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.paymentRepo.GetPaymentsByCategory(apartmentID, period)
}

// GetRecentPayments отримує останні платежі.
//
// Параметри:
//   - userID: ID користувача
//   - limit: кількість записів
//
// Повертає:
//   - []*domain.Payment: останні платежі
//   - error: помилки
func (s *PaymentService) GetRecentPayments(userID, limit int) ([]*domain.Payment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.paymentRepo.GetRecentPayments(limit)
}

// ValidatePaymentData перевіряє дані платежу.
//
// Параметри:
//   - payment: дані для валідації
//
// Повертає:
//   - error: nil якщо дані валідні
func (s *PaymentService) ValidatePaymentData(payment *domain.Payment) error {
	// Базова валідація
	if err := payment.Validate(); err != nil {
		return err
	}

	// Бізнес-правила: сума не може бути занадто великою
	if payment.Amount > 1000000 {
		return fmt.Errorf("сума платежу занадто велика (макс. 1,000,000 грн)")
	}

	// Бізнес-правило: дата платежу не може бути в майбутньому
	if payment.PaymentDate.After(time.Now().AddDate(0, 0, 1)) {
		return fmt.Errorf("дата платежу не може бути в майбутньому")
	}

	// Бізнес-правило: дата не може бути занадто старою (більше 5 років)
	fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
	if payment.PaymentDate.Before(fiveYearsAgo) {
		return fmt.Errorf("дата платежу занадто стара (більше 5 років)")
	}

	return nil
}

// GetCurrentPeriod повертає поточний період у форматі YYYY-MM.
func (s *PaymentService) GetCurrentPeriod() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}
