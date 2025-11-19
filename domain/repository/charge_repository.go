// domain/repository/charge_repository.go
package repository

import (
	"context"
	"osbb-accounting/domain/entity"
	"time"
)

// ChargeRepository визначає контракт для роботи з нарахуваннями.
type ChargeRepository interface {
	// Create створює нове нарахування
	Create(ctx context.Context, charge *entity.Charge) error
	// GetByID отримує нарахування за ID
	GetByID(ctx context.Context, id int64) (*entity.Charge, error)

	// Update оновлює дані нарахування
	Update(ctx context.Context, charge *entity.Charge) error

	// SoftDelete м'яке видалення нарахування
	SoftDelete(ctx context.Context, id int64) error

	// Restore відновлює видалене нарахування
	Restore(ctx context.Context, id int64) error

	// List отримує список нарахувань з фільтрацією
	List(ctx context.Context, filter ChargeFilter) ([]*entity.Charge, error)

	// Count отримує загальну кількість нарахувань
	Count(ctx context.Context, filter ChargeFilter) (int64, error)

	// GetByOwnershipShareID отримує всі нарахування для частки власності
	GetByOwnershipShareID(ctx context.Context, ownershipShareID int64) ([]*entity.Charge, error)

	// GetByPeriod отримує нарахування за період
	GetByPeriod(ctx context.Context, month, year int) ([]*entity.Charge, error)

	// GetByOwnershipShareAndPeriod отримує нарахування для частки за період
	GetByOwnershipShareAndPeriod(ctx context.Context, ownershipShareID int64, month, year int) ([]*entity.Charge, error)

	// CheckDuplicatePeriod перевіряє наявність нарахування за період
	CheckDuplicatePeriod(ctx context.Context, ownershipShareID int64, chargeType entity.ChargeType, month, year int, excludeID *int64) (bool, error)

	// CalculateTotalForPeriod розраховує загальну суму нарахувань за період
	CalculateTotalForPeriod(ctx context.Context, month, year int) (float64, error)

	// CalculateTotalForOwnershipShare розраховує загальну суму нарахувань для частки
	CalculateTotalForOwnershipShare(ctx context.Context, ownershipShareID int64) (float64, error)

	// GetWithDetails отримує нарахування з повною інформацією
	GetWithDetails(ctx context.Context, id int64) (*ChargeDetails, error)

	// ListWithDetails отримує список нарахувань з повною інформацією
	ListWithDetails(ctx context.Context, filter ChargeFilter) ([]*ChargeDetails, error)

	// GetOverdueCharges отримує прострочені нарахування
	GetOverdueCharges(ctx context.Context, asOfDate time.Time) ([]*entity.Charge, error)

	// GetStatistics отримує статистику по нарахуваннях
	GetStatistics(ctx context.Context, filter ChargeStatisticsFilter) (*ChargeStatistics, error)

	// GetByApartmentID отримує всі нарахування для квартири
	GetByApartmentID(ctx context.Context, apartmentID int64) ([]*entity.Charge, error)

	// GetByOwnerID отримує всі нарахування для власника
	GetByOwnerID(ctx context.Context, ownerID int64) ([]*entity.Charge, error)

	// BulkCreate створює багато нарахувань за одну транзакцію
	BulkCreate(ctx context.Context, charges []*entity.Charge) error

	// GetPeriodRange отримує нарахування за діапазон періодів
	GetPeriodRange(ctx context.Context, startMonth, startYear, endMonth, endYear int) ([]*entity.Charge, error)
}

// ChargeFilter - критерії пошуку нарахувань
type ChargeFilter struct {
	IncludeDeleted   bool
	OwnershipShareID *int64
	ApartmentID      *int64 // Фільтр по квартирі (через ownership_shares)
	OwnerID          *int64 // Фільтр по власнику (через ownership_shares)
	ChargeType       *entity.ChargeType
	PeriodMonth      *int
	PeriodYear       *int
	StartDate        *time.Time
	EndDate          *time.Time
	MinAmount        *float64
	MaxAmount        *float64
	IsOverdue        *bool
	SearchQuery      string // Пошук по description
	Limit            int
	Offset           int
	OrderBy          string // "charge_date", "period", "amount", "created_at"
	OrderDesc        bool
}

// ChargeDetails - нарахування з повною інформацією
type ChargeDetails struct {
	Charge            *entity.Charge
	OwnerName         string
	OwnerPhone        *string
	OwnerEmail        *string
	ApartmentNumber   string
	ApartmentFloor    int
	ApartmentEntrance *int
	ShareFraction     string  // Частка власності (напр. "1/2")
	SharePercent      float64 // Відсоток власності
}

// ChargeStatisticsFilter - фільтр для статистики
type ChargeStatisticsFilter struct {
	PeriodMonth *int
	PeriodYear  *int
	StartDate   *time.Time
	EndDate     *time.Time
	ChargeType  *entity.ChargeType
	ApartmentID *int64
	OwnerID     *int64
}

// ChargeStatistics - статистика по нарахуваннях
type ChargeStatistics struct {
	TotalCharges       int64
	TotalAmount        float64
	AverageAmount      float64
	MinAmount          float64
	MaxAmount          float64
	ByType             map[entity.ChargeType]TypeStatistics
	ByPeriod           map[string]PeriodStatistics // Key: "MM/YYYY"
	OverdueCharges     int64
	OverdueAmount      float64
	ChargesWithTariff  int64
	ChargesWithPenalty int64
}

// TypeStatistics - статистика по типу нарахувань
type TypeStatistics struct {
	Count       int64
	TotalAmount float64
	AvgAmount   float64
}

// PeriodStatistics - статистика по періоду
type PeriodStatistics struct {
	Period      string // "MM/YYYY"
	Count       int64
	TotalAmount float64
	AvgAmount   float64
}
