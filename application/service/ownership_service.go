// application/service/ownership_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/ownership"
	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/repository"
)

type OwnershipServiceInterface interface {
	Create(ctx context.Context, input ownership.CreateOwnershipShareInput) (*ownership.OwnershipShareOutput, error)
	Get(ctx context.Context, input ownership.GetOwnershipShareInput) (*ownership.OwnershipShareDetailsOutput, error)
	Update(ctx context.Context, input ownership.UpdateOwnershipShareInput) (*ownership.OwnershipShareOutput, error)
	Delete(ctx context.Context, input ownership.DeleteOwnershipShareInput) (*shared.DeleteOutput, error)
	List(ctx context.Context, input ownership.ListOwnershipSharesInput) (*ownership.ListOwnershipSharesOutput, error)
	GetByApartment(ctx context.Context, input ownership.GetByApartmentInput) ([]*ownership.OwnershipShareDetailsOutput, error)
	GetByOwner(ctx context.Context, input ownership.GetByOwnerInput) ([]*ownership.OwnershipShareDetailsOutput, error)
	CalculateRemaining(ctx context.Context, input ownership.CalculateRemainingShareInput) (*ownership.CalculateRemainingShareOutput, error)
}

// OwnershipService - фасад для всіх операцій з частками власності.
type OwnershipService struct {
	createShare        *ownership.CreateOwnershipShareUseCase
	getShare           *ownership.GetOwnershipShareUseCase
	updateShare        *ownership.UpdateOwnershipShareUseCase
	deleteShare        *ownership.DeleteOwnershipShareUseCase
	listShares         *ownership.ListOwnershipSharesUseCase
	getByApartment     *ownership.GetByApartmentUseCase
	getByOwner         *ownership.GetByOwnerUseCase
	calculateRemaining *ownership.CalculateRemainingShareUseCase
}

// NewOwnershipService створює новий OwnershipService.
func NewOwnershipService(
	shareRepo repository.OwnershipShareRepository,
	ownerRepo repository.OwnerRepository,
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *OwnershipService {
	return &OwnershipService{
		createShare: ownership.NewCreateOwnershipShareUseCase(
			shareRepo, ownerRepo, apartmentRepo, permissionRepo,
		),
		getShare: ownership.NewGetOwnershipShareUseCase(
			shareRepo, permissionRepo,
		),
		updateShare: ownership.NewUpdateOwnershipShareUseCase(
			shareRepo, permissionRepo,
		),
		deleteShare: ownership.NewDeleteOwnershipShareUseCase(
			shareRepo, permissionRepo,
		),
		listShares: ownership.NewListOwnershipSharesUseCase(
			shareRepo, permissionRepo,
		),
		getByApartment: ownership.NewGetByApartmentUseCase(
			shareRepo, permissionRepo,
		),
		getByOwner: ownership.NewGetByOwnerUseCase(
			shareRepo, permissionRepo,
		),
		calculateRemaining: ownership.NewCalculateRemainingShareUseCase(
			shareRepo, permissionRepo,
		),
	}
}

// Create створює нову частку власності.
func (s *OwnershipService) Create(
	ctx context.Context,
	input ownership.CreateOwnershipShareInput,
) (*ownership.OwnershipShareOutput, error) {
	return s.createShare.Execute(ctx, input)
}

// Get отримує частку власності за ID з деталями.
func (s *OwnershipService) Get(
	ctx context.Context,
	input ownership.GetOwnershipShareInput,
) (*ownership.OwnershipShareDetailsOutput, error) {
	return s.getShare.Execute(ctx, input)
}

// Update оновлює дані частки власності.
func (s *OwnershipService) Update(
	ctx context.Context,
	input ownership.UpdateOwnershipShareInput,
) (*ownership.OwnershipShareOutput, error) {
	return s.updateShare.Execute(ctx, input)
}

// Delete видаляє частку власності.
func (s *OwnershipService) Delete(
	ctx context.Context,
	input ownership.DeleteOwnershipShareInput,
) (*shared.DeleteOutput, error) {
	return s.deleteShare.Execute(ctx, input)
}

// List отримує список часток власності з фільтрацією.
func (s *OwnershipService) List(
	ctx context.Context,
	input ownership.ListOwnershipSharesInput,
) (*ownership.ListOwnershipSharesOutput, error) {
	return s.listShares.Execute(ctx, input)
}

// GetByApartment отримує всіх власників квартири.
func (s *OwnershipService) GetByApartment(
	ctx context.Context,
	input ownership.GetByApartmentInput,
) ([]*ownership.OwnershipShareDetailsOutput, error) {
	return s.getByApartment.Execute(ctx, input)
}

// GetByOwner отримує всі квартири власника.
func (s *OwnershipService) GetByOwner(
	ctx context.Context,
	input ownership.GetByOwnerInput,
) ([]*ownership.OwnershipShareDetailsOutput, error) {
	return s.getByOwner.Execute(ctx, input)
}

// CalculateRemaining розраховує залишок частки власності для квартири.
func (s *OwnershipService) CalculateRemaining(
	ctx context.Context,
	input ownership.CalculateRemainingShareInput,
) (*ownership.CalculateRemainingShareOutput, error) {
	return s.calculateRemaining.Execute(ctx, input)
}
