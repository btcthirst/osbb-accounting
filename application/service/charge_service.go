package service

import (
	"context"

	"osbb-accounting/application/usecase/charge"
	"osbb-accounting/domain/repository"
)

type ChargeServiceInterface interface {
	Create(ctx context.Context, input charge.CreateChargeInput) (*charge.ChargeOutput, error)
	Get(ctx context.Context, input charge.GetChargeInput) (*charge.ChargeOutput, error)
	Update(ctx context.Context, input charge.UpdateChargeInput) (*charge.ChargeOutput, error)
	Delete(ctx context.Context, input charge.DeleteChargeInput) (*charge.DeleteChargeOutput, error)
	List(ctx context.Context, input charge.ListChargesInput) (*charge.ListChargesOutput, error)
	GetStatistics(ctx context.Context, input charge.GetChargeStatisticsInput) (*repository.ChargeStatistics, error)
}

// ChargeService - фасад для всіх операцій з нарахуваннями.
type ChargeService struct {
	createCharge  *charge.CreateChargeUseCase
	getCharge     *charge.GetChargeUseCase
	updateCharge  *charge.UpdateChargeUseCase
	deleteCharge  *charge.DeleteChargeUseCase
	listCharges   *charge.ListChargesUseCase
	getStatistics *charge.GetChargeStatisticsUseCase
}

// NewChargeService створює новий ChargeService.
func NewChargeService(
	chargeRepo repository.ChargeRepository,
	permissionRepo repository.PermissionRepository,
) *ChargeService {
	return &ChargeService{
		createCharge:  charge.NewCreateChargeUseCase(chargeRepo, permissionRepo),
		getCharge:     charge.NewGetChargeUseCase(chargeRepo, permissionRepo),
		updateCharge:  charge.NewUpdateChargeUseCase(chargeRepo, permissionRepo),
		deleteCharge:  charge.NewDeleteChargeUseCase(chargeRepo, permissionRepo),
		listCharges:   charge.NewListChargesUseCase(chargeRepo, permissionRepo),
		getStatistics: charge.NewGetChargeStatisticsUseCase(chargeRepo, permissionRepo),
	}
}

// Create створює нове нарахування.
func (s *ChargeService) Create(ctx context.Context, input charge.CreateChargeInput) (*charge.ChargeOutput, error) {
	return s.createCharge.Execute(ctx, input)
}

// Get отримує нарахування за ID.
func (s *ChargeService) Get(ctx context.Context, input charge.GetChargeInput) (*charge.ChargeOutput, error) {
	return s.getCharge.Execute(ctx, input)
}

// Update оновлює дані нарахування.
func (s *ChargeService) Update(ctx context.Context, input charge.UpdateChargeInput) (*charge.ChargeOutput, error) {
	return s.updateCharge.Execute(ctx, input)
}

// Delete видаляє нарахування.
func (s *ChargeService) Delete(ctx context.Context, input charge.DeleteChargeInput) (*charge.DeleteChargeOutput, error) {
	return s.deleteCharge.Execute(ctx, input)
}

// List отримує список нарахувань з фільтрацією.
func (s *ChargeService) List(ctx context.Context, input charge.ListChargesInput) (*charge.ListChargesOutput, error) {
	return s.listCharges.Execute(ctx, input)
}

// GetStatistics отримує статистику по нарахуванням.
func (s *ChargeService) GetStatistics(ctx context.Context, input charge.GetChargeStatisticsInput) (*repository.ChargeStatistics, error) {
	return s.getStatistics.Execute(ctx, input)
}
