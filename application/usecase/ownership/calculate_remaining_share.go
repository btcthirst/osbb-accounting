// application/usecase/ownership/ownership_usecases.go
package ownership

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

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
