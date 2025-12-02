// application/usecase/apartment/list_apartments.go
package apartment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ListApartmentsUseCase отримує список квартир.
type ListApartmentsUseCase struct {
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

// NewListApartmentsUseCase створює новий use case.
func NewListApartmentsUseCase(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *ListApartmentsUseCase {
	return &ListApartmentsUseCase{
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

// ListApartmentsInput - вхідні дані.
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

// ListApartmentsOutput - результат.
type ListApartmentsOutput struct {
	Apartments []*ApartmentOutput
	Total      int64
	Limit      int
	Offset     int
	HasMore    bool
}

// Execute виконує отримання списку квартир.
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
