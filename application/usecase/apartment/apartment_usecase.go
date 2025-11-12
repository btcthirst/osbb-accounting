// application/usecase/apartment/apartment_usecases.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ApartmentOutput - результат операцій з квартирою.
type ApartmentOutput struct {
	ID              int64
	ApartmentNumber string
	Floor           int
	Entrance        *int
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
	IsActive        bool
	DisplayName     string
	FullInfo        string
	AreaInfo        string
}

// ============================================================================
// CreateApartmentUseCase
// ============================================================================

type CreateApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewCreateApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *CreateApartmentUseCase {
	return &CreateApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type CreateApartmentInput struct {
	CurrentUserID   int64
	ApartmentNumber string
	Floor           int
	Entrance        *int
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
}

func (uc *CreateApartmentUseCase) Execute(ctx context.Context, input CreateApartmentInput) (*ApartmentOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Перевірка унікальності номеру
	exists, err := uc.apartmentRepo.ExistsByNumber(ctx, input.ApartmentNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check apartment number: %w", err)
	}
	if exists {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			"apartment with this number already exists",
			domainErrors.ErrAlreadyExists,
		).WithDetails("field", "apartment_number")
	}

	// Створення entity
	apartment, err := entity.NewApartment(
		input.ApartmentNumber,
		input.Floor,
		input.AreaTotal,
		input.Entrance,
		input.AreaLiving,
		input.RoomsCount,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid apartment data",
			err,
		)
	}

	apartment.CadastralNumber = input.CadastralNumber
	apartment.Notes = input.Notes

	// Збереження
	if err := uc.apartmentRepo.Create(ctx, apartment); err != nil {
		return nil, fmt.Errorf("failed to create apartment: %w", err)
	}

	return mapApartmentToOutput(apartment), nil
}

// ============================================================================
// GetApartmentUseCase
// ============================================================================

type GetApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewGetApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *GetApartmentUseCase {
	return &GetApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type GetApartmentInput struct {
	CurrentUserID int64
	ApartmentID   int64
}

func (uc *GetApartmentUseCase) Execute(ctx context.Context, input GetApartmentInput) (*ApartmentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	apartment, err := uc.apartmentRepo.GetByID(ctx, input.ApartmentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"apartment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}

	return mapApartmentToOutput(apartment), nil
}

// ============================================================================
// UpdateApartmentUseCase
// ============================================================================

type UpdateApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewUpdateApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateApartmentUseCase {
	return &UpdateApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type UpdateApartmentInput struct {
	CurrentUserID   int64
	ApartmentID     int64
	Floor           int
	Entrance        *int
	AreaTotal       float64
	AreaLiving      *float64
	RoomsCount      *int
	CadastralNumber *string
	Notes           *string
}

func (uc *UpdateApartmentUseCase) Execute(ctx context.Context, input UpdateApartmentInput) (*ApartmentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	apartment, err := uc.apartmentRepo.GetByID(ctx, input.ApartmentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"apartment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}

	if err := apartment.Update(
		input.Floor,
		input.AreaTotal,
		input.Entrance,
		input.AreaLiving,
		input.RoomsCount,
		input.CadastralNumber,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid apartment data",
			err,
		)
	}

	if err := uc.apartmentRepo.Update(ctx, apartment); err != nil {
		return nil, fmt.Errorf("failed to update apartment: %w", err)
	}

	return mapApartmentToOutput(apartment), nil
}

// ============================================================================
// DeleteApartmentUseCase
// ============================================================================

type DeleteApartmentUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewDeleteApartmentUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteApartmentUseCase {
	return &DeleteApartmentUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type DeleteApartmentInput struct {
	CurrentUserID int64
	ApartmentID   int64
}

type DeleteApartmentOutput struct {
	Success bool
	Message string
}

func (uc *DeleteApartmentUseCase) Execute(ctx context.Context, input DeleteApartmentInput) (*DeleteApartmentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	apartment, err := uc.apartmentRepo.GetByID(ctx, input.ApartmentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"apartment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}

	if err := uc.apartmentRepo.SoftDelete(ctx, input.ApartmentID); err != nil {
		return nil, fmt.Errorf("failed to delete apartment: %w", err)
	}

	return &DeleteApartmentOutput{
		Success: true,
		Message: fmt.Sprintf("Apartment '%s' successfully deleted", apartment.GetDisplayName()),
	}, nil
}

// ============================================================================
// ListApartmentsUseCase
// ============================================================================

type ListApartmentsUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewListApartmentsUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *ListApartmentsUseCase {
	return &ListApartmentsUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type ListApartmentsInput struct {
	CurrentUserID  int64
	IncludeDeleted bool
	IsActive       *bool
	SearchQuery    string
	Floor          *int
	Entrance       *int
	MinArea        *float64
	MaxArea        *float64
	HasOwners      *bool
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

type ListApartmentsOutput struct {
	Apartments []*ApartmentOutput
	Total      int64
	Limit      int
	Offset     int
	HasMore    bool
}

func (uc *ListApartmentsUseCase) Execute(ctx context.Context, input ListApartmentsInput) (*ListApartmentsOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.ApartmentFilter{
		IncludeDeleted: input.IncludeDeleted,
		IsActive:       input.IsActive,
		SearchQuery:    input.SearchQuery,
		Floor:          input.Floor,
		Entrance:       input.Entrance,
		MinArea:        input.MinArea,
		MaxArea:        input.MaxArea,
		HasOwners:      input.HasOwners,
		Limit:          input.Limit,
		Offset:         input.Offset,
		OrderBy:        input.OrderBy,
		OrderDesc:      input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	apartments, err := uc.apartmentRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list apartments: %w", err)
	}

	total, err := uc.apartmentRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count apartments: %w", err)
	}

	outputs := make([]*ApartmentOutput, len(apartments))
	for i, apt := range apartments {
		outputs[i] = mapApartmentToOutput(apt)
	}

	return &ListApartmentsOutput{
		Apartments: outputs,
		Total:      total,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		HasMore:    int64(filter.Offset+filter.Limit) < total,
	}, nil
}

// ============================================================================
// GetApartmentStatisticsUseCase
// ============================================================================

type GetApartmentStatisticsUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewGetApartmentStatisticsUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *GetApartmentStatisticsUseCase {
	return &GetApartmentStatisticsUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type GetApartmentStatisticsInput struct {
	CurrentUserID int64
}

type ApartmentStatisticsOutput struct {
	TotalApartments int64
	TotalArea       float64
	AverageArea     float64
	MinArea         float64
	MaxArea         float64
	FloorCount      int
	EntranceCount   int
	WithOwners      int64
	WithoutOwners   int64
}

func (uc *GetApartmentStatisticsUseCase) Execute(ctx context.Context, input GetApartmentStatisticsInput) (*ApartmentStatisticsOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "apartments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	stats, err := uc.apartmentRepo.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &ApartmentStatisticsOutput{
		TotalApartments: stats.TotalApartments,
		TotalArea:       stats.TotalArea,
		AverageArea:     stats.AverageArea,
		MinArea:         stats.MinArea,
		MaxArea:         stats.MaxArea,
		FloorCount:      stats.FloorCount,
		EntranceCount:   stats.EntranceCount,
		WithOwners:      stats.WithOwners,
		WithoutOwners:   stats.WithoutOwners,
	}, nil
}

// ============================================================================
// Helper
// ============================================================================

func mapApartmentToOutput(apt *entity.Apartment) *ApartmentOutput {
	return &ApartmentOutput{
		ID:              apt.ID,
		ApartmentNumber: apt.ApartmentNumber,
		Floor:           apt.Floor,
		Entrance:        apt.Entrance,
		AreaTotal:       apt.AreaTotal,
		AreaLiving:      apt.AreaLiving,
		RoomsCount:      apt.RoomsCount,
		CadastralNumber: apt.CadastralNumber,
		Notes:           apt.Notes,
		IsActive:        apt.IsActive,
		DisplayName:     apt.GetDisplayName(),
		FullInfo:        apt.GetFullInfo(),
		AreaInfo:        apt.GetAreaInfo(),
	}
}
