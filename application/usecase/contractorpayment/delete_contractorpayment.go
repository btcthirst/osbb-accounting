// application/usecase/contractorpayment/delete_contractorpayment.go
package contractorpayment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// DeleteContractorPaymentUseCase видаляє платіж контрагента.
type DeleteContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

// NewDeleteContractorPaymentUseCase створює новий use case.
func NewDeleteContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteContractorPaymentUseCase {
	return &DeleteContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

// DeleteContractorPaymentInput - вхідні дані.
type DeleteContractorPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

// Execute виконує видалення платежу.
func (uc *DeleteContractorPaymentUseCase) Execute(ctx context.Context, input DeleteContractorPaymentInput) (*DeleteContractorPaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	if err := uc.contractorPaymentRepo.SoftDelete(ctx, input.PaymentID); err != nil {
		return nil, fmt.Errorf("failed to delete contractor payment: %w", err)
	}

	return &DeleteContractorPaymentOutput{
		Success: true,
		Message: "Contractor payment successfully deleted",
	}, nil
}
