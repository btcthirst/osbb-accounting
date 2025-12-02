// application/service/owner_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/owner"
	"osbb-accounting/application/usecase/shared"
	"osbb-accounting/domain/repository"
)

type OwnerServiceInterface interface {
	Create(ctx context.Context, input owner.CreateOwnerInput) (*owner.OwnerOutput, error)
	Get(ctx context.Context, input owner.GetOwnerInput) (*owner.OwnerOutput, error)
	Update(ctx context.Context, input owner.UpdateOwnerInput) (*owner.OwnerOutput, error)
	Delete(ctx context.Context, input owner.DeleteOwnerInput) (*shared.DeleteOutput, error)
	List(ctx context.Context, input owner.ListOwnersInput) (*owner.ListOwnersOutput, error)
	Search(ctx context.Context, input owner.SearchOwnersInput) ([]*owner.OwnerOutput, error)
}

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
func (s *OwnerService) Delete(ctx context.Context, input owner.DeleteOwnerInput) (*shared.DeleteOutput, error) {
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
