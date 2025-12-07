package importusecase

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

type ProcessCashFlowImportBatchUseCase struct {
	batchRepo           repository.ImportBatchRepository
	cashFlowRepo        repository.ImportedCashFlowRecordRepository
	paymentRepo         repository.PaymentRepository
	expenseRepo         repository.ExpenseRepository
	contractorRepo      repository.ContractorRepository
	expenseCategoryRepo repository.ExpenseCategoryRepository
}

func NewProcessCashFlowImportBatchUseCase(
	batchRepo repository.ImportBatchRepository,
	cashFlowRepo repository.ImportedCashFlowRecordRepository,
	paymentRepo repository.PaymentRepository,
	expenseRepo repository.ExpenseRepository,
	contractorRepo repository.ContractorRepository,
	expenseCategoryRepo repository.ExpenseCategoryRepository,
) *ProcessCashFlowImportBatchUseCase {
	return &ProcessCashFlowImportBatchUseCase{
		batchRepo:           batchRepo,
		cashFlowRepo:        cashFlowRepo,
		paymentRepo:         paymentRepo,
		expenseRepo:         expenseRepo,
		contractorRepo:      contractorRepo,
		expenseCategoryRepo: expenseCategoryRepo,
	}
}

func (uc *ProcessCashFlowImportBatchUseCase) Execute(ctx context.Context, batchID int64) (*ProcessResult, error) {
	// 1. Get Batch
	batch, err := uc.batchRepo.FindByID(batchID)
	if err != nil {
		return nil, fmt.Errorf("batch not found: %w", err)
	}

	// 2. Get Records
	records, err := uc.cashFlowRepo.FindByBatchID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}

	result := &ProcessResult{
		SuccessCount: 0,
		FailCount:    0,
		Errors:       make([]string, 0),
	}

	// 3. Process each record
	for _, record := range records {
		if record.IsMigrated {
			continue
		}

		err := uc.processRecord(ctx, record)
		if err != nil {
			result.FailCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Record %d: %v", record.ID, err))
		} else {
			result.SuccessCount++
			record.IsMigrated = true
			_ = uc.cashFlowRepo.Update(ctx, record)
		}
	}

	// 4. Update Batch Status
	status := entity.ImportStatusCompleted
	if result.FailCount > 0 {
		status = entity.ImportStatusFailed // Or partial
	}

	// Only update if we processed something
	if len(records) > 0 {
		_ = uc.batchRepo.UpdateStatus(batch.ID, status, nil)
	}

	return result, nil
}

func (uc *ProcessCashFlowImportBatchUseCase) processRecord(ctx context.Context, record *entity.ImportedCashFlowRecord) error {
	// 1. Find or Create Contractor
	contractor, err := uc.getOrCreateContractor(ctx, record.ContractorName)
	if err != nil {
		return fmt.Errorf("failed to handle contractor: %w", err)
	}

	if record.OperationType == entity.CashFlowOperationDebit {
		// Create Payment (Income)
		payment := &entity.Payment{
			Amount:        record.Amount,
			PaymentDate:   record.Date,
			Notes:         &record.Description,
			ContractorID:  &contractor.ID,
			PaymentMethod: entity.PaymentMethodBankTransfer,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := uc.paymentRepo.Create(ctx, payment); err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

	} else if record.OperationType == entity.CashFlowOperationCredit {
		// Create Expense (Outcome)

		// Find Category by Code
		category, err := uc.getCategoryByCode(ctx, record.CategoryCode)
		if err != nil {
			return fmt.Errorf("failed to find category %s: %w", record.CategoryCode, err)
		}

		expense := &entity.Expense{
			Amount:       record.Amount,
			ExpenseDate:  record.Date,
			Description:  record.Description,
			ContractorID: &contractor.ID,
			CategoryID:   category.ID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := uc.expenseRepo.Create(ctx, expense); err != nil {
			return fmt.Errorf("failed to create expense: %w", err)
		}
	}

	return nil
}

func (uc *ProcessCashFlowImportBatchUseCase) getOrCreateContractor(ctx context.Context, name string) (*entity.Contractor, error) {
	// Try to find by name using SearchQuery filter
	filter := repository.ContractorFilter{
		SearchQuery: name,
	}
	contractors, err := uc.contractorRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, c := range contractors {
		if c.Name == name {
			return c, nil
		}
	}

	// Create new
	newContractor := &entity.Contractor{
		Name:           name,
		ContractorType: entity.ContractorTypeOther,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := uc.contractorRepo.Create(ctx, newContractor); err != nil {
		return nil, err
	}
	return newContractor, nil
}

func (uc *ProcessCashFlowImportBatchUseCase) getCategoryByCode(ctx context.Context, code string) (*entity.ExpenseCategory, error) {
	// Try to find by code
	category, err := uc.expenseCategoryRepo.GetByCode(ctx, code)
	if err == nil {
		return category, nil
	}

	// If not found (or other error), create new.
	// Ideally we should check if error is "not found", but for now we assume so.

	// Create new category
	newCategory := &entity.ExpenseCategory{
		Name:         code,
		Code:         &code,
		Description:  &code, // Or "Code 63"
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CategoryType: entity.CategoryTypeOther, // Default type
	}
	if err := uc.expenseCategoryRepo.Create(ctx, newCategory); err != nil {
		return nil, err
	}
	return newCategory, nil
}
