package importusecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

type ProcessCashFlowImportBatchUseCase struct {
	batchRepo             repository.ImportBatchRepository
	cashFlowRepo          repository.ImportedCashFlowRecordRepository
	paymentRepo           repository.PaymentRepository
	contractorPaymentRepo repository.ContractorPaymentRepository
	expenseRepo           repository.ExpenseRepository
	contractorRepo        repository.ContractorRepository
	expenseCategoryRepo   repository.ExpenseCategoryRepository
	ownershipRepo         repository.OwnershipShareRepository
	ownerRepo             repository.OwnerRepository
}

func NewProcessCashFlowImportBatchUseCase(
	batchRepo repository.ImportBatchRepository,
	cashFlowRepo repository.ImportedCashFlowRecordRepository,
	paymentRepo repository.PaymentRepository,
	contractorPaymentRepo repository.ContractorPaymentRepository,
	expenseRepo repository.ExpenseRepository,
	contractorRepo repository.ContractorRepository,
	expenseCategoryRepo repository.ExpenseCategoryRepository,
	ownershipRepo repository.OwnershipShareRepository,
	ownerRepo repository.OwnerRepository,
) *ProcessCashFlowImportBatchUseCase {
	return &ProcessCashFlowImportBatchUseCase{
		batchRepo:             batchRepo,
		cashFlowRepo:          cashFlowRepo,
		paymentRepo:           paymentRepo,
		contractorPaymentRepo: contractorPaymentRepo,
		expenseRepo:           expenseRepo,
		contractorRepo:        contractorRepo,
		expenseCategoryRepo:   expenseCategoryRepo,
		ownershipRepo:         ownershipRepo,
		ownerRepo:             ownerRepo,
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
		// Create Payment (Income) or ContractorPayment
		// Find ownership share by contractor name
		ownershipShare, err := uc.findOwnershipShareByName(ctx, record.ContractorName)

		if err == nil && ownershipShare != nil {
			// Found ownership share - create regular Payment (from apartment owner)
			payment := &entity.Payment{
				OwnershipShareID: ownershipShare.ID,
				Amount:           record.Amount,
				PaymentDate:      record.Date,
				Notes:            &record.Description,
				ContractorID:     &contractor.ID,
				PaymentMethod:    entity.PaymentMethodBankTransfer,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := uc.paymentRepo.Create(ctx, payment); err != nil {
				return fmt.Errorf("failed to create payment: %w", err)
			}
		} else {
			// No ownership share found - create ContractorPayment (from business contractor)
			contractorPayment := &entity.ContractorPayment{
				ContractorID:  contractor.ID,
				Amount:        record.Amount,
				PaymentDate:   record.Date,
				PaymentMethod: entity.PaymentMethodBankTransfer,
				Purpose:       record.Description,
				Notes:         &record.Description,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := uc.contractorPaymentRepo.Create(ctx, contractorPayment); err != nil {
				return fmt.Errorf("failed to create contractor payment: %w", err)
			}
		}

	} else if record.OperationType == entity.CashFlowOperationCredit {
		// Create Expense (Outcome)

		// Find Category by Code
		category, err := uc.getCategoryByCode(ctx, record.CategoryCode)
		if err != nil {
			return fmt.Errorf("failed to find category %s: %w", record.CategoryCode, err)
		}

		expense := &entity.Expense{
			Amount:        record.Amount,
			ExpenseDate:   record.Date,
			Description:   record.Description,
			ContractorID:  &contractor.ID,
			CategoryID:    category.ID,
			PaymentStatus: entity.PaymentStatusPending,
			PaidAmount:    0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
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

// findOwnershipShareByName tries to find an ownership share by matching the owner's last name
func (uc *ProcessCashFlowImportBatchUseCase) findOwnershipShareByName(ctx context.Context, contractorName string) (*entity.OwnershipShare, error) {
	// Get all ownership shares
	shares, err := uc.ownershipRepo.List(ctx, repository.OwnershipShareFilter{
		Limit: 10000, // Get all shares
	})
	if err != nil {
		return nil, err
	}

	// Try to match by owner's last name (case-insensitive)
	contractorNameUpper := strings.ToUpper(strings.TrimSpace(contractorName))

	for _, share := range shares {
		owner, err := uc.ownerRepo.GetByID(ctx, share.OwnerID)
		if err != nil || owner == nil {
			continue
		}

		ownerLastNameUpper := strings.ToUpper(owner.LastName)

		// Match if contractor name contains or equals owner's last name
		if ownerLastNameUpper == contractorNameUpper || strings.Contains(contractorNameUpper, ownerLastNameUpper) {
			return share, nil
		}
	}

	return nil, fmt.Errorf("no ownership share found for contractor: %s", contractorName)
}
