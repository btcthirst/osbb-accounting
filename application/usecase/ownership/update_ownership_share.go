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
