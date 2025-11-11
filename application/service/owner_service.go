// application/service/owner_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/osbb"
	"osbb-accounting/application/usecase/owner"
	"osbb-accounting/domain/repository"
)

// OwnerService - фасад для всіх операцій з власниками.
type OwnerService struct {
	createOwner  *owner.CreateOwnerUseCase
	getOwner     *owner.GetOwnerUseCase
	updateOwner  *owner.UpdateOwnerUseCase
	deleteOwner  *owner.DeleteOwnerUseCase
	listOwners   *owner.ListOwnersUseCase
	searchOwners *owner.SearchOwnersUseCase
}

// NewOwnerService створює новий OwnerService.
func NewOwnerService(
	ownerRepo repository.OwnerRepository,
	permissionRepo repository.PermissionRepository,
) *OwnerService {
	return &OwnerService{
		createOwner:  owner.NewCreateOwnerUseCase(ownerRepo, permissionRepo),
		getOwner:     owner.NewGetOwnerUseCase(ownerRepo, permissionRepo),
		updateOwner:  owner.NewUpdateOwnerUseCase(ownerRepo, permissionRepo),
		deleteOwner:  owner.NewDeleteOwnerUseCase(ownerRepo, permissionRepo),
		listOwners:   owner.NewListOwnersUseCase(ownerRepo, permissionRepo),
		searchOwners: owner.NewSearchOwnersUseCase(ownerRepo, permissionRepo),
	}
}

// Create створює нового власника.
func (s *OwnerService) Create(ctx context.Context, input owner.CreateOwnerInput) (*owner.OwnerOutput, error) {
	return s.createOwner.Execute(ctx, input)
}

// Get отримує власника за ID.
func (s *OwnerService) Get(ctx context.Context, input owner.GetOwnerInput) (*owner.OwnerOutput, error) {
	return s.getOwner.Execute(ctx, input)
}

// Update оновлює дані власника.
func (s *OwnerService) Update(ctx context.Context, input owner.UpdateOwnerInput) (*owner.OwnerOutput, error) {
	return s.updateOwner.Execute(ctx, input)
}

// Delete видаляє власника.
func (s *OwnerService) Delete(ctx context.Context, input owner.DeleteOwnerInput) (*owner.DeleteOwnerOutput, error) {
	return s.deleteOwner.Execute(ctx, input)
}

// List отримує список власників з фільтрацією.
func (s *OwnerService) List(ctx context.Context, input owner.ListOwnersInput) (*owner.ListOwnersOutput, error) {
	return s.listOwners.Execute(ctx, input)
}

// Search шукає власників (для autocomplete).
func (s *OwnerService) Search(ctx context.Context, input owner.SearchOwnersInput) ([]*owner.OwnerOutput, error) {
	return s.searchOwners.Execute(ctx, input)
}

// ============================================================================

// OSBBService - фасад для операцій з ОСББ.
type OSBBService struct {
	getOSBB    *osbb.GetOSBBUseCase
	createOSBB *osbb.CreateOSBBUseCase
	updateOSBB *osbb.UpdateOSBBUseCase
}

// NewOSBBService створює новий OSBBService.
func NewOSBBService(
	osbbRepo repository.OSBBRepository,
	permissionRepo repository.PermissionRepository,
) *OSBBService {
	return &OSBBService{
		getOSBB:    osbb.NewGetOSBBUseCase(osbbRepo, permissionRepo),
		createOSBB: osbb.NewCreateOSBBUseCase(osbbRepo, permissionRepo),
		updateOSBB: osbb.NewUpdateOSBBUseCase(osbbRepo, permissionRepo),
	}
}

// Get отримує дані ОСББ.
func (s *OSBBService) Get(ctx context.Context, input osbb.GetOSBBInput) (*osbb.GetOSBBOutput, error) {
	return s.getOSBB.Execute(ctx, input)
}

// Create створює організацію ОСББ.
func (s *OSBBService) Create(ctx context.Context, input osbb.CreateOSBBInput) (*osbb.GetOSBBOutput, error) {
	return s.createOSBB.Execute(ctx, input)
}

// Update оновлює дані ОСББ.
func (s *OSBBService) Update(ctx context.Context, input osbb.UpdateOSBBInput) (*osbb.GetOSBBOutput, error) {
	return s.updateOSBB.Execute(ctx, input)
}
