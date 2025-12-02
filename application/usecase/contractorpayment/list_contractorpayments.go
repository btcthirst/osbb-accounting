// application/usecase/contractorpayment/list_contractorpayments.go
package contractorpayment

import (
	"context"
	"fmt"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ListContractorPaymentsUseCase отримує список платежів контрагентів.
type ListContractorPaymentsUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

// NewListContractorPaymentsUseCase створює новий use case.
func NewListContractorPaymentsUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *ListContractorPaymentsUseCase {
	return &ListContractorPaymentsUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

// ListContractorPaymentsInput - вхідні дані.
type ListContractorPaymentsInput struct {
	CurrentUserID  int64
	IncludeDeleted bool
	ContractorID   *int64
	StartDate      *int64
	EndDate        *int64
	PeriodMonth    *int
	PeriodYear     *int
	PaymentMethod  *entity.PaymentMethod
	MinAmount      *float64
	MaxAmount      *float64
	SearchQuery    string
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

// ListContractorPaymentsOutput - результат.
type ListContractorPaymentsOutput struct {
	Payments []*ContractorPaymentOutput
	Total    int64
	Limit    int
	Offset   int
	HasMore  bool
}

// Execute виконує отримання списку платежів.
func (uc *ListContractorPaymentsUseCase) Execute(ctx context.Context, input ListContractorPaymentsInput) (*ListContractorPaymentsOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.ContractorPaymentFilter{
		IncludeDeleted: input.IncludeDeleted,
		ContractorID:   input.ContractorID,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		PeriodMonth:    input.PeriodMonth,
		PeriodYear:     input.PeriodYear,
		PaymentMethod:  input.PaymentMethod,
		MinAmount:      input.MinAmount,
		MaxAmount:      input.MaxAmount,
		SearchQuery:    input.SearchQuery,
		Limit:          input.Limit,
		Offset:         input.Offset,
		OrderBy:        input.OrderBy,
		OrderDesc:      input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	payments, err := uc.contractorPaymentRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list contractor payments: %w", err)
	}

	total, err := uc.contractorPaymentRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count contractor payments: %w", err)
	}

	outputs := make([]*ContractorPaymentOutput, len(payments))
	for i, p := range payments {
		outputs[i] = mapContractorPaymentToOutput(p)
	}

	return &ListContractorPaymentsOutput{
		Payments: outputs,
		Total:    total,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
		HasMore:  int64(filter.Offset+filter.Limit) < total,
	}, nil
}
