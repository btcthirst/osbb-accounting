package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

type sqliteImportedCashFlowRecordRepository struct {
	db *sql.DB
}

func NewImportedCashFlowRecordRepository(db *sql.DB) repository.ImportedCashFlowRecordRepository {
	return &sqliteImportedCashFlowRecordRepository{db: db}
}

func (r *sqliteImportedCashFlowRecordRepository) Create(ctx context.Context, record *entity.ImportedCashFlowRecord) error {
	query := `
		INSERT INTO imported_cashflow_records (
			import_batch_id, date, contractor_name, operation_type, amount, category_code, description, is_migrated, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		record.ImportBatchID,
		record.Date,
		record.ContractorName,
		record.OperationType,
		record.Amount,
		record.CategoryCode,
		record.Description,
		record.IsMigrated,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert imported cashflow record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	record.ID = id
	return nil
}

func (r *sqliteImportedCashFlowRecordRepository) FindByBatchID(ctx context.Context, batchID int64) ([]*entity.ImportedCashFlowRecord, error) {
	query := `
		SELECT id, import_batch_id, date, contractor_name, operation_type, amount, category_code, description, is_migrated, created_at
		FROM imported_cashflow_records
		WHERE import_batch_id = ?
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query imported cashflow records: %w", err)
	}
	defer rows.Close()

	var records []*entity.ImportedCashFlowRecord
	for rows.Next() {
		var r entity.ImportedCashFlowRecord
		var categoryCode sql.NullString
		var description sql.NullString

		if err := rows.Scan(
			&r.ID,
			&r.ImportBatchID,
			&r.Date,
			&r.ContractorName,
			&r.OperationType,
			&r.Amount,
			&categoryCode,
			&description,
			&r.IsMigrated,
			&r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan imported cashflow record: %w", err)
		}

		if categoryCode.Valid {
			r.CategoryCode = categoryCode.String
		}
		if description.Valid {
			r.Description = description.String
		}

		records = append(records, &r)
	}

	return records, nil
}

func (r *sqliteImportedCashFlowRecordRepository) Update(ctx context.Context, record *entity.ImportedCashFlowRecord) error {
	query := `
		UPDATE imported_cashflow_records
		SET is_migrated = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, record.IsMigrated, record.ID)
	if err != nil {
		return fmt.Errorf("failed to update imported cashflow record: %w", err)
	}
	return nil
}

func (r *sqliteImportedCashFlowRecordRepository) DeleteByBatchID(ctx context.Context, batchID int64) error {
	query := `DELETE FROM imported_cashflow_records WHERE import_batch_id = ?`
	_, err := r.db.ExecContext(ctx, query, batchID)
	if err != nil {
		return fmt.Errorf("failed to delete imported cashflow records: %w", err)
	}
	return nil
}
