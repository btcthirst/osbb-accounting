package charge

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ChargeOutput - результат операцій з нарахуванням.
type ChargeOutput struct {
	ID               int64
	OwnershipShareID int64
	ChargeType       string
	ChargeDate       time.Time
	PeriodMonth      int
	PeriodYear       int
	Amount           float64
	Tariff           *float64
	Quantity         *float64
	Description      *string
	Notes            *string
	PeriodDisplay    string
	AmountDisplay    string
	TypeName         string
}

// ============================================================================
// CreateChargeUseCase
// ============================================================================

type CreateChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

func NewCreateChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *CreateChargeUseCase {
	return &CreateChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

type CreateChargeInput struct {
	CurrentUserID    int64
	OwnershipShareID int64
	ChargeType       entity.ChargeType
	ChargeDate       time.Time
	PeriodMonth      int
	PeriodYear       int
	Amount           float64
	Tariff           *float64
	Quantity         *float64
	Description      *string
	Notes            *string
}

func (uc *CreateChargeUseCase) Execute(ctx context.Context, input CreateChargeInput) (*ChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Check for duplicates
	isDuplicate, err := uc.chargeRepo.CheckDuplicatePeriod(
		ctx, input.OwnershipShareID, input.ChargeType, input.PeriodMonth, input.PeriodYear, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if isDuplicate {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			"charge for this period already exists",
			domainErrors.ErrAlreadyExists,
		)
	}

	charge, err := entity.NewCharge(
		input.OwnershipShareID,
		input.ChargeType,
		input.ChargeDate,
		input.PeriodMonth,
		input.PeriodYear,
		input.Amount,
		input.Tariff,
		input.Quantity,
		input.Description,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid charge data",
			err,
		)
	}
	charge.Notes = input.Notes

	if err := uc.chargeRepo.Create(ctx, charge); err != nil {
		return nil, fmt.Errorf("failed to create charge: %w", err)
	}

	return mapChargeToOutput(charge), nil
}

// ============================================================================
// GetChargeUseCase
// ============================================================================

type GetChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

func NewGetChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *GetChargeUseCase {
	return &GetChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

type GetChargeInput struct {
	CurrentUserID int64
	ChargeID      int64
}

func (uc *GetChargeUseCase) Execute(ctx context.Context, input GetChargeInput) (*ChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	charge, err := uc.chargeRepo.GetByID(ctx, input.ChargeID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"charge not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get charge: %w", err)
	}

	return mapChargeToOutput(charge), nil
}

// ============================================================================
// UpdateChargeUseCase
// ============================================================================

type UpdateChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

func NewUpdateChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateChargeUseCase {
	return &UpdateChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

type UpdateChargeInput struct {
	CurrentUserID int64
	ChargeID      int64
	Amount        float64
	Tariff        *float64
	Quantity      *float64
	Description   *string
	Notes         *string
}

func (uc *UpdateChargeUseCase) Execute(ctx context.Context, input UpdateChargeInput) (*ChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	charge, err := uc.chargeRepo.GetByID(ctx, input.ChargeID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"charge not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get charge: %w", err)
	}

	if err := charge.Update(
		input.Amount,
		input.Tariff,
		input.Quantity,
		input.Description,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid charge data",
			err,
		)
	}

	if err := uc.chargeRepo.Update(ctx, charge); err != nil {
		return nil, fmt.Errorf("failed to update charge: %w", err)
	}

	return mapChargeToOutput(charge), nil
}

// ============================================================================
// DeleteChargeUseCase
// ============================================================================

type DeleteChargeUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

func NewDeleteChargeUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteChargeUseCase {
	return &DeleteChargeUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

type DeleteChargeInput struct {
	CurrentUserID int64
	ChargeID      int64
}

type DeleteChargeOutput struct {
	Success bool
	Message string
}

func (uc *DeleteChargeUseCase) Execute(ctx context.Context, input DeleteChargeInput) (*DeleteChargeOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	if err := uc.chargeRepo.SoftDelete(ctx, input.ChargeID); err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"charge not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to delete charge: %w", err)
	}

	return &DeleteChargeOutput{
		Success: true,
		Message: "Charge successfully deleted",
	}, nil
}

// ============================================================================
// ListChargesUseCase
// ============================================================================

type ListChargesUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

func NewListChargesUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *ListChargesUseCase {
	return &ListChargesUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

type ListChargesInput struct {
	CurrentUserID    int64
	IncludeDeleted   bool
	OwnershipShareID *int64
	ApartmentID      *int64
	OwnerID          *int64
	ChargeType       *entity.ChargeType
	PeriodMonth      *int
	PeriodYear       *int
	StartDate        *time.Time
	EndDate          *time.Time
	MinAmount        *float64
	MaxAmount        *float64
	IsOverdue        *bool
	SearchQuery      string
	Limit            int
	Offset           int
	OrderBy          string
	OrderDesc        bool
}

type ListChargesOutput struct {
	Charges []*ChargeOutput
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

func (uc *ListChargesUseCase) Execute(ctx context.Context, input ListChargesInput) (*ListChargesOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.ChargeFilter{
		IncludeDeleted:   input.IncludeDeleted,
		OwnershipShareID: input.OwnershipShareID,
		ApartmentID:      input.ApartmentID,
		OwnerID:          input.OwnerID,
		ChargeType:       input.ChargeType,
		PeriodMonth:      input.PeriodMonth,
		PeriodYear:       input.PeriodYear,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		MinAmount:        input.MinAmount,
		MaxAmount:        input.MaxAmount,
		IsOverdue:        input.IsOverdue,
		SearchQuery:      input.SearchQuery,
		Limit:            input.Limit,
		Offset:           input.Offset,
		OrderBy:          input.OrderBy,
		OrderDesc:        input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	charges, err := uc.chargeRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list charges: %w", err)
	}

	total, err := uc.chargeRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count charges: %w", err)
	}

	outputs := make([]*ChargeOutput, len(charges))
	for i, c := range charges {
		outputs[i] = mapChargeToOutput(c)
	}

	return &ListChargesOutput{
		Charges: outputs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
	}, nil
}

// ============================================================================
// GetChargeStatisticsUseCase
// ============================================================================

type GetChargeStatisticsUseCase struct {
	chargeRepo     repository.ChargeRepository
	permissionRepo repository.PermissionRepository
}

func NewGetChargeStatisticsUseCase(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *GetChargeStatisticsUseCase {
	return &GetChargeStatisticsUseCase{
		chargeRepo:     chargeRepo,
		permissionRepo: permissionRepo,
	}
}

type GetChargeStatisticsInput struct {
	CurrentUserID int64
	PeriodMonth   *int
	PeriodYear    *int
	StartDate     *time.Time
	EndDate       *time.Time
	ChargeType    *entity.ChargeType
	ApartmentID   *int64
	OwnerID       *int64
}

func (uc *GetChargeStatisticsUseCase) Execute(ctx context.Context, input GetChargeStatisticsInput) (*repository.ChargeStatistics, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "charges", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.ChargeStatisticsFilter{
		PeriodMonth: input.PeriodMonth,
		PeriodYear:  input.PeriodYear,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		ChargeType:  input.ChargeType,
		ApartmentID: input.ApartmentID,
		OwnerID:     input.OwnerID,
	}

	stats, err := uc.chargeRepo.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

// ============================================================================
// Helper
// ============================================================================

func mapChargeToOutput(c *entity.Charge) *ChargeOutput {
	return &ChargeOutput{
		ID:               c.ID,
		OwnershipShareID: c.OwnershipShareID,
		ChargeType:       string(c.ChargeType),
		ChargeDate:       c.ChargeDate,
		PeriodMonth:      c.PeriodMonth,
		PeriodYear:       c.PeriodYear,
		Amount:           c.Amount,
		Tariff:           c.Tariff,
		Quantity:         c.Quantity,
		Description:      c.Description,
		Notes:            c.Notes,
		PeriodDisplay:    c.GetPeriodDisplay(),
		AmountDisplay:    c.GetAmountDisplay(),
		TypeName:         c.ChargeType.GetDisplayName(),
	}
}
