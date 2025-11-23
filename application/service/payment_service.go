package service

import (
	"context"

	"osbb-accounting/application/usecase/payment"
	"osbb-accounting/domain/repository"
)

type PaymentServiceInterface interface {
	Create(context.Context, payment.CreatePaymentInput) (*payment.PaymentOutput, error)
	Get(context.Context, payment.GetPaymentInput) (*payment.PaymentOutput, error)
	Update(context.Context, payment.UpdatePaymentInput) (*payment.PaymentOutput, error)
	Delete(context.Context, payment.DeletePaymentInput) (*payment.DeletePaymentOutput, error)
	List(context.Context, payment.ListPaymentsInput) (*payment.ListPaymentsOutput, error)
	Approve(context.Context, payment.ApprovePaymentInput) (*payment.PaymentOutput, error)
	Unapprove(context.Context, payment.UnapprovePaymentInput) (*payment.PaymentOutput, error)
	GetStatistics(context.Context, payment.GetPaymentStatisticsInput) (*repository.PaymentStatistics, error)
}

// PaymentService - фасад для всіх операцій з платежами.
type PaymentService struct {
	createPayment    *payment.CreatePaymentUseCase
	getPayment       *payment.GetPaymentUseCase
	updatePayment    *payment.UpdatePaymentUseCase
	deletePayment    *payment.DeletePaymentUseCase
	listPayments     *payment.ListPaymentsUseCase
	approvePayment   *payment.ApprovePaymentUseCase
	unapprovePayment *payment.UnapprovePaymentUseCase
	getStatistics    *payment.GetPaymentStatisticsUseCase
}

// NewPaymentService створює новий PaymentService.
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	permissionRepo repository.PermissionRepository,
) *PaymentService {
	return &PaymentService{
		createPayment:    payment.NewCreatePaymentUseCase(paymentRepo, permissionRepo),
		getPayment:       payment.NewGetPaymentUseCase(paymentRepo, permissionRepo),
		updatePayment:    payment.NewUpdatePaymentUseCase(paymentRepo, permissionRepo),
		deletePayment:    payment.NewDeletePaymentUseCase(paymentRepo, permissionRepo),
		listPayments:     payment.NewListPaymentsUseCase(paymentRepo, permissionRepo),
		approvePayment:   payment.NewApprovePaymentUseCase(paymentRepo, permissionRepo),
		unapprovePayment: payment.NewUnapprovePaymentUseCase(paymentRepo, permissionRepo),
		getStatistics:    payment.NewGetPaymentStatisticsUseCase(paymentRepo, permissionRepo),
	}
}

func (s *PaymentService) Create(ctx context.Context, input payment.CreatePaymentInput) (*payment.PaymentOutput, error) {
	return s.createPayment.Execute(ctx, input)
}

func (s *PaymentService) Get(ctx context.Context, input payment.GetPaymentInput) (*payment.PaymentOutput, error) {
	return s.getPayment.Execute(ctx, input)
}

func (s *PaymentService) Update(ctx context.Context, input payment.UpdatePaymentInput) (*payment.PaymentOutput, error) {
	return s.updatePayment.Execute(ctx, input)
}

func (s *PaymentService) Delete(ctx context.Context, input payment.DeletePaymentInput) (*payment.DeletePaymentOutput, error) {
	return s.deletePayment.Execute(ctx, input)
}

func (s *PaymentService) List(ctx context.Context, input payment.ListPaymentsInput) (*payment.ListPaymentsOutput, error) {
	return s.listPayments.Execute(ctx, input)
}

func (s *PaymentService) Approve(ctx context.Context, input payment.ApprovePaymentInput) (*payment.PaymentOutput, error) {
	return s.approvePayment.Execute(ctx, input)
}

func (s *PaymentService) Unapprove(ctx context.Context, input payment.UnapprovePaymentInput) (*payment.PaymentOutput, error) {
	return s.unapprovePayment.Execute(ctx, input)
}

func (s *PaymentService) GetStatistics(ctx context.Context, input payment.GetPaymentStatisticsInput) (*repository.PaymentStatistics, error) {
	return s.getStatistics.Execute(ctx, input)
}
