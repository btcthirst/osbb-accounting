// application/usecase/contractorpayment/update_contractorpayment.go
package contractorpayment

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// UpdateContractorPaymentUseCase оновлює платіж контрагента.
type UpdateContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

// NewUpdateContractorPaymentUseCase створює новий use case.
func NewUpdateContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateContractorPaymentUseCase {
	return &UpdateContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

// UpdateContractorPaymentInput - вхідні дані.
type UpdateContractorPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
	Amount        float64
	PaymentMethod entity.PaymentMethod
	Purpose       string
	PaymentDate   time.Time
	PeriodMonth   *int
	PeriodYear    *int
	ReceiptNumber *string
	Notes         *string
}

// Execute виконує оновлення платежу.
func (uc *UpdateContractorPaymentUseCase) Execute(ctx context.Context, input UpdateContractorPaymentInput) (*ContractorPaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionUpdate,
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

	if err := payment.Update(
		input.Amount,
		input.PaymentMethod,
		input.Purpose,
		input.PaymentDate,
		input.ReceiptNumber,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid contractor payment data",
			err,
		)
	}

	payment.PeriodMonth = input.PeriodMonth
	payment.PeriodYear = input.PeriodYear

	if err := uc.contractorPaymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update contractor payment: %w", err)
	}

	return mapContractorPaymentToOutput(payment), nil
}
