package service

import (
	"osbb-accounting/domain"
	"osbb-accounting/repository"
)

// ApartmentService надає бізнес-логіку для управління квартирами.
// Відокремлює бізнес-правила від UI та забезпечує перевірку прав доступу.
type ApartmentService struct {
	apartmentRepo repository.ApartmentRepository
	userRepo      repository.UserRepository
}

// NewApartmentService створює новий екземпляр сервісу квартир.
//
// Параметри:
//   - apartmentRepo: репозиторій для роботи з квартирами
//   - userRepo: репозиторій користувачів (для перевірки прав)
//
// Повертає:
//   - *ApartmentService: новий екземпляр сервісу
func NewApartmentService(
	apartmentRepo repository.ApartmentRepository,
	userRepo repository.UserRepository,
) *ApartmentService {
	return &ApartmentService{
		apartmentRepo: apartmentRepo,
		userRepo:      userRepo,
	}
}

// CreateApartment створює нову квартиру.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача, який створює квартиру
//   - apartment: дані квартири для створення
//
// Повертає:
//   - error: nil при успіху, domain.ErrAccessDenied якщо немає прав
func (s *ApartmentService) CreateApartment(userID int, apartment *domain.Apartment) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Валідація даних
	if err := apartment.Validate(); err != nil {
		return err
	}

	// Бізнес-правило: номер квартири має бути унікальним
	existing, err := s.apartmentRepo.FindByNumber(apartment.ApartmentNumber)
	if err == nil && existing != nil {
		return domain.ErrApartmentAlreadyExists
	}

	// Збереження
	return s.apartmentRepo.Save(apartment)
}

// GetApartmentByID отримує квартиру за ID.
// Доступно всім авторизованим користувачам.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//
// Повертає:
//   - *domain.Apartment: знайдена квартира
//   - error: помилки доступу або пошуку
func (s *ApartmentService) GetApartmentByID(userID, apartmentID int) (*domain.Apartment, error) {
	// Перевірка, що користувач авторизований
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.FindByID(apartmentID)
}

// GetApartmentByNumber отримує квартиру за номером.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentNumber: номер квартири
//
// Повертає:
//   - *domain.Apartment: знайдена квартира
//   - error: помилки доступу або пошуку
func (s *ApartmentService) GetApartmentByNumber(userID int, apartmentNumber string) (*domain.Apartment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.FindByNumber(apartmentNumber)
}

// GetAllApartments повертає список всіх квартир.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - []*domain.Apartment: список квартир
//   - error: помилки доступу
func (s *ApartmentService) GetAllApartments(userID int) ([]*domain.Apartment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.FindAll()
}

// SearchApartments шукає квартири за фільтром.
//
// Параметри:
//   - userID: ID користувача
//   - filter: параметри пошуку
//
// Повертає:
//   - []*domain.Apartment: знайдені квартири
//   - error: помилки доступу
func (s *ApartmentService) SearchApartments(userID int, filter *domain.ApartmentFilter) ([]*domain.Apartment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.FindByFilter(filter)
}

// UpdateApartment оновлює дані квартири.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - apartment: оновлені дані квартири
//
// Повертає:
//   - error: nil при успіху
func (s *ApartmentService) UpdateApartment(userID int, apartment *domain.Apartment) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	// Валідація даних
	if err := apartment.Validate(); err != nil {
		return err
	}

	// Бізнес-правило: перевірка, чи квартира існує
	existing, err := s.apartmentRepo.FindByID(apartment.ID)
	if err != nil {
		return err
	}

	// Бізнес-правило: якщо змінюється номер, перевіряємо унікальність
	if existing.ApartmentNumber != apartment.ApartmentNumber {
		duplicate, err := s.apartmentRepo.FindByNumber(apartment.ApartmentNumber)
		if err == nil && duplicate != nil && duplicate.ID != apartment.ID {
			return domain.ErrApartmentAlreadyExists
		}
	}

	return s.apartmentRepo.Update(apartment)
}

// DeleteApartment видаляє квартиру (м'яке видалення).
// Доступно тільки адміністраторам.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири для видалення
//
// Повертає:
//   - error: nil при успіху
func (s *ApartmentService) DeleteApartment(userID, apartmentID int) error {
	// Перевірка прав доступу (тільки адміни можуть видаляти)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin() {
		return domain.ErrAccessDenied
	}

	// Бізнес-правило: перевірка, чи квартира існує
	_, err = s.apartmentRepo.FindByID(apartmentID)
	if err != nil {
		return err
	}

	// TODO: Перевірити, чи немає непогашених боргів
	// Це буде додано після реалізації модуля платежів

	return s.apartmentRepo.Deactivate(apartmentID)
}

// ActivateApartment активує квартиру.
// Доступно адмінам та бухгалтерам.
//
// Параметри:
//   - userID: ID користувача
//   - apartmentID: ID квартири
//
// Повертає:
//   - error: nil при успіху
func (s *ApartmentService) ActivateApartment(userID, apartmentID int) error {
	// Перевірка прав доступу
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.CanManageFinances() {
		return domain.ErrAccessDenied
	}

	return s.apartmentRepo.Activate(apartmentID)
}

// GetStatistics отримує статистику по квартирах.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - *repository.ApartmentStatistics: статистика
//   - error: помилки доступу
func (s *ApartmentService) GetStatistics(userID int) (*repository.ApartmentStatistics, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.GetStatistics()
}

// GetApartmentsByFloor отримує квартири на вказаному поверсі.
//
// Параметри:
//   - userID: ID користувача
//   - floor: номер поверху
//
// Повертає:
//   - []*domain.Apartment: квартири на поверсі
//   - error: помилки доступу
func (s *ApartmentService) GetApartmentsByFloor(userID, floor int) ([]*domain.Apartment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.FindByFloor(floor)
}

// GetApartmentsByEntrance отримує квартири у вказаному під'їзді.
//
// Параметри:
//   - userID: ID користувача
//   - entrance: номер під'їзду
//
// Повертає:
//   - []*domain.Apartment: квартири у під'їзді
//   - error: помилки доступу
func (s *ApartmentService) GetApartmentsByEntrance(userID, entrance int) ([]*domain.Apartment, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.apartmentRepo.FindByEntrance(entrance)
}

// ValidateApartmentData перевіряє дані квартири перед збереженням.
// Додаткові бізнес-правила можуть бути додані тут.
//
// Параметри:
//   - apartment: дані для валідації
//
// Повертає:
//   - error: nil якщо дані валідні
func (s *ApartmentService) ValidateApartmentData(apartment *domain.Apartment) error {
	// Базова валідація з domain
	if err := apartment.Validate(); err != nil {
		return err
	}

	// Додаткові бізнес-правила

	// Правило 1: Площа має бути розумною (10-500 м²)
	if apartment.Area < 10 || apartment.Area > 500 {
		return domain.ErrInvalidArea
	}

	// Правило 2: Кількість кімнат відповідає площі
	if apartment.Rooms > 0 {
		minAreaPerRoom := apartment.Area / float64(apartment.Rooms)
		if minAreaPerRoom < 8 { // менше 8м² на кімнату - підозріло
			return domain.ErrInvalidRooms
		}
	}

	// Правило 3: Поверх не може бути надто великим
	if apartment.Floor > 50 {
		return domain.ErrInvalidFloor
	}

	// Правило 4: Кількість мешканців має бути розумною
	if apartment.ResidentsCount > 20 {
		return domain.ErrInvalidResidentsCount
	}

	return nil
}

// CalculateTotalArea розраховує загальну площу всіх квартир.
// Корисно для звітів.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - float64: загальна площа в м²
//   - error: помилки доступу
func (s *ApartmentService) CalculateTotalArea(userID int) (float64, error) {
	stats, err := s.GetStatistics(userID)
	if err != nil {
		return 0, err
	}
	return stats.TotalArea, nil
}

// CountApartments підраховує кількість квартир.
//
// Параметри:
//   - userID: ID користувача
//
// Повертає:
//   - int: кількість квартир
//   - error: помилки доступу
func (s *ApartmentService) CountApartments(userID int) (int, error) {
	// Перевірка авторизації
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}

	return s.apartmentRepo.Count()
}
