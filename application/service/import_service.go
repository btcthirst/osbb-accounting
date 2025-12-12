// application/service/import_service.go
package service

import (
	"context"
	"fmt"
	"path/filepath"

	importusecase "osbb-accounting/application/usecase/import"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

// ImportServiceInterface defines the contract for import operations.
type ImportServiceInterface interface {
	ImportXLSXFile(ctx context.Context, filePath string, importedBy int64) (*ImportSummary, error)
	ImportCashFlowFile(ctx context.Context, filePath string, importedBy int64) (*ImportSummary, error)
	GetImportBatches(ctx context.Context) ([]*entity.ImportBatch, error)
	GetImportBatchByID(ctx context.Context, id int64) (*entity.ImportBatch, error)
	GetImportedRecords(ctx context.Context, batchID int64) ([]*entity.ImportedMonthlyRecord, error)
	GetImportedCashFlowRecords(ctx context.Context, batchID int64) ([]*entity.ImportedCashFlowRecord, error)
	GetRecordsByPeriod(ctx context.Context, month, year int) ([]*entity.ImportedMonthlyRecord, error)
	DeleteImportBatch(ctx context.Context, id int64) error
	ProcessBatch(ctx context.Context, batchID int64) (*importusecase.ProcessResult, error)
}

// ImportSummary represents the summary of an import operation.
type ImportSummary struct {
	BatchID           int64
	FileName          string
	TotalSheets       int
	TotalRecords      int
	SuccessfulRecords int
	FailedRecords     int
	Errors            []string
	DurationSeconds   float64
}

// ImportService implements ImportServiceInterface.
type ImportService struct {
	importXLSXUseCase         *importusecase.ImportXLSXUseCase
	importCashFlowXLSXUseCase *importusecase.ImportCashFlowXLSXUseCase
	processUseCase            *importusecase.ProcessImportBatchUseCase
	processCashFlowUseCase    *importusecase.ProcessCashFlowImportBatchUseCase
	batchRepo                 repository.ImportBatchRepository
	recordRepo                repository.ImportedRecordRepository
	cashFlowRepo              repository.ImportedCashFlowRecordRepository
}

// NewImportService creates a new ImportService.
func NewImportService(
	batchRepo repository.ImportBatchRepository,
	recordRepo repository.ImportedRecordRepository,
	cashFlowRepo repository.ImportedCashFlowRecordRepository,
	apartmentRepo repository.ApartmentRepository,
	ownerRepo repository.OwnerRepository,
	ownershipShareRepo repository.OwnershipShareRepository,
	chargeRepo repository.ChargeRepository,
	paymentRepo repository.PaymentRepository,
	expenseRepo repository.ExpenseRepository,
	contractorPaymentRepo repository.ContractorPaymentRepository,
	contractorRepo repository.ContractorRepository,
	expenseCategoryRepo repository.ExpenseCategoryRepository,
) *ImportService {
	return &ImportService{
		importXLSXUseCase:         importusecase.NewImportXLSXUseCase(batchRepo, recordRepo),
		importCashFlowXLSXUseCase: importusecase.NewImportCashFlowXLSXUseCase(batchRepo, cashFlowRepo),
		processUseCase: importusecase.NewProcessImportBatchUseCase(
			batchRepo,
			recordRepo,
			apartmentRepo,
			ownerRepo,
			ownershipShareRepo,
			chargeRepo,
			paymentRepo,
		),
		processCashFlowUseCase: importusecase.NewProcessCashFlowImportBatchUseCase(
			batchRepo,
			cashFlowRepo,
			paymentRepo,
			contractorPaymentRepo,
			expenseRepo,
			contractorRepo,
			expenseCategoryRepo,
			ownershipShareRepo,
			ownerRepo,
		),
		batchRepo:    batchRepo,
		recordRepo:   recordRepo,
		cashFlowRepo: cashFlowRepo,
	}
}

// ImportXLSXFile imports an XLSX file.
func (s *ImportService) ImportXLSXFile(ctx context.Context, filePath string, importedBy int64) (*ImportSummary, error) {
	fileName := filepath.Base(filePath)

	input := importusecase.ImportXLSXInput{
		FilePath:   filePath,
		FileName:   fileName,
		ImportedBy: importedBy,
	}

	result, err := s.importXLSXUseCase.Execute(input)
	if err != nil {
		return nil, fmt.Errorf("import failed: %w", err)
	}

	summary := &ImportSummary{
		BatchID:           result.BatchID,
		FileName:          fileName,
		TotalSheets:       result.TotalSheets,
		TotalRecords:      result.TotalRecords,
		SuccessfulRecords: result.SuccessfulRecords,
		FailedRecords:     result.FailedRecords,
		Errors:            result.Errors,
		DurationSeconds:   result.Duration.Seconds(),
	}

	return summary, nil
}

// ImportCashFlowFile imports a Cash Flow XLSX file.
func (s *ImportService) ImportCashFlowFile(ctx context.Context, filePath string, importedBy int64) (*ImportSummary, error) {
	fileName := filepath.Base(filePath)

	input := importusecase.ImportXLSXInput{
		FilePath:   filePath,
		FileName:   fileName,
		ImportedBy: importedBy,
	}

	result, err := s.importCashFlowXLSXUseCase.Execute(input)
	if err != nil {
		return nil, fmt.Errorf("cash flow import failed: %w", err)
	}

	summary := &ImportSummary{
		BatchID:           result.BatchID,
		FileName:          fileName,
		TotalSheets:       result.TotalSheets,
		TotalRecords:      result.TotalRecords,
		SuccessfulRecords: result.SuccessfulRecords,
		FailedRecords:     result.FailedRecords,
		Errors:            result.Errors,
		DurationSeconds:   result.Duration.Seconds(),
	}

	return summary, nil
}

// ProcessBatch processes an imported batch.
func (s *ImportService) ProcessBatch(ctx context.Context, batchID int64) (*importusecase.ProcessResult, error) {
	// Determine batch type?
	// Currently we don't store type.
	// But we can check if it has records in imported_records or imported_cashflow_records.

	// Try standard records first
	records, _ := s.recordRepo.FindByBatchID(batchID)
	if len(records) > 0 {
		return s.processUseCase.Execute(ctx, batchID)
	}

	// Try cash flow records
	cfRecords, _ := s.cashFlowRepo.FindByBatchID(ctx, batchID)
	if len(cfRecords) > 0 {
		return s.processCashFlowUseCase.Execute(ctx, batchID)
	}

	return nil, fmt.Errorf("unknown batch type or empty batch")
}

// GetImportBatches returns all import batches.
func (s *ImportService) GetImportBatches(ctx context.Context) ([]*entity.ImportBatch, error) {
	return s.batchRepo.FindAll()
}

// GetImportBatchByID returns a batch by ID.
func (s *ImportService) GetImportBatchByID(ctx context.Context, id int64) (*entity.ImportBatch, error) {
	return s.batchRepo.FindByID(id)
}

// GetImportedRecords returns records for a batch.
func (s *ImportService) GetImportedRecords(ctx context.Context, batchID int64) ([]*entity.ImportedMonthlyRecord, error) {
	return s.recordRepo.FindByBatchID(batchID)
}

// GetImportedCashFlowRecords returns cash flow records for a batch.
func (s *ImportService) GetImportedCashFlowRecords(ctx context.Context, batchID int64) ([]*entity.ImportedCashFlowRecord, error) {
	return s.cashFlowRepo.FindByBatchID(ctx, batchID)
}

// GetRecordsByPeriod returns records for a specific period.
func (s *ImportService) GetRecordsByPeriod(ctx context.Context, month, year int) ([]*entity.ImportedMonthlyRecord, error) {
	return s.recordRepo.FindByPeriod(month, year)
}

// DeleteImportBatch deletes a batch and its records.
func (s *ImportService) DeleteImportBatch(ctx context.Context, id int64) error {
	// Cascade delete should handle records if configured in DB.
	// But repositories might need explicit delete calls if not cascading.
	// Our migration has ON DELETE CASCADE.
	return s.batchRepo.Delete(id)
}
