// application/usecase/contractorpayment/get_contractorpayment.go
package contractorpayment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// GetContractorPaymentUseCase отримує платіж контрагента за ID.
type GetContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

// NewGetContractorPaymentUseCase створює новий use case.
func NewGetContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *GetContractorPaymentUseCase {
	return &GetContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

// GetContractorPaymentInput - вхідні дані.
type GetContractorPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

// Execute виконує отримання платежу.
func (uc *GetContractorPaymentUseCase) Execute(ctx context.Context, input GetContractorPaymentInput) (*ContractorPaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.contractorPaymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"contractor payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get contractor payment: %w", err)
	}

	return mapContractorPaymentToOutput(payment), nil
}
