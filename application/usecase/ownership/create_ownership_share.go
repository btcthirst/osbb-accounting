package ownership

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
	"time"
)

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
