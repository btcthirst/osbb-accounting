package service

import (
	"context"

	"osbb-accounting/application/usecase/contractorpayment"
	"osbb-accounting/domain/repository"
)

type ContractorPaymentServiceInterface interface {
	Create(context.Context, contractorpayment.CreateContractorPaymentInput) (*contractorpayment.ContractorPaymentOutput, error)
	Get(context.Context, contractorpayment.GetContractorPaymentInput) (*contractorpayment.ContractorPaymentOutput, error)
	Update(context.Context, contractorpayment.UpdateContractorPaymentInput) (*contractorpayment.ContractorPaymentOutput, error)
	Delete(context.Context, contractorpayment.DeleteContractorPaymentInput) (*contractorpayment.DeleteContractorPaymentOutput, error)
	List(context.Context, contractorpayment.ListContractorPaymentsInput) (*contractorpayment.ListContractorPaymentsOutput, error)
}

// ContractorPaymentService - фасад для всіх операцій з платежами контрагентів.
type ContractorPaymentService struct {
	createPayment *contractorpayment.CreateContractorPaymentUseCase
	getPayment    *contractorpayment.GetContractorPaymentUseCase
	updatePayment *contractorpayment.UpdateContractorPaymentUseCase
	deletePayment *contractorpayment.DeleteContractorPaymentUseCase
	listPayments  *contractorpayment.ListContractorPaymentsUseCase
}

// NewContractorPaymentService створює новий ContractorPaymentService.
func NewContractorPaymentService(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *ContractorPaymentService {
	return &ContractorPaymentService{
		createPayment: contractorpayment.NewCreateContractorPaymentUseCase(contractorPaymentRepo, permissionRepo),
		getPayment:    contractorpayment.NewGetContractorPaymentUseCase(contractorPaymentRepo, permissionRepo),
		updatePayment: contractorpayment.NewUpdateContractorPaymentUseCase(contractorPaymentRepo, permissionRepo),
		deletePayment: contractorpayment.NewDeleteContractorPaymentUseCase(contractorPaymentRepo, permissionRepo),
		listPayments:  contractorpayment.NewListContractorPaymentsUseCase(contractorPaymentRepo, permissionRepo),
	}
}

func (s *ContractorPaymentService) Create(ctx context.Context, input contractorpayment.CreateContractorPaymentInput) (*contractorpayment.ContractorPaymentOutput, error) {
	return s.createPayment.Execute(ctx, input)
}

func (s *ContractorPaymentService) Get(ctx context.Context, input contractorpayment.GetContractorPaymentInput) (*contractorpayment.ContractorPaymentOutput, error) {
	return s.getPayment.Execute(ctx, input)
}

func (s *ContractorPaymentService) Update(ctx context.Context, input contractorpayment.UpdateContractorPaymentInput) (*contractorpayment.ContractorPaymentOutput, error) {
	return s.updatePayment.Execute(ctx, input)
}

func (s *ContractorPaymentService) Delete(ctx context.Context, input contractorpayment.DeleteContractorPaymentInput) (*contractorpayment.DeleteContractorPaymentOutput, error) {
	return s.deletePayment.Execute(ctx, input)
}

func (s *ContractorPaymentService) List(ctx context.Context, input contractorpayment.ListContractorPaymentsInput) (*contractorpayment.ListContractorPaymentsOutput, error) {
	return s.listPayments.Execute(ctx, input)
}
