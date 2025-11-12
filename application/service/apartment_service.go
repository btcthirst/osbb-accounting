// application/service/apartment_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/domain/repository"
)

// ApartmentService - фасад для всіх операцій з квартирами.
type ApartmentService struct {
	createApartment *apartment.CreateApartmentUseCase
	getApartment    *apartment.GetApartmentUseCase
	updateApartment *apartment.UpdateApartmentUseCase
	deleteApartment *apartment.DeleteApartmentUseCase
	listApartments  *apartment.ListApartmentsUseCase
	getStatistics   *apartment.GetApartmentStatisticsUseCase
}

// NewApartmentService створює новий ApartmentService.
func NewApartmentService(
	apartmentRepo repository.ApartmentRepository,
	permissionRepo repository.PermissionRepository,
) *ApartmentService {
	return &ApartmentService{
		createApartment: apartment.NewCreateApartmentUseCase(apartmentRepo, permissionRepo),
		getApartment:    apartment.NewGetApartmentUseCase(apartmentRepo, permissionRepo),
		updateApartment: apartment.NewUpdateApartmentUseCase(apartmentRepo, permissionRepo),
		deleteApartment: apartment.NewDeleteApartmentUseCase(apartmentRepo, permissionRepo),
		listApartments:  apartment.NewListApartmentsUseCase(apartmentRepo, permissionRepo),
		getStatistics:   apartment.NewGetApartmentStatisticsUseCase(apartmentRepo, permissionRepo),
	}
}

func (s *ApartmentService) Create(ctx context.Context, input apartment.CreateApartmentInput) (*apartment.ApartmentOutput, error) {
	return s.createApartment.Execute(ctx, input)
}

func (s *ApartmentService) Get(ctx context.Context, input apartment.GetApartmentInput) (*apartment.ApartmentOutput, error) {
	return s.getApartment.Execute(ctx, input)
}

func (s *ApartmentService) Update(ctx context.Context, input apartment.UpdateApartmentInput) (*apartment.ApartmentOutput, error) {
	return s.updateApartment.Execute(ctx, input)
}

func (s *ApartmentService) Delete(ctx context.Context, input apartment.DeleteApartmentInput) (*apartment.DeleteApartmentOutput, error) {
	return s.deleteApartment.Execute(ctx, input)
}

func (s *ApartmentService) List(ctx context.Context, input apartment.ListApartmentsInput) (*apartment.ListApartmentsOutput, error) {
	return s.listApartments.Execute(ctx, input)
}

func (s *ApartmentService) GetStatistics(ctx context.Context, input apartment.GetApartmentStatisticsInput) (*apartment.ApartmentStatisticsOutput, error) {
	return s.getStatistics.Execute(ctx, input)
}
