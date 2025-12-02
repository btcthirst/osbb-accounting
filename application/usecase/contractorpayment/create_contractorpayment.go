// application/usecase/contractorpayment/create_contractorpayment.go
package contractorpayment

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// CreateContractorPaymentUseCase створює новий платіж контрагента.
type CreateContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

// NewCreateContractorPaymentUseCase створює новий use case.
func NewCreateContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *CreateContractorPaymentUseCase {
	return &CreateContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

// CreateContractorPaymentInput - вхідні дані.
type CreateContractorPaymentInput struct {
	CurrentUserID int64
	ContractorID  int64
	Amount        float64
	PaymentMethod entity.PaymentMethod
	Purpose       string
	PaymentDate   time.Time
	PeriodMonth   *int
	PeriodYear    *int
	ReceiptNumber *string
	Notes         *string
}

// Execute виконує створення платежу.
func (uc *CreateContractorPaymentUseCase) Execute(ctx context.Context, input CreateContractorPaymentInput) (*ContractorPaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := entity.NewContractorPayment(
		input.ContractorID,
		input.Amount,
		input.PaymentMethod,
		input.Purpose,
		input.PaymentDate,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid contractor payment data",
			err,
		)
	}

	payment.PeriodMonth = input.PeriodMonth
	payment.PeriodYear = input.PeriodYear
	payment.ReceiptNumber = input.ReceiptNumber
	payment.Notes = input.Notes

	if err := uc.contractorPaymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create contractor payment: %w", err)
	}

	return mapContractorPaymentToOutput(payment), nil
}
