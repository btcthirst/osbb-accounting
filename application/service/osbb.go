package service

import (
	"context"
	"osbb-accounting/application/usecase/osbb"
	"osbb-accounting/domain/repository"
)

type OSBBServiceI interface {
	Create(ctx context.Context, input osbb.CreateOSBBInput) (*osbb.GetOSBBOutput, error)
	Update(ctx context.Context, input osbb.UpdateOSBBInput) (*osbb.GetOSBBOutput, error)
	Get(ctx context.Context, input osbb.GetOSBBInput) (*osbb.GetOSBBOutput, error)
}

// OSBBService - фасад для операцій з ОСББ.
type OSBBService struct {
	createOSBB *osbb.CreateOSBBUseCase
	updateOSBB *osbb.UpdateOSBBUseCase
	getOSBB    *osbb.GetOSBBUseCase
}

// NewOSBBService створює новий OSBBService.
func NewOSBBService(
	osbbRepo repository.OSBBRepository,
	permissionRepo repository.PermissionRepository,
) *OSBBService {
	return &OSBBService{
		createOSBB: osbb.NewCreateOSBBUseCase(osbbRepo, permissionRepo),
		updateOSBB: osbb.NewUpdateOSBBUseCase(osbbRepo, permissionRepo),
		getOSBB:    osbb.NewGetOSBBUseCase(osbbRepo, permissionRepo),
	}
}

// Create створює організацію ОСББ.
func (s *OSBBService) Create(ctx context.Context, input osbb.CreateOSBBInput) (*osbb.GetOSBBOutput, error) {
	return s.createOSBB.Execute(ctx, input)
}

// Update оновлює дані ОСББ.
func (s *OSBBService) Update(ctx context.Context, input osbb.UpdateOSBBInput) (*osbb.GetOSBBOutput, error) {
	return s.updateOSBB.Execute(ctx, input)
}

// Get отримує дані ОСББ.
func (s *OSBBService) Get(ctx context.Context, input osbb.GetOSBBInput) (*osbb.GetOSBBOutput, error) {
	return s.getOSBB.Execute(ctx, input)
}
