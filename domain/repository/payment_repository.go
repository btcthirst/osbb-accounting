// domain/repository/payment_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// PaymentRepository визначає інтерфейс для роботи з платежами.
// Цей інтерфейс має бути реалізований в infrastructure шарі.
type PaymentRepository interface {
	// Create створює новий запис платежу
	Create(ctx context.Context, payment *entity.Payment) error

	// GetByID отримує платіж за унікальним ідентифікатором
	GetByID(ctx context.Context, id int64) (*entity.Payment, error)

	// Update оновлює дані існуючого платежу
	Update(ctx context.Context, payment *entity.Payment) error

	// SoftDelete виконує м'яке видалення платежу
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видалений платіж
	Restore(ctx context.Context, id int64) error

	// List отримує список платежів відповідно до фільтру
	List(ctx context.Context, filter PaymentFilter) ([]*entity.Payment, error)

	// Count рахує кількість платежів, що відповідають фільтру (для пагінації)
	Count(ctx context.Context, filter PaymentFilter) (int64, error)

	// GetByOwnershipShareID отримує всі платежі для конкретної частки власності
	GetByOwnershipShareID(ctx context.Context, shareID int64) ([]*entity.Payment, error)

	// GetStatistics отримує агреговану статистику по платежах
	GetStatistics(ctx context.Context, filter PaymentFilter) (*PaymentStatistics, error)

	// GetTotalByPeriod повертає суму платежів за певний місяць і рік
	// Може використовуватися для звірки з нарахуваннями
	GetTotalByPeriod(ctx context.Context, month, year int) (float64, error)
}

// PaymentFilter визначає критерії пошуку платежів.
type PaymentFilter struct {
	IncludeDeleted   bool
	SearchQuery      string // Пошук по призначенню, номеру квитанції
	OwnershipShareID *int64
	OwnerID          *int64 // Для пошуку всіх платежів власника (по всіх його квартирах)
	ApartmentID      *int64 // Для пошуку всіх платежів по квартирі

	// Фільтри дати оплати
	StartDate *int64 // UNIX timestamp
	EndDate   *int64 // UNIX timestamp

	// Фільтри періоду (за що платили)
	PeriodMonth *int
	PeriodYear  *int

	PaymentMethod *entity.PaymentMethod
	IsApproved    *bool
	ApprovedBy    *int64

	MinAmount *float64
	MaxAmount *float64

	Limit  int
	Offset int

	OrderBy   string // "payment_date", "amount", "created_at", "period"
	OrderDesc bool
}

// PaymentStatistics містить зведену інформацію про платежі.
type PaymentStatistics struct {
	TotalCount    int64
	TotalAmount   float64
	AverageAmount float64
	MinAmount     float64
	MaxAmount     float64

	// Розбивка по методах оплати (кількість)
	CountByCash         int64
	CountByCard         int64
	CountByBankTransfer int64
	CountByOther        int64

	// Розбивка по методах оплати (сума)
	AmountByCash         float64
	AmountByCard         float64
	AmountByBankTransfer float64
	AmountByOther        float64

	ApprovedCount    int64
	NotApprovedCount int64
}
