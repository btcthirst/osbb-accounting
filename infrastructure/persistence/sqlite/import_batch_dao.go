// infrastructure/persistence/sqlite/import_batch_dao.go
package sqlite

import (
	"database/sql"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

// ImportBatchDAO implements ImportBatchRepository.
type ImportBatchDAO struct {
	db *sql.DB
}

// NewImportBatchDAO creates a new DAO.
func NewImportBatchDAO(db *sql.DB) repository.ImportBatchRepository {
	return &ImportBatchDAO{db: db}
}

// Create створює новий батч імпорту.
func (dao *ImportBatchDAO) Create(batch *entity.ImportBatch) (int64, error) {
	query := `
		INSERT INTO import_batches (
			file_name, file_path, status, total_sheets, total_records,
			successful_records, failed_records, error_message, metadata,
			imported_by, imported_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := dao.db.Exec(
		query,
		batch.FileName,
		batch.FilePath,
		batch.Status,
		batch.TotalSheets,
		batch.TotalRecords,
		batch.SuccessfulRecords,
		batch.FailedRecords,
		batch.ErrorMessage,
		batch.Metadata,
		batch.ImportedBy,
		batch.ImportedAt.Unix(),
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// FindByID знаходить батч за ID.
func (dao *ImportBatchDAO) FindByID(id int64) (*entity.ImportBatch, error) {
	query := `
		SELECT id, file_name, file_path, status, total_sheets, total_records,
		       successful_records, failed_records, error_message, metadata,
		       imported_by, imported_at, completed_at, created_at, updated_at
		FROM import_batches
		WHERE id = ?
	`

	var batch entity.ImportBatch
	var filePath, errorMessage, metadata sql.NullString
	var completedAt, importedAtUnix, createdAtUnix, updatedAtUnix sql.NullInt64

	err := dao.db.QueryRow(query, id).Scan(
		&batch.ID,
		&batch.FileName,
		&filePath,
		&batch.Status,
		&batch.TotalSheets,
		&batch.TotalRecords,
		&batch.SuccessfulRecords,
		&batch.FailedRecords,
		&errorMessage,
		&metadata,
		&batch.ImportedBy,
		&importedAtUnix,
		&completedAt,
		&createdAtUnix,
		&updatedAtUnix,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Convert NULLs
	if filePath.Valid {
		batch.FilePath = &filePath.String
	}
	if errorMessage.Valid {
		batch.ErrorMessage = &errorMessage.String
	}
	if metadata.Valid {
		batch.Metadata = &metadata.String
	}
	if completedAt.Valid {
		t := time.Unix(completedAt.Int64, 0)
		batch.CompletedAt = &t
	}
	if importedAtUnix.Valid {
		batch.ImportedAt = time.Unix(importedAtUnix.Int64, 0)
	}
	if createdAtUnix.Valid {
		batch.CreatedAt = time.Unix(createdAtUnix.Int64, 0)
	}
	if updatedAtUnix.Valid {
		batch.UpdatedAt = time.Unix(updatedAtUnix.Int64, 0)
	}

	return &batch, nil
}

// FindAll повертає всі батчі.
func (dao *ImportBatchDAO) FindAll() ([]*entity.ImportBatch, error) {
	query := `
		SELECT id, file_name, file_path, status, total_sheets, total_records,
		       successful_records, failed_records, error_message, metadata,
		       imported_by, imported_at, completed_at, created_at, updated_at
		FROM import_batches
		ORDER BY imported_at DESC
	`

	rows, err := dao.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dao.scanBatches(rows)
}

// FindByImporter знаходить батчі користувача.
func (dao *ImportBatchDAO) FindByImporter(importerID int64) ([]*entity.ImportBatch, error) {
	query := `
		SELECT id, file_name, file_path, status, total_sheets, total_records,
		       successful_records, failed_records, error_message, metadata,
		       imported_by, imported_at, completed_at, created_at, updated_at
		FROM import_batches
		WHERE imported_by = ?
		ORDER BY imported_at DESC
	`

	rows, err := dao.db.Query(query, importerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dao.scanBatches(rows)
}

// Update оновлює батч.
func (dao *ImportBatchDAO) Update(batch *entity.ImportBatch) error {
	query := `
		UPDATE import_batches
		SET file_name = ?, file_path = ?, status = ?, total_sheets = ?,
		    total_records = ?, successful_records = ?, failed_records = ?,
		    error_message = ?, metadata = ?, completed_at = ?
		WHERE id = ?
	`

	var completedAt sql.NullInt64
	if batch.CompletedAt != nil {
		completedAt = sql.NullInt64{Int64: batch.CompletedAt.Unix(), Valid: true}
	}

	_, err := dao.db.Exec(
		query,
		batch.FileName,
		batch.FilePath,
		batch.Status,
		batch.TotalSheets,
		batch.TotalRecords,
		batch.SuccessfulRecords,
		batch.FailedRecords,
		batch.ErrorMessage,
		batch.Metadata,
		completedAt,
		batch.ID,
	)

	return err
}

// Delete видаляє батч (каскадно видалить всі записи).
func (dao *ImportBatchDAO) Delete(id int64) error {
	query := `DELETE FROM import_batches WHERE id = ?`
	_, err := dao.db.Exec(query, id)
	return err
}

// UpdateStatus оновлює статус батчу.
func (dao *ImportBatchDAO) UpdateStatus(id int64, status entity.ImportStatus, errorMessage *string) error {
	query := `
		UPDATE import_batches
		SET status = ?, error_message = ?
		WHERE id = ?
	`

	_, err := dao.db.Exec(query, status, errorMessage, id)
	return err
}

// UpdateStats оновлює статистику батчу.
func (dao *ImportBatchDAO) UpdateStats(id int64, totalSheets, totalRecords, successfulRecords, failedRecords int) error {
	query := `
		UPDATE import_batches
		SET total_sheets = ?, total_records = ?, successful_records = ?, failed_records = ?
		WHERE id = ?
	`

	_, err := dao.db.Exec(query, totalSheets, totalRecords, successfulRecords, failedRecords, id)
	return err
}

// scanBatches helper function to scan multiple batches.
func (dao *ImportBatchDAO) scanBatches(rows *sql.Rows) ([]*entity.ImportBatch, error) {
	var batches []*entity.ImportBatch

	for rows.Next() {
		var batch entity.ImportBatch
		var filePath, errorMessage, metadata sql.NullString
		var completedAt, importedAtUnix, createdAtUnix, updatedAtUnix sql.NullInt64

		err := rows.Scan(
			&batch.ID,
			&batch.FileName,
			&filePath,
			&batch.Status,
			&batch.TotalSheets,
			&batch.TotalRecords,
			&batch.SuccessfulRecords,
			&batch.FailedRecords,
			&errorMessage,
			&metadata,
			&batch.ImportedBy,
			&importedAtUnix,
			&completedAt,
			&createdAtUnix,
			&updatedAtUnix,
		)

		if err != nil {
			return nil, err
		}

		// Convert NULLs
		if filePath.Valid {
			batch.FilePath = &filePath.String
		}
		if errorMessage.Valid {
			batch.ErrorMessage = &errorMessage.String
		}
		if metadata.Valid {
			batch.Metadata = &metadata.String
		}
		if completedAt.Valid {
			t := time.Unix(completedAt.Int64, 0)
			batch.CompletedAt = &t
		}
		if importedAtUnix.Valid {
			batch.ImportedAt = time.Unix(importedAtUnix.Int64, 0)
		}
		if createdAtUnix.Valid {
			batch.CreatedAt = time.Unix(createdAtUnix.Int64, 0)
		}
		if updatedAtUnix.Valid {
			batch.UpdatedAt = time.Unix(updatedAtUnix.Int64, 0)
		}

		batches = append(batches, &batch)
	}

	return batches, rows.Err()
}
