// infrastructure/persistence/sqlite/osbb_repository.go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
)

// OSBBRepository реалізує repository.OSBBRepository для SQLite.
type OSBBRepository struct {
	db *sql.DB
}

// NewOSBBRepository створює новий OSBBRepository.
func NewOSBBRepository(db *sql.DB) *OSBBRepository {
	return &OSBBRepository{db: db}
}

// Get отримує дані ОСББ (завжди ID = 1).
func (r *OSBBRepository) Get(ctx context.Context) (*entity.OSBB, error) {
	query := `
		SELECT 
			id, name, edrpou, legal_address, actual_address,
			phone, email, website, chairman_name,
			created_at, updated_at
		FROM osbb
		WHERE id = 1
	`

	osbb := &entity.OSBB{}
	var actualAddress, phone, email, website sql.NullString
	var createdAt, updatedAt int64

	err := r.db.QueryRowContext(ctx, query).Scan(
		&osbb.ID,
		&osbb.Name,
		&osbb.EDRPOU,
		&osbb.LegalAddress,
		&actualAddress,
		&phone,
		&email,
		&website,
		&osbb.ChairmanName,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get OSBB: %w", err)
	}

	// Конвертація nullable полів
	osbb.ActualAddress = nullStringToPtr(actualAddress)
	osbb.Phone = nullStringToPtr(phone)
	osbb.Email = nullStringToPtr(email)
	osbb.Website = nullStringToPtr(website)
	osbb.CreatedAt = time.Unix(createdAt, 0)
	osbb.UpdatedAt = time.Unix(updatedAt, 0)

	return osbb, nil
}

// Create створює організацію ОСББ (може бути викликано тільки один раз).
func (r *OSBBRepository) Create(ctx context.Context, osbb *entity.OSBB) error {
	// Перевіряємо чи вже існує запис
	exists, err := r.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check OSBB existence: %w", err)
	}
	if exists {
		return domainErrors.NewDomainError(
			domainErrors.CodeAlreadyExists,
			"OSBB organization already exists",
			domainErrors.ErrAlreadyExists,
		)
	}

	query := `
		INSERT INTO osbb (
			id, name, edrpou, legal_address, actual_address,
			phone, email, website, chairman_name,
			created_at, updated_at
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = r.db.ExecContext(ctx, query,
		osbb.Name,
		osbb.EDRPOU,
		osbb.LegalAddress,
		osbb.ActualAddress,
		osbb.Phone,
		osbb.Email,
		osbb.Website,
		osbb.ChairmanName,
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "edrpou") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"EDRPOU already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "edrpou")
			}
		}
		return fmt.Errorf("failed to create OSBB: %w", err)
	}

	osbb.ID = 1
	osbb.CreatedAt = time.Unix(now, 0)
	osbb.UpdatedAt = time.Unix(now, 0)

	return nil
}

// Update оновлює дані ОСББ.
func (r *OSBBRepository) Update(ctx context.Context, osbb *entity.OSBB) error {
	query := `
		UPDATE osbb
		SET name = ?,
		    legal_address = ?,
		    actual_address = ?,
		    phone = ?,
		    email = ?,
		    website = ?,
		    chairman_name = ?,
		    updated_at = ?
		WHERE id = 1
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		osbb.Name,
		osbb.LegalAddress,
		osbb.ActualAddress,
		osbb.Phone,
		osbb.Email,
		osbb.Website,
		osbb.ChairmanName,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to update OSBB: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	osbb.UpdatedAt = time.Unix(now, 0)

	return nil
}

// Exists перевіряє чи існує запис ОСББ.
func (r *OSBBRepository) Exists(ctx context.Context) (bool, error) {
	query := `SELECT COUNT(*) FROM osbb WHERE id = 1`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check OSBB existence: %w", err)
	}

	return count > 0, nil
}

// GetByEDRPOU отримує ОСББ за ЄДРПОУ.
func (r *OSBBRepository) GetByEDRPOU(ctx context.Context, edrpou string) (*entity.OSBB, error) {
	query := `
		SELECT 
			id, name, edrpou, legal_address, actual_address,
			phone, email, website, chairman_name,
			created_at, updated_at
		FROM osbb
		WHERE edrpou = ?
	`

	osbb := &entity.OSBB{}
	var actualAddress, phone, email, website sql.NullString
	var createdAt, updatedAt int64

	err := r.db.QueryRowContext(ctx, query, edrpou).Scan(
		&osbb.ID,
		&osbb.Name,
		&osbb.EDRPOU,
		&osbb.LegalAddress,
		&actualAddress,
		&phone,
		&email,
		&website,
		&osbb.ChairmanName,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get OSBB by EDRPOU: %w", err)
	}

	osbb.ActualAddress = nullStringToPtr(actualAddress)
	osbb.Phone = nullStringToPtr(phone)
	osbb.Email = nullStringToPtr(email)
	osbb.Website = nullStringToPtr(website)
	osbb.CreatedAt = time.Unix(createdAt, 0)
	osbb.UpdatedAt = time.Unix(updatedAt, 0)

	return osbb, nil
}
