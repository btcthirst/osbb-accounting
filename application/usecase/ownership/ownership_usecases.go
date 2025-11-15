// application/usecase/ownership/ownership_usecases.go
package ownership

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// OUTPUT TYPES
// ============================================================================

// OwnershipShareOutput - результат операцій з часткою власності.
type OwnershipShareOutput struct {
	ID                int64
	OwnerID           int64
	ApartmentID       int64
	ShareNumerator    int
	ShareDenominator  int
	ShareFraction     string
	SharePercentage   float64
	OwnershipType     entity.OwnershipType
	OwnershipTypeName string
	StartDate         time.Time
	EndDate           *time.Time
	DocumentType      *string
	DocumentNumber    *string
	DocumentDate      *time.Time
	Notes             *string
	IsActive          bool
	IsCurrentlyActive bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// OwnershipShareDetailsOutput - детальна інформація про частку.
type OwnershipShareDetailsOutput struct {
	OwnershipShareOutput
	OwnerName         string
	OwnerPhone        *string
	OwnerEmail        *string
	ApartmentNumber   string
	ApartmentFloor    int
	ApartmentEntrance *int
	AreaOwned         float64
}

// ============================================================================
// CreateOwnershipShareUseCase
// ============================================================================

type CreateOwnershipShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	ownerRepo      repository.OwnerRepository
	apartmentRepo  repository.ApartmentRepository
	permissionRepo repository.PermissionRepository
}

func NewCreateOwnershipShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	ownerRepo repository.OwnerRepository,
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *CreateOwnershipShareUseCase {
	return &CreateOwnershipShareUseCase{
		shareRepo:      shareRepo,
		ownerRepo:      ownerRepo,
		apartmentRepo:  apartmentRepo,
		permissionRepo: permissionRepo,
	}
}

type CreateOwnershipShareInput struct {
	CurrentUserID    int64
	OwnerID          int64
	ApartmentID      int64
	ShareNumerator   int
	ShareDenominator int
	OwnershipType    entity.OwnershipType
	StartDate        time.Time
	EndDate          *time.Time
	DocumentType     *string
	DocumentNumber   *string
	DocumentDate     *time.Time
	Notes            *string
}

func (uc *CreateOwnershipShareUseCase) Execute(
	ctx context.Context,
	input CreateOwnershipShareInput,
) (*OwnershipShareOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Перевірка існування власника
	owner, err := uc.ownerRepo.GetByID(ctx, input.OwnerID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"owner not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	// Перевірка існування квартири
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

	// Перевірка дублювання активної частки
	hasDuplicate, err := uc.shareRepo.CheckDuplicateActiveOwnership(
		ctx, input.OwnerID, input.ApartmentID, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if hasDuplicate {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			fmt.Sprintf("owner '%s' already has an active ownership share for apartment '%s'",
				owner.FullName(), apartment.GetDisplayName()),
			domainErrors.ErrAlreadyExists,
		)
	}

	// Перевірка суми часток (не більше 100%)
	totalShare, err := uc.shareRepo.CalculateTotalShareForApartment(
		ctx, input.ApartmentID, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total share: %w", err)
	}

	newShare := float64(input.ShareNumerator) / float64(input.ShareDenominator) * 100
	if totalShare+newShare > 100.01 { // Невелика похибка для float
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			fmt.Sprintf("total ownership share would exceed 100%% (current: %.2f%%, adding: %.2f%%)",
				totalShare, newShare),
			entity.ErrOwnershipShareExceedsTotal,
		)
	}

	// Створення entity
	share, err := entity.NewOwnershipShare(
		input.OwnerID,
		input.ApartmentID,
		input.ShareNumerator,
		input.ShareDenominator,
		input.OwnershipType,
		input.StartDate,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid ownership share data",
			err,
		)
	}

	// Додаткові поля
	share.EndDate = input.EndDate
	share.DocumentType = input.DocumentType
	share.DocumentNumber = input.DocumentNumber
	share.DocumentDate = input.DocumentDate
	share.Notes = input.Notes

	// Збереження
	if err := uc.shareRepo.Create(ctx, share); err != nil {
		return nil, fmt.Errorf("failed to create ownership share: %w", err)
	}

	return mapShareToOutput(share), nil
}

// ============================================================================
// GetOwnershipShareUseCase
// ============================================================================

type GetOwnershipShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewGetOwnershipShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *GetOwnershipShareUseCase {
	return &GetOwnershipShareUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type GetOwnershipShareInput struct {
	CurrentUserID int64
	ShareID       int64
}

func (uc *GetOwnershipShareUseCase) Execute(
	ctx context.Context,
	input GetOwnershipShareInput,
) (*OwnershipShareDetailsOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання з деталями
	details, err := uc.shareRepo.GetWithDetails(ctx, input.ShareID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"ownership share not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get ownership share: %w", err)
	}

	return mapDetailsToOutput(details), nil
}

// ============================================================================
// UpdateOwnershipShareUseCase
// ============================================================================

type UpdateOwnershipShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewUpdateOwnershipShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateOwnershipShareUseCase {
	return &UpdateOwnershipShareUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type UpdateOwnershipShareInput struct {
	CurrentUserID    int64
	ShareID          int64
	ShareNumerator   int
	ShareDenominator int
	OwnershipType    entity.OwnershipType
	EndDate          *time.Time
	DocumentType     *string
	DocumentNumber   *string
	DocumentDate     *time.Time
	Notes            *string
}

func (uc *UpdateOwnershipShareUseCase) Execute(
	ctx context.Context,
	input UpdateOwnershipShareInput,
) (*OwnershipShareOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання поточних даних
	share, err := uc.shareRepo.GetByID(ctx, input.ShareID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"ownership share not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get ownership share: %w", err)
	}

	// Оновлення
	if err := share.Update(
		input.ShareNumerator,
		input.ShareDenominator,
		input.OwnershipType,
		input.EndDate,
		input.DocumentType,
		input.DocumentNumber,
		input.DocumentDate,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid ownership share data",
			err,
		)
	}

	// Збереження
	if err := uc.shareRepo.Update(ctx, share); err != nil {
		return nil, fmt.Errorf("failed to update ownership share: %w", err)
	}

	// Отримуємо деталі для відповіді
	details, _ := uc.shareRepo.GetWithDetails(ctx, share.ID)

	return &OwnershipShareOutput{
		ID:                details.Share.ID,
		OwnerID:           details.Share.OwnerID,
		ApartmentID:       details.Share.ApartmentID,
		ShareNumerator:    details.Share.ShareNumerator,
		ShareDenominator:  details.Share.ShareDenominator,
		ShareFraction:     details.Share.GetShareFraction(),
		SharePercentage:   details.Share.GetSharePercentage(),
		OwnershipType:     details.Share.OwnershipType,
		OwnershipTypeName: details.Share.OwnershipType.GetDisplayName(),
		StartDate:         details.Share.StartDate,
		EndDate:           details.Share.EndDate,
		DocumentType:      details.Share.DocumentType,
		DocumentNumber:    details.Share.DocumentNumber,
		DocumentDate:      details.Share.DocumentDate,
		Notes:             details.Share.Notes,
		IsActive:          details.Share.IsActive,
		IsCurrentlyActive: details.Share.IsCurrentlyActive(),
		CreatedAt:         details.Share.CreatedAt,
		UpdatedAt:         details.Share.UpdatedAt,
	}, nil
}

// ============================================================================
// DeleteOwnershipShareUseCase
// ============================================================================

type DeleteOwnershipShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewDeleteOwnershipShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteOwnershipShareUseCase {
	return &DeleteOwnershipShareUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type DeleteOwnershipShareInput struct {
	CurrentUserID int64
	ShareID       int64
}

type DeleteOwnershipShareOutput struct {
	Success bool
	Message string
}

func (uc *DeleteOwnershipShareUseCase) Execute(
	ctx context.Context,
	input DeleteOwnershipShareInput,
) (*DeleteOwnershipShareOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Отримання для інформації
	details, err := uc.shareRepo.GetWithDetails(ctx, input.ShareID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"ownership share not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get ownership share: %w", err)
	}

	// Видалення
	if err := uc.shareRepo.SoftDelete(ctx, input.ShareID); err != nil {
		return nil, fmt.Errorf("failed to delete ownership share: %w", err)
	}

	return &DeleteOwnershipShareOutput{
		Success: true,
		Message: fmt.Sprintf("Ownership share deleted: %s - %s (%s)",
			details.OwnerName,
			details.ApartmentNumber,
			details.Share.GetShareDisplay()),
	}, nil
}

// ============================================================================
// ListOwnershipSharesUseCase
// ============================================================================

type ListOwnershipSharesUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewListOwnershipSharesUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *ListOwnershipSharesUseCase {
	return &ListOwnershipSharesUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type ListOwnershipSharesInput struct {
	CurrentUserID     int64
	IncludeDeleted    bool
	IsActive          *bool
	OwnerID           *int64
	ApartmentID       *int64
	OwnershipType     *entity.OwnershipType
	IsCurrentlyActive bool
	Limit             int
	Offset            int
	OrderBy           string
	OrderDesc         bool
}

type ListOwnershipSharesOutput struct {
	Shares  []*OwnershipShareDetailsOutput
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

func (uc *ListOwnershipSharesUseCase) Execute(
	ctx context.Context,
	input ListOwnershipSharesInput,
) (*ListOwnershipSharesOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Підготовка фільтру
	filter := repository.OwnershipShareFilter{
		IncludeDeleted:    input.IncludeDeleted,
		IsActive:          input.IsActive,
		OwnerID:           input.OwnerID,
		ApartmentID:       input.ApartmentID,
		OwnershipType:     input.OwnershipType,
		IsCurrentlyActive: input.IsCurrentlyActive,
		Limit:             input.Limit,
		Offset:            input.Offset,
		OrderBy:           input.OrderBy,
		OrderDesc:         input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Отримання списку з деталями
	detailsList, err := uc.shareRepo.ListWithDetails(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list ownership shares: %w", err)
	}

	// Підрахунок загальної кількості
	total, err := uc.shareRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count ownership shares: %w", err)
	}

	// Конвертація в output
	outputs := make([]*OwnershipShareDetailsOutput, len(detailsList))
	for i, details := range detailsList {
		outputs[i] = mapDetailsToOutput(details)
	}

	return &ListOwnershipSharesOutput{
		Shares:  outputs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: int64(filter.Offset+filter.Limit) < total,
	}, nil
}

// ============================================================================
// GetByApartmentUseCase - отримати всіх власників квартири
// ============================================================================

type GetByApartmentUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewGetByApartmentUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *GetByApartmentUseCase {
	return &GetByApartmentUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type GetByApartmentInput struct {
	CurrentUserID int64
	ApartmentID   int64
	OnlyActive    bool
}

func (uc *GetByApartmentUseCase) Execute(
	ctx context.Context,
	input GetByApartmentInput,
) ([]*OwnershipShareDetailsOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Фільтр
	filter := repository.OwnershipShareFilter{
		ApartmentID:       &input.ApartmentID,
		IsCurrentlyActive: input.OnlyActive,
		Limit:             1000,
		OrderBy:           "start_date",
		OrderDesc:         true,
	}

	// Отримання
	detailsList, err := uc.shareRepo.ListWithDetails(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ownership shares: %w", err)
	}

	// Конвертація
	outputs := make([]*OwnershipShareDetailsOutput, len(detailsList))
	for i, details := range detailsList {
		outputs[i] = mapDetailsToOutput(details)
	}

	return outputs, nil
}

// ============================================================================
// GetByOwnerUseCase - отримати всі квартири власника
// ============================================================================

type GetByOwnerUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewGetByOwnerUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *GetByOwnerUseCase {
	return &GetByOwnerUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type GetByOwnerInput struct {
	CurrentUserID int64
	OwnerID       int64
	OnlyActive    bool
}

func (uc *GetByOwnerUseCase) Execute(
	ctx context.Context,
	input GetByOwnerInput,
) ([]*OwnershipShareDetailsOutput, error) {
	// Перевірка дозволів
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	// Фільтр
	filter := repository.OwnershipShareFilter{
		OwnerID:           &input.OwnerID,
		IsCurrentlyActive: input.OnlyActive,
		Limit:             1000,
		OrderBy:           "start_date",
		OrderDesc:         true,
	}

	// Отримання
	detailsList, err := uc.shareRepo.ListWithDetails(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ownership shares: %w", err)
	}

	// Конвертація
	outputs := make([]*OwnershipShareDetailsOutput, len(detailsList))
	for i, details := range detailsList {
		outputs[i] = mapDetailsToOutput(details)
	}

	return outputs, nil
}

// ============================================================================
// CALCULATE REMAINING
// ============================================================================

type CalculateRemainingShareUseCase struct {
	shareRepo      repository.OwnershipShareRepository
	permissionRepo repository.PermissionRepository
}

func NewCalculateRemainingShareUseCase(
	shareRepo repository.OwnershipShareRepository,
	permissionRepo repository.PermissionRepository,
) *CalculateRemainingShareUseCase {
	return &CalculateRemainingShareUseCase{
		shareRepo:      shareRepo,
		permissionRepo: permissionRepo,
	}
}

type CalculateRemainingShareInput struct {
	CurrentUserID  int64
	ApartmentID    int64
	ExcludeShareID *int64
}

type CalculateRemainingShareOutput struct {
	TotalSharePercentage     float64
	RemainingSharePercentage float64
	CanAddFullShare          bool
}

func (uc *CalculateRemainingShareUseCase) Execute(ctx context.Context, input CalculateRemainingShareInput) (*CalculateRemainingShareOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "ownership", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	totalShare, err := uc.shareRepo.CalculateTotalShareForApartment(ctx, input.ApartmentID, input.ExcludeShareID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total share: %w", err)
	}

	remaining := 100.0 - totalShare

	return &CalculateRemainingShareOutput{
		TotalSharePercentage:     totalShare,
		RemainingSharePercentage: remaining,
		CanAddFullShare:          remaining >= 99.99, // Невелика похибка
	}, nil
}

// ============================================================================
// Helper функції
// ============================================================================

func mapShareToOutput(
	share *entity.OwnershipShare,
) *OwnershipShareOutput {
	return &OwnershipShareOutput{
		ID:                share.ID,
		OwnerID:           share.OwnerID,
		ApartmentID:       share.ApartmentID,
		ShareNumerator:    share.ShareNumerator,
		ShareDenominator:  share.ShareDenominator,
		ShareFraction:     share.GetShareFraction(),
		SharePercentage:   share.GetSharePercentage(),
		OwnershipType:     share.OwnershipType,
		OwnershipTypeName: share.OwnershipType.GetDisplayName(),
		StartDate:         share.StartDate,
		EndDate:           share.EndDate,
		DocumentType:      share.DocumentType,
		DocumentNumber:    share.DocumentNumber,
		DocumentDate:      share.DocumentDate,
		Notes:             share.Notes,
		IsActive:          share.IsActive,
		IsCurrentlyActive: share.IsCurrentlyActive(),
		CreatedAt:         share.CreatedAt,
		UpdatedAt:         share.UpdatedAt,
	}
}

func mapDetailsToOutput(details *repository.OwnershipShareDetails) *OwnershipShareDetailsOutput {
	areaOwned := 0.0
	if details.Share != nil {
		// Припускаємо що площа квартири * частка = площа у власності
		// (потрібно отримати area_total з apartments, але це можна додати пізніше)
		areaOwned = details.Share.GetSharePercentage()
	}

	return &OwnershipShareDetailsOutput{
		OwnershipShareOutput: OwnershipShareOutput{
			ID:                details.Share.ID,
			OwnerID:           details.Share.OwnerID,
			ApartmentID:       details.Share.ApartmentID,
			ShareNumerator:    details.Share.ShareNumerator,
			ShareDenominator:  details.Share.ShareDenominator,
			ShareFraction:     details.Share.GetShareFraction(),
			SharePercentage:   details.Share.GetSharePercentage(),
			OwnershipType:     details.Share.OwnershipType,
			OwnershipTypeName: details.Share.OwnershipType.GetDisplayName(),
			StartDate:         details.Share.StartDate,
			EndDate:           details.Share.EndDate,
			DocumentType:      details.Share.DocumentType,
			DocumentNumber:    details.Share.DocumentNumber,
			DocumentDate:      details.Share.DocumentDate,
			Notes:             details.Share.Notes,
			IsActive:          details.Share.IsActive,
			IsCurrentlyActive: details.Share.IsCurrentlyActive(),
			CreatedAt:         details.Share.CreatedAt,
			UpdatedAt:         details.Share.UpdatedAt,
		},
		OwnerName:         details.OwnerName,
		OwnerPhone:        details.OwnerPhone,
		OwnerEmail:        details.OwnerEmail,
		ApartmentNumber:   details.ApartmentNumber,
		ApartmentFloor:    details.ApartmentFloor,
		ApartmentEntrance: details.ApartmentEntrance,
		AreaOwned:         areaOwned,
	}
}
