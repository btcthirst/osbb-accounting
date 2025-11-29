package contractorpayment

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// ContractorPaymentOutput - результат операцій з платежем контрагента.
type ContractorPaymentOutput struct {
	ID            int64
	ContractorID  int64
	PaymentDate   time.Time
	Amount        float64
	PaymentMethod string
	Purpose       string
	PeriodMonth   *int
	PeriodYear    *int
	ReceiptNumber *string
	Notes         *string
	MethodName    string
}

// ============================================================================
// CreateContractorPaymentUseCase
// ============================================================================

type CreateContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

func NewCreateContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *CreateContractorPaymentUseCase {
	return &CreateContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

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

// ============================================================================
// GetContractorPaymentUseCase
// ============================================================================

type GetContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

func NewGetContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *GetContractorPaymentUseCase {
	return &GetContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

type GetContractorPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

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

// ============================================================================
// UpdateContractorPaymentUseCase
// ============================================================================

type UpdateContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

func NewUpdateContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *UpdateContractorPaymentUseCase {
	return &UpdateContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

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

// ============================================================================
// DeleteContractorPaymentUseCase
// ============================================================================

type DeleteContractorPaymentUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

func NewDeleteContractorPaymentUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *DeleteContractorPaymentUseCase {
	return &DeleteContractorPaymentUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

type DeleteContractorPaymentInput struct {
	CurrentUserID int64
	PaymentID     int64
}

type DeleteContractorPaymentOutput struct {
	Success bool
	Message string
}

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

// ============================================================================
// ListContractorPaymentsUseCase
// ============================================================================

type ListContractorPaymentsUseCase struct {
	contractorPaymentRepo repository.ContractorPaymentRepository
	permissionRepo        repository.PermissionRepository
}

func NewListContractorPaymentsUseCase(
	contractorPaymentRepo repository.ContractorPaymentRepository,
	permissionRepo repository.PermissionRepository,
) *ListContractorPaymentsUseCase {
	return &ListContractorPaymentsUseCase{
		contractorPaymentRepo: contractorPaymentRepo,
		permissionRepo:        permissionRepo,
	}
}

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

type ListContractorPaymentsOutput struct {
	Payments []*ContractorPaymentOutput
	Total    int64
	Limit    int
	Offset   int
	HasMore  bool
}

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

// ============================================================================
// Helper
// ============================================================================

func mapContractorPaymentToOutput(p *entity.ContractorPayment) *ContractorPaymentOutput {
	return &ContractorPaymentOutput{
		ID:            p.ID,
		ContractorID:  p.ContractorID,
		PaymentDate:   p.PaymentDate,
		Amount:        p.Amount,
		PaymentMethod: string(p.PaymentMethod),
		Purpose:       p.Purpose,
		PeriodMonth:   p.PeriodMonth,
		PeriodYear:    p.PeriodYear,
		ReceiptNumber: p.ReceiptNumber,
		Notes:         p.Notes,
		MethodName:    p.PaymentMethod.GetDisplayName(),
	}
}
