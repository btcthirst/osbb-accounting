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

// ImportService надає високорівневі операції імпорту.
type ImportServiceInterface interface {
	ImportXLSXFile(ctx context.Context, filePath string, importedBy int64) (*ImportSummary, error)
	GetImportBatches(ctx context.Context) ([]*entity.ImportBatch, error)
	GetImportBatchByID(ctx context.Context, id int64) (*entity.ImportBatch, error)
	GetImportedRecords(ctx context.Context, batchID int64) ([]*entity.ImportedMonthlyRecord, error)
	GetRecordsByPeriod(ctx context.Context, month, year int) ([]*entity.ImportedMonthlyRecord, error)
	DeleteImportBatch(ctx context.Context, id int64) error
}

// ImportSummary представляє підсумок імпорту.
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

// importServiceImpl реалізація ImportService.
type importServiceImpl struct {
	importUseCase *importusecase.ImportXLSXUseCase
	batchRepo     repository.ImportBatchRepository
	recordRepo    repository.ImportedRecordRepository
}

// NewImportService створює новий сервіс.
func NewImportService(
	batchRepo repository.ImportBatchRepository,
	recordRepo repository.ImportedRecordRepository,
) ImportServiceInterface {
	importUC := importusecase.NewImportXLSXUseCase(batchRepo, recordRepo)

	return &importServiceImpl{
		importUseCase: importUC,
		batchRepo:     batchRepo,
		recordRepo:    recordRepo,
	}
}

// ImportXLSXFile імпортує XLSX файл.
func (s *importServiceImpl) ImportXLSXFile(ctx context.Context, filePath string, importedBy int64) (*ImportSummary, error) {
	fileName := filepath.Base(filePath)

	input := importusecase.ImportXLSXInput{
		FilePath:   filePath,
		FileName:   fileName,
		ImportedBy: importedBy,
	}

	result, err := s.importUseCase.Execute(input)
	if err != nil {
		return nil, fmt.Errorf("import failed: %w", err)
	}

	return &ImportSummary{
		BatchID:           result.BatchID,
		FileName:          fileName,
		TotalSheets:       result.TotalSheets,
		TotalRecords:      result.TotalRecords,
		SuccessfulRecords: result.SuccessfulRecords,
		FailedRecords:     result.FailedRecords,
		Errors:            result.Errors,
		DurationSeconds:   result.Duration.Seconds(),
	}, nil
}

// GetImportBatches повертає всі батчі.
func (s *importServiceImpl) GetImportBatches(ctx context.Context) ([]*entity.ImportBatch, error) {
	return s.batchRepo.FindAll()
}

// GetImportBatchByID повертає батч за ID.
func (s *importServiceImpl) GetImportBatchByID(ctx context.Context, id int64) (*entity.ImportBatch, error) {
	return s.batchRepo.FindByID(id)
}

// GetImportedRecords повертає записи батчу.
func (s *importServiceImpl) GetImportedRecords(ctx context.Context, batchID int64) ([]*entity.ImportedMonthlyRecord, error) {
	return s.recordRepo.FindByBatchID(batchID)
}

// GetRecordsByPeriod повертає записи за період.
func (s *importServiceImpl) GetRecordsByPeriod(ctx context.Context, month, year int) ([]*entity.ImportedMonthlyRecord, error) {
	return s.recordRepo.FindByPeriod(month, year)
}

// DeleteImportBatch видаляє батч (каскадно видалить всі записи).
func (s *importServiceImpl) DeleteImportBatch(ctx context.Context, id int64) error {
	return s.batchRepo.Delete(id)
}
