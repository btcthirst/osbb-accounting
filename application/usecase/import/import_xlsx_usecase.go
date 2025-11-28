// application/usecase/import/import_xlsx_usecase.go
package importusecase

import (
	"fmt"
	"time"

	"osbb-accounting/application/importer"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

// ImportXLSXUseCase координує процес імпорту з XLSX.
type ImportXLSXUseCase struct {
	parser     *importer.XLSXParser
	batchRepo  repository.ImportBatchRepository
	recordRepo repository.ImportedRecordRepository
}

// ImportXLSXInput вхідні дані для імпорту.
type ImportXLSXInput struct {
	FilePath   string
	FileName   string
	ImportedBy int64
}

// ImportXLSXOutput результат імпорту.
type ImportXLSXOutput struct {
	BatchID           int64
	TotalSheets       int
	TotalRecords      int
	SuccessfulRecords int
	FailedRecords     int
	Errors            []string
	Duration          time.Duration
}

// NewImportXLSXUseCase створює новий use case.
func NewImportXLSXUseCase(
	batchRepo repository.ImportBatchRepository,
	recordRepo repository.ImportedRecordRepository,
) *ImportXLSXUseCase {
	return &ImportXLSXUseCase{
		parser:     importer.NewXLSXParser(),
		batchRepo:  batchRepo,
		recordRepo: recordRepo,
	}
}

// Execute виконує імпорт.
func (uc *ImportXLSXUseCase) Execute(input ImportXLSXInput) (*ImportXLSXOutput, error) {
	startTime := time.Now()

	// 1. Створити батч
	batch, err := entity.NewImportBatch(input.FileName, &input.FilePath, input.ImportedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	batchID, err := uc.batchRepo.Create(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	// 2. Позначити як в обробці
	err = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusProcessing, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update batch status: %w", err)
	}

	// 3. Парсити файл
	parsedData, err := uc.parser.ParseFile(input.FilePath, batchID)
	if err != nil {
		// Позначити як невдалий
		errMsg := err.Error()
		_ = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusFailed, &errMsg)
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// 4. Зберегти записи
	if len(parsedData.Records) > 0 {
		err = uc.recordRepo.BulkCreate(parsedData.Records)
		if err != nil {
			// Позначити як невдалий
			errMsg := fmt.Sprintf("failed to save records: %v", err)
			_ = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusFailed, &errMsg)
			return nil, fmt.Errorf("failed to save records: %w", err)
		}
	}

	// 5. Оновити статистику
	err = uc.batchRepo.UpdateStats(
		batchID,
		parsedData.TotalSheets,
		len(parsedData.Records),
		parsedData.SuccessCount,
		parsedData.FailedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update stats: %w", err)
	}

	// 6. Позначити як завершений
	err = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusCompleted, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete batch: %w", err)
	}

	duration := time.Since(startTime)

	return &ImportXLSXOutput{
		BatchID:           batchID,
		TotalSheets:       parsedData.TotalSheets,
		TotalRecords:      len(parsedData.Records),
		SuccessfulRecords: parsedData.SuccessCount,
		FailedRecords:     parsedData.FailedCount,
		Errors:            parsedData.Errors,
		Duration:          duration,
	}, nil
}
