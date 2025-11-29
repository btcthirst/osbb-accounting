// domain/repository/contractor_payment_repository.go
package repository

import (
	"context"

	"osbb-accounting/domain/entity"
)

// ContractorPaymentFilter визначає фільтри для пошуку платежів контрагентів.
type ContractorPaymentFilter struct {
	ContractorID   *int64
	PaymentMethod  *entity.PaymentMethod
	PeriodMonth    *int
	PeriodYear     *int
	StartDate      *int64 // UNIX timestamp
	EndDate        *int64 // UNIX timestamp
	MinAmount      *float64
	MaxAmount      *float64
	SearchQuery    string // Пошук по purpose, receipt_number
	OrderBy        string // payment_date, amount, period
	OrderDesc      bool
	Limit          int
	Offset         int
	IncludeDeleted bool
}

// ContractorPaymentRepository визначає інтерфейс для роботи з платежами контрагентів.
type ContractorPaymentRepository interface {
	// CRUD операції
	Create(ctx context.Context, payment *entity.ContractorPayment) error
	GetByID(ctx context.Context, id int64) (*entity.ContractorPayment, error)
	Update(ctx context.Context, payment *entity.ContractorPayment) error
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error

	// Пошук та фільтрація
	List(ctx context.Context, filter ContractorPaymentFilter) ([]*entity.ContractorPayment, error)
	Count(ctx context.Context, filter ContractorPaymentFilter) (int64, error)

	// Спеціалізовані методи
	GetByContractorID(ctx context.Context, contractorID int64) ([]*entity.ContractorPayment, error)
	GetTotalByPeriod(ctx context.Context, month, year int) (float64, error)
}
