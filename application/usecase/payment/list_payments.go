package payment

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ============================================================================
// ListPaymentsUseCase
// ============================================================================

type ListPaymentsUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewListPaymentsUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *ListPaymentsUseCase {
	return &ListPaymentsUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type ListPaymentsInput struct {
	CurrentUserID    int64
	IncludeDeleted   bool
	OwnershipShareID *int64
	OwnerID          *int64
	ApartmentID      *int64
	StartDate        *int64
	EndDate          *int64
	PeriodMonth      *int
	PeriodYear       *int
	PaymentMethod    *entity.PaymentMethod
	IsApproved       *bool
	ApprovedBy       *int64
	MinAmount        *float64
	MaxAmount        *float64
	SearchQuery      string
	Limit            int
	Offset           int
	OrderBy          string
	OrderDesc        bool
}

type ListPaymentsOutput struct {
	Payments []*PaymentOutput
	Total    int64
	Limit    int
	Offset   int
	HasMore  bool
}

func (uc *ListPaymentsUseCase) Execute(ctx context.Context, input ListPaymentsInput) (*ListPaymentsOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	filter := repository.PaymentFilter{
		IncludeDeleted:   input.IncludeDeleted,
		OwnershipShareID: input.OwnershipShareID,
		OwnerID:          input.OwnerID,
		ApartmentID:      input.ApartmentID,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		PeriodMonth:      input.PeriodMonth,
		PeriodYear:       input.PeriodYear,
		PaymentMethod:    input.PaymentMethod,
		IsApproved:       input.IsApproved,
		ApprovedBy:       input.ApprovedBy,
		MinAmount:        input.MinAmount,
		MaxAmount:        input.MaxAmount,
		SearchQuery:      input.SearchQuery,
		Limit:            input.Limit,
		Offset:           input.Offset,
		OrderBy:          input.OrderBy,
		OrderDesc:        input.OrderDesc,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	payments, err := uc.paymentRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	total, err := uc.paymentRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	outputs := make([]*PaymentOutput, len(payments))
	for i, p := range payments {
		outputs[i] = mapPaymentToOutput(p)
	}

	return &ListPaymentsOutput{
		Payments: outputs,
		Total:    total,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
		HasMore:  int64(filter.Offset+filter.Limit) < total,
	}, nil
}
