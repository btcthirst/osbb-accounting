// domain/entity/import_batch.go
package entity

import (
	"errors"
	"time"
)

// ImportBatch представляє операцію імпорту даних з файлу.
type ImportBatch struct {
	ID       int64
	FileName string
	FilePath *string

	// Статус
	Status ImportStatus

	// Статистика
	TotalSheets       int
	TotalRecords      int
	SuccessfulRecords int
	FailedRecords     int

	// Помилки
	ErrorMessage *string

	// Метадані (JSON)
	Metadata *string

	// Аудит
	ImportedBy  int64
	ImportedAt  time.Time
	CompletedAt *time.Time

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ImportStatus представляє статус імпорту.
type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

// Validation errors
var (
	ErrImportBatchFileNameRequired = errors.New("file name is required")
	ErrImportBatchImporterRequired = errors.New("importer user ID is required")
	ErrImportBatchInvalidStatus    = errors.New("invalid import status")
)

// NewImportBatch створює новий батч імпорту з валідацією.
func NewImportBatch(
	fileName string,
	filePath *string,
	importedBy int64,
) (*ImportBatch, error) {
	batch := &ImportBatch{
		FileName:   fileName,
		FilePath:   filePath,
		Status:     ImportStatusPending,
		ImportedBy: importedBy,
		ImportedAt: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := batch.Validate(); err != nil {
		return nil, err
	}

	return batch, nil
}

// Validate перевіряє коректність даних батчу.
func (b *ImportBatch) Validate() error {
	if b.FileName == "" {
		return ErrImportBatchFileNameRequired
	}

	if b.ImportedBy <= 0 {
		return ErrImportBatchImporterRequired
	}

	if !b.Status.IsValid() {
		return ErrImportBatchInvalidStatus
	}

	return nil
}

// IsValid перевіряє чи валідний статус імпорту.
func (s ImportStatus) IsValid() bool {
	switch s {
	case ImportStatusPending, ImportStatusProcessing, ImportStatusCompleted, ImportStatusFailed:
		return true
	default:
		return false
	}
}

// StartProcessing позначає батч як опрацьовуваний.
func (b *ImportBatch) StartProcessing() {
	b.Status = ImportStatusProcessing
	b.UpdatedAt = time.Now()
}

// Complete позначає батч як завершений успішно.
func (b *ImportBatch) Complete() {
	b.Status = ImportStatusCompleted
	now := time.Now()
	b.CompletedAt = &now
	b.UpdatedAt = now
}

// Fail позначає батч як невдалий.
func (b *ImportBatch) Fail(errorMessage string) {
	b.Status = ImportStatusFailed
	b.ErrorMessage = &errorMessage
	now := time.Now()
	b.CompletedAt = &now
	b.UpdatedAt = now
}

// UpdateStats оновлює статистику імпорту.
func (b *ImportBatch) UpdateStats(totalSheets, totalRecords, successfulRecords, failedRecords int) {
	b.TotalSheets = totalSheets
	b.TotalRecords = totalRecords
	b.SuccessfulRecords = successfulRecords
	b.FailedRecords = failedRecords
	b.UpdatedAt = time.Now()
}

// IsCompleted перевіряє чи завершено імпорт.
func (b *ImportBatch) IsCompleted() bool {
	return b.Status == ImportStatusCompleted
}

// IsFailed перевіряє чи невдалий імпорт.
func (b *ImportBatch) IsFailed() bool {
	return b.Status == ImportStatusFailed
}

// GetStatusDisplay повертає читабельну назву статусу.
func (s ImportStatus) GetDisplayName() string {
	switch s {
	case ImportStatusPending:
		return "Очікує"
	case ImportStatusProcessing:
		return "Обробляється"
	case ImportStatusCompleted:
		return "Завершено"
	case ImportStatusFailed:
		return "Помилка"
	default:
		return "Невідомо"
	}
}
