// domain/repository/imported_record_repository.go
package repository

import (
	"osbb-accounting/domain/entity"
)

// ImportedRecordRepository визначає операції з імпортованими записами.
type ImportedRecordRepository interface {
	// Create створює новий запис.
	Create(record *entity.ImportedMonthlyRecord) (id int64, err error)

	// BulkCreate створює багато записів ефективно.
	BulkCreate(records []*entity.ImportedMonthlyRecord) error

	// FindByID знаходить запис за ID.
	FindByID(id int64) (*entity.ImportedMonthlyRecord, error)

	// FindByBatchID знаходить всі записи батчу.
	FindByBatchID(batchID int64) ([]*entity.ImportedMonthlyRecord, error)

	// FindByPeriod знаходить записи за період.
	FindByPeriod(month, year int) ([]*entity.ImportedMonthlyRecord, error)

	// FindByApartmentNumber знаходить записи за номером квартири.
	FindByApartmentNumber(apartmentNumber string) ([]*entity.ImportedMonthlyRecord, error)

	// FindUnmigrated знаходить незміґровані записи.
	FindUnmigrated(batchID *int64) ([]*entity.ImportedMonthlyRecord, error)

	// Update оновлює запис.
	Update(record *entity.ImportedMonthlyRecord) error

	// MarkAsMigrated позначає запис як міґрований.
	MarkAsMigrated(
		id int64,
		apartmentID, ownerID, ownershipShareID, chargeID, paymentID *int64,
	) error

	// Delete видаляє запис.
	Delete(id int64) error

	// DeleteByBatchID видаляє всі записи батчу.
	DeleteByBatchID(batchID int64) error

	// CountByBatchID рахує записи в батчі.
	CountByBatchID(batchID int64) (int, error)
}
