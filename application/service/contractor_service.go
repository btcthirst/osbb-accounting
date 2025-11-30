// application/service/contractor_service.go
package service

import (
	"context"

	"osbb-accounting/application/usecase/contractor"
	"osbb-accounting/domain/repository"
)

type ContractorServiceInterface interface {
	Create(ctx context.Context, input contractor.CreateContractorInput) (*contractor.ContractorOutput, error)
	Get(ctx context.Context, input contractor.GetContractorInput) (*contractor.ContractorOutput, error)
	Update(ctx context.Context, input contractor.UpdateContractorInput) (*contractor.ContractorOutput, error)
	Delete(ctx context.Context, input contractor.DeleteContractorInput) (*contractor.DeleteContractorOutput, error)
	List(ctx context.Context, input contractor.ListContractorsInput) (*contractor.ListContractorsOutput, error)
	Search(ctx context.Context, input contractor.SearchContractorsInput) ([]*contractor.ContractorOutput, error)
}

// ContractorService - фасад для всіх операцій з контрагентами.
type ContractorService struct {
	createContractor  *contractor.CreateContractorUseCase
	getContractor     *contractor.GetContractorUseCase
	updateContractor  *contractor.UpdateContractorUseCase
	deleteContractor  *contractor.DeleteContractorUseCase
	listContractors   *contractor.ListContractorsUseCase
	searchContractors *contractor.SearchContractorsUseCase
}

// NewContractorService створює новий ContractorService.
func NewContractorService(
	contractorRepo repository.ContractorRepository,
	permissionRepo repository.PermissionRepository,
) *ContractorService {
	return &ContractorService{
		createContractor:  contractor.NewCreateContractorUseCase(contractorRepo, permissionRepo),
		getContractor:     contractor.NewGetContractorUseCase(contractorRepo, permissionRepo),
		updateContractor:  contractor.NewUpdateContractorUseCase(contractorRepo, permissionRepo),
		deleteContractor:  contractor.NewDeleteContractorUseCase(contractorRepo, permissionRepo),
		listContractors:   contractor.NewListContractorsUseCase(contractorRepo, permissionRepo),
		searchContractors: contractor.NewSearchContractorsUseCase(contractorRepo, permissionRepo),
	}
}

// Create створює нового контрагента.
func (s *ContractorService) Create(ctx context.Context, input contractor.CreateContractorInput) (*contractor.ContractorOutput, error) {
	return s.createContractor.Execute(ctx, input)
}

// Get отримує контрагента за ID.
func (s *ContractorService) Get(ctx context.Context, input contractor.GetContractorInput) (*contractor.ContractorOutput, error) {
	return s.getContractor.Execute(ctx, input)
}

// Update оновлює дані контрагента.
func (s *ContractorService) Update(ctx context.Context, input contractor.UpdateContractorInput) (*contractor.ContractorOutput, error) {
	return s.updateContractor.Execute(ctx, input)
}

// Delete видаляє контрагента.
func (s *ContractorService) Delete(ctx context.Context, input contractor.DeleteContractorInput) (*contractor.DeleteContractorOutput, error) {
	return s.deleteContractor.Execute(ctx, input)
}

// List отримує список контрагентів з фільтрацією.
func (s *ContractorService) List(ctx context.Context, input contractor.ListContractorsInput) (*contractor.ListContractorsOutput, error) {
	return s.listContractors.Execute(ctx, input)
}

// Search шукає контрагентів (для autocomplete).
func (s *ContractorService) Search(ctx context.Context, input contractor.SearchContractorsInput) ([]*contractor.ContractorOutput, error) {
	return s.searchContractors.Execute(ctx, input)
}
