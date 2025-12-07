package repository

import (
	"context"
	"osbb-accounting/domain/entity"
)

// ImportedCashFlowRecordRepository defines the interface for interacting with imported cash flow records.
type ImportedCashFlowRecordRepository interface {
	// Create adds a new record to the database.
	Create(ctx context.Context, record *entity.ImportedCashFlowRecord) error

	// FindByBatchID retrieves all records associated with a specific import batch.
	FindByBatchID(ctx context.Context, batchID int64) ([]*entity.ImportedCashFlowRecord, error)

	// Update modifies an existing record.
	Update(ctx context.Context, record *entity.ImportedCashFlowRecord) error

	// DeleteByBatchID removes all records associated with a specific import batch.
	DeleteByBatchID(ctx context.Context, batchID int64) error
}
