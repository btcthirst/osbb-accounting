package payment

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// PaymentOutput - результат операцій з платежем.
type PaymentOutput struct {
	ID               int64
	OwnershipShareID int64
	PaymentDate      time.Time
	Amount           float64
	PaymentMethod    string
	PaymentPurpose   string
	PeriodMonth      *int
	PeriodYear       *int
	ReceiptNumber    *string
	Notes            *string
	ApprovedBy       *int64
	ApprovedAt       *time.Time
	IsApproved       bool
	MethodName       string
}

// ============================================================================
// CreatePaymentUseCase
// ============================================================================

type CreatePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewCreatePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type CreatePaymentInput struct {
	CurrentUserID    int64
	OwnershipShareID int64
	Amount           float64
	PaymentMethod    entity.PaymentMethod
	PaymentPurpose   string
	PaymentDate      time.Time
	PeriodMonth      *int
	PeriodYear       *int
	ReceiptNumber    *string
	Notes            *string
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, input CreatePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionCreate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := entity.NewPayment(
		input.OwnershipShareID,
		input.Amount,
		input.PaymentMethod,
		input.PaymentPurpose,
		input.PaymentDate,
	)
	if err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid payment data",
			err,
		)
	}

	payment.PeriodMonth = input.PeriodMonth
	payment.PeriodYear = input.PeriodYear
	payment.ReceiptNumber = input.ReceiptNumber
	payment.Notes = input.Notes

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}

// ============================================================================
// GetPaymentUseCase
// ============================================================================

type GetPaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewGetPaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *GetPaymentUseCase {
	return &GetPaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type GetPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *GetPaymentUseCase) Execute(ctx context.Context, input GetPaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionRead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.paymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}

// ============================================================================
// UpdatePaymentUseCase
// ============================================================================

type UpdatePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewUpdatePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *UpdatePaymentUseCase {
	return &UpdatePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type UpdatePaymentInput struct {
	CurrentUserID  int64
	PaymentID      int64
	Amount         float64
	PaymentMethod  entity.PaymentMethod
	PaymentPurpose string
	PaymentDate    time.Time
	PeriodMonth    *int
	PeriodYear     *int
	ReceiptNumber  *string
	Notes          *string
}

func (uc *UpdatePaymentUseCase) Execute(ctx context.Context, input UpdatePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.paymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.IsApproved() {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"cannot update approved payment",
			nil,
		)
	}

	if err := payment.Update(
		input.Amount,
		input.PaymentMethod,
		input.PaymentPurpose,
		input.PaymentDate,
		input.ReceiptNumber,
		input.Notes,
	); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"invalid payment data",
			err,
		)
	}

	payment.PeriodMonth = input.PeriodMonth
	payment.PeriodYear = input.PeriodYear

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}

// ============================================================================
// DeletePaymentUseCase
// ============================================================================

type DeletePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewDeletePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *DeletePaymentUseCase {
	return &DeletePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type DeletePaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

type DeletePaymentOutput struct {
	Success bool
	Message string
}

func (uc *DeletePaymentUseCase) Execute(ctx context.Context, input DeletePaymentInput) (*DeletePaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionDelete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.paymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.IsApproved() {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"cannot delete approved payment",
			nil,
		)
	}

	if err := uc.paymentRepo.SoftDelete(ctx, input.PaymentID); err != nil {
		return nil, fmt.Errorf("failed to delete payment: %w", err)
	}

	return &DeletePaymentOutput{
		Success: true,
		Message: "Payment successfully deleted",
	}, nil
}

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

// ============================================================================
// ApprovePaymentUseCase
// ============================================================================

type ApprovePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewApprovePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *ApprovePaymentUseCase {
	return &ApprovePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type ApprovePaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *ApprovePaymentUseCase) Execute(ctx context.Context, input ApprovePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionUpdate, // Or a specific 'approve' action if defined
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.paymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if err := payment.Approve(input.CurrentUserID); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"failed to approve payment",
			err,
		)
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}

// ============================================================================
// UnapprovePaymentUseCase
// ============================================================================

type UnapprovePaymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewUnapprovePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *UnapprovePaymentUseCase {
	return &UnapprovePaymentUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type UnapprovePaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

func (uc *UnapprovePaymentUseCase) Execute(ctx context.Context, input UnapprovePaymentInput) (*PaymentOutput, error) {
	hasPermission, err := uc.permissionRepo.HasPermissionForResource(
		ctx, input.CurrentUserID, "payments", entity.ActionUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, domainErrors.ErrPermissionDenied
	}

	payment, err := uc.paymentRepo.GetByID(ctx, input.PaymentID)
	if err != nil {
		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			return nil, domainErrors.NewDomainError(
				domainErrors.CodeNotFound,
				"payment not found",
				err,
			)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if err := payment.Unapprove(); err != nil {
		return nil, domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			"failed to unapprove payment",
			err,
		)
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return mapPaymentToOutput(payment), nil
}

// ============================================================================
// GetPaymentStatisticsUseCase
// ============================================================================

type GetPaymentStatisticsUseCase struct {
	paymentRepo    repository.PaymentRepository
	permissionRepo repository.PermissionRepository
}

func NewGetPaymentStatisticsUseCase(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *GetPaymentStatisticsUseCase {
	return &GetPaymentStatisticsUseCase{
		paymentRepo:    paymentRepo,
		permissionRepo: permissionRepo,
	}
}

type GetPaymentStatisticsInput struct {
	CurrentUserID int64
	PeriodMonth   *int
	PeriodYear    *int
}

func (uc *GetPaymentStatisticsUseCase) Execute(ctx context.Context, input GetPaymentStatisticsInput) (*repository.PaymentStatistics, error) {
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
		PeriodMonth: input.PeriodMonth,
		PeriodYear:  input.PeriodYear,
	}

	stats, err := uc.paymentRepo.GetStatistics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

// ============================================================================
// Helper
// ============================================================================

func mapPaymentToOutput(p *entity.Payment) *PaymentOutput {
	return &PaymentOutput{
		ID:               p.ID,
		OwnershipShareID: p.OwnershipShareID,
		PaymentDate:      p.PaymentDate,
		Amount:           p.Amount,
		PaymentMethod:    string(p.PaymentMethod),
		PaymentPurpose:   p.PaymentPurpose,
		PeriodMonth:      p.PeriodMonth,
		PeriodYear:       p.PeriodYear,
		ReceiptNumber:    p.ReceiptNumber,
		Notes:            p.Notes,
		ApprovedBy:       p.ApprovedBy,
		ApprovedAt:       p.ApprovedAt,
		IsApproved:       p.IsApproved(),
		MethodName:       p.PaymentMethod.GetDisplayName(),
	}
}
