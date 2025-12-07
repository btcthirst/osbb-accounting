package importusecase

import (
	"context"
	"fmt"
	"time"

	"osbb-accounting/application/importer"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

type ImportCashFlowXLSXUseCase struct {
	parser       *importer.CashFlowXLSXParser
	batchRepo    repository.ImportBatchRepository
	cashFlowRepo repository.ImportedCashFlowRecordRepository
}

func NewImportCashFlowXLSXUseCase(
	batchRepo repository.ImportBatchRepository,
	cashFlowRepo repository.ImportedCashFlowRecordRepository,
) *ImportCashFlowXLSXUseCase {
	return &ImportCashFlowXLSXUseCase{
		parser:       importer.NewCashFlowXLSXParser(),
		batchRepo:    batchRepo,
		cashFlowRepo: cashFlowRepo,
	}
}

func (uc *ImportCashFlowXLSXUseCase) Execute(input ImportXLSXInput) (*ImportXLSXOutput, error) {
	startTime := time.Now()

	// 1. Create Batch
	batch, err := entity.NewImportBatch(input.FileName, &input.FilePath, input.ImportedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}
	// TODO: Add ImportType to ImportBatch entity?
	// Currently ImportBatch doesn't distinguish types.
	// We might need to add a Type field or just assume context.
	// For now, let's assume we use the same ImportBatch entity.

	batchID, err := uc.batchRepo.Create(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	// 2. Update Status
	err = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusProcessing, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update batch status: %w", err)
	}

	// 3. Parse File
	parsedData, err := uc.parser.ParseFile(input.FilePath, batchID)
	if err != nil {
		errMsg := err.Error()
		_ = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusFailed, &errMsg)
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// 4. Save Records
	// Repo needs BulkCreate? I only implemented Create.
	// I should update repo to support BulkCreate or loop.
	// Loop is fine for now as volume is low (hundreds).
	for _, record := range parsedData.Records {
		if err := uc.cashFlowRepo.Create(context.Background(), record); err != nil {
			// Log error but continue? Or fail?
			// Let's fail for now to be safe.
			errMsg := fmt.Sprintf("failed to save record: %v", err)
			_ = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusFailed, &errMsg)
			return nil, err
		}
	}

	// 5. Update Stats
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

	// 6. Complete
	err = uc.batchRepo.UpdateStatus(batchID, entity.ImportStatusCompleted, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to complete batch: %w", err)
	}

	return &ImportXLSXOutput{
		BatchID:           batchID,
		TotalSheets:       parsedData.TotalSheets,
		TotalRecords:      len(parsedData.Records),
		SuccessfulRecords: parsedData.SuccessCount,
		FailedRecords:     parsedData.FailedCount,
		Errors:            parsedData.Errors,
		Duration:          time.Since(startTime),
	}, nil
}
