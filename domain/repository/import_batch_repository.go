// domain/repository/import_batch_repository.go
package repository

import (
	"osbb-accounting/domain/entity"
)

// ImportBatchRepository визначає операції з батчами імпорту.
type ImportBatchRepository interface {
	// Create створює новий батч імпорту.
	Create(batch *entity.ImportBatch) (id int64, err error)

	// FindByID знаходить батч за ID.
	FindByID(id int64) (*entity.ImportBatch, error)

	// FindAll повертає всі батчі.
	FindAll() ([]*entity.ImportBatch, error)

	// FindByImporter знаходить батчі користувача.
	FindByImporter(importerID int64) ([]*entity.ImportBatch, error)

	// Update оновлює батч.
	Update(batch *entity.ImportBatch) error

	// Delete видаляє батч (каскадно видалить всі записи).
	Delete(id int64) error

	// UpdateStatus оновлює статус батчу.
	UpdateStatus(id int64, status entity.ImportStatus, errorMessage *string) error

	// UpdateStats оновлює статистику батчу.
	UpdateStats(id int64, totalSheets, totalRecords, successfulRecords, failedRecords int) error
}
