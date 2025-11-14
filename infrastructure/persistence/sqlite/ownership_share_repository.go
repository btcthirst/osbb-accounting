// infrastructure/persistence/sqlite/ownership_share_repository.go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/domain/repository"
)

// OwnershipShareRepository реалізує repository.OwnershipShareRepository для SQLite.
type OwnershipShareRepository struct {
	db *sql.DB
}

// NewOwnershipShareRepository створює новий OwnershipShareRepository.
func NewOwnershipShareRepository(db *sql.DB) *OwnershipShareRepository {
	return &OwnershipShareRepository{db: db}
}

// Create створює нову частку власності.
func (r *OwnershipShareRepository) Create(ctx context.Context, share *entity.OwnershipShare) error {
	// Перевірка дублювання активної частки
	hasDuplicate, err := r.CheckDuplicateActiveOwnership(ctx, share.OwnerID, share.ApartmentID, nil)
	if err != nil {
		return fmt.Errorf("failed to check duplicate: %w", err)
	}
	if hasDuplicate {
		return domainErrors.NewDomainError(
			domainErrors.CodeDuplicateEntry,
			"owner already has an active ownership share for this apartment",
			domainErrors.ErrAlreadyExists,
		)
	}

	// Перевірка що сума часток не перевищує 100%
	totalShare, err := r.CalculateTotalShareForApartment(ctx, share.ApartmentID, nil)
	if err != nil {
		return fmt.Errorf("failed to calculate total share: %w", err)
	}

	newShare := share.GetSharePercentage()
	if totalShare+newShare > 100.01 { // Невелика похибка для float
		return domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			fmt.Sprintf("total ownership share would exceed 100%% (current: %.1f%%, adding: %.1f%%)",
				totalShare, newShare),
			entity.ErrOwnershipShareExceedsTotal,
		)
	}

	query := `
		INSERT INTO ownership_shares (
			owner_id, apartment_id, share_numerator, share_denominator,
			ownership_type, start_date, end_date, document_type,
			document_number, document_date, notes, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		share.OwnerID,
		share.ApartmentID,
		share.ShareNumerator,
		share.ShareDenominator,
		string(share.OwnershipType),
		share.StartDate.Unix(),
		timeToNullInt64(share.EndDate),
		share.DocumentType,
		share.DocumentNumber,
		timeToNullInt64(share.DocumentDate),
		share.Notes,
		boolToInt(share.IsActive),
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create ownership share: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	share.ID = id
	share.CreatedAt = time.Unix(now, 0)
	share.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує частку за ID.
func (r *OwnershipShareRepository) GetByID(ctx context.Context, id int64) (*entity.OwnershipShare, error) {
	query := `
		SELECT 
			id, owner_id, apartment_id, share_numerator, share_denominator,
			ownership_type, start_date, end_date, document_type,
			document_number, document_date, notes, is_active,
			deleted_at, created_at, updated_at
		FROM ownership_shares
		WHERE id = ? AND deleted_at IS NULL
	`

	return r.scanShare(ctx, query, id)
}

// Update оновлює частку власності.
func (r *OwnershipShareRepository) Update(ctx context.Context, share *entity.OwnershipShare) error {
	// Перевірка що сума часток не перевищує 100%
	totalShare, err := r.CalculateTotalShareForApartment(ctx, share.ApartmentID, &share.ID)
	if err != nil {
		return fmt.Errorf("failed to calculate total share: %w", err)
	}

	newShare := share.GetSharePercentage()
	if totalShare+newShare > 100.01 {
		return domainErrors.NewDomainError(
			domainErrors.CodeValidationFailed,
			fmt.Sprintf("total ownership share would exceed 100%% (current: %.1f%%, adding: %.1f%%)",
				totalShare, newShare),
			entity.ErrOwnershipShareExceedsTotal,
		)
	}

	query := `
		UPDATE ownership_shares
		SET share_numerator = ?,
		    share_denominator = ?,
		    ownership_type = ?,
		    end_date = ?,
		    document_type = ?,
		    document_number = ?,
		    document_date = ?,
		    notes = ?,
		    is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		share.ShareNumerator,
		share.ShareDenominator,
		string(share.OwnershipType),
		timeToNullInt64(share.EndDate),
		share.DocumentType,
		share.DocumentNumber,
		timeToNullInt64(share.DocumentDate),
		share.Notes,
		boolToInt(share.IsActive),
		now,
		share.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ownership share: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	share.UpdatedAt = time.Unix(now, 0)

	return nil
}

// SoftDelete виконує м'яке видалення частки.
func (r *OwnershipShareRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE ownership_shares
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete ownership share: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}

// List отримує список часток з фільтрацією.
func (r *OwnershipShareRepository) List(ctx context.Context, filter repository.OwnershipShareFilter) ([]*entity.OwnershipShare, error) {
	query := `
		SELECT 
			id, owner_id, apartment_id, share_numerator, share_denominator,
			ownership_type, start_date, end_date, document_type,
			document_number, document_date, notes, is_active,
			deleted_at, created_at, updated_at
		FROM ownership_shares
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.OwnerID != nil {
		query += " AND owner_id = ?"
		args = append(args, *filter.OwnerID)
	}

	if filter.ApartmentID != nil {
		query += " AND apartment_id = ?"
		args = append(args, *filter.ApartmentID)
	}

	if filter.OwnershipType != nil {
		query += " AND ownership_type = ?"
		args = append(args, string(*filter.OwnershipType))
	}

	if filter.IsCurrentlyActive {
		now := time.Now().Unix()
		query += " AND is_active = 1 AND start_date <= ? AND (end_date IS NULL OR end_date > ?)"
		args = append(args, now, now)
	}

	// Сортування
	orderBy := "start_date DESC"
	if filter.OrderBy != "" {
		switch filter.OrderBy {
		case "start_date":
			orderBy = "start_date"
		case "owner":
			orderBy = "owner_id"
		case "apartment":
			orderBy = "apartment_id"
		}
		if filter.OrderDesc {
			orderBy += " DESC"
		}
	}
	query += " ORDER BY " + orderBy

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ownership shares: %w", err)
	}
	defer rows.Close()

	return r.scanShares(rows)
}

// Count отримує загальну кількість часток.
func (r *OwnershipShareRepository) Count(ctx context.Context, filter repository.OwnershipShareFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM ownership_shares WHERE 1=1`
	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.OwnerID != nil {
		query += " AND owner_id = ?"
		args = append(args, *filter.OwnerID)
	}

	if filter.ApartmentID != nil {
		query += " AND apartment_id = ?"
		args = append(args, *filter.ApartmentID)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ownership shares: %w", err)
	}

	return count, nil
}

// GetByOwnerID отримує всі частки власника.
func (r *OwnershipShareRepository) GetByOwnerID(ctx context.Context, ownerID int64) ([]*entity.OwnershipShare, error) {
	return r.List(ctx, repository.OwnershipShareFilter{
		OwnerID: &ownerID,
	})
}

// GetByApartmentID отримує всі частки квартири.
func (r *OwnershipShareRepository) GetByApartmentID(ctx context.Context, apartmentID int64) ([]*entity.OwnershipShare, error) {
	return r.List(ctx, repository.OwnershipShareFilter{
		ApartmentID: &apartmentID,
	})
}

// GetActiveByApartment отримує тільки активні частки квартири.
func (r *OwnershipShareRepository) GetActiveByApartment(ctx context.Context, apartmentID int64) ([]*entity.OwnershipShare, error) {
	isActive := true
	return r.List(ctx, repository.OwnershipShareFilter{
		ApartmentID:       &apartmentID,
		IsActive:          &isActive,
		IsCurrentlyActive: true,
	})
}

// CheckDuplicateActiveOwnership перевіряє наявність активної частки.
func (r *OwnershipShareRepository) CheckDuplicateActiveOwnership(ctx context.Context, ownerID, apartmentID int64, excludeID *int64) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM ownership_shares 
		WHERE owner_id = ? 
		  AND apartment_id = ? 
		  AND is_active = 1 
		  AND deleted_at IS NULL
	`
	args := []interface{}{ownerID, apartmentID}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return count > 0, nil
}

// CalculateTotalShareForApartment обчислює загальну суму часток у квартирі.
func (r *OwnershipShareRepository) CalculateTotalShareForApartment(ctx context.Context, apartmentID int64, excludeID *int64) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CAST(share_numerator AS REAL) / share_denominator), 0) * 100
		FROM ownership_shares
		WHERE apartment_id = ? 
		  AND is_active = 1 
		  AND deleted_at IS NULL
	`
	args := []interface{}{apartmentID}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	var totalShare float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&totalShare)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total share: %w", err)
	}

	return totalShare, nil
}

// GetWithDetails отримує частку з повною інформацією.
func (r *OwnershipShareRepository) GetWithDetails(ctx context.Context, id int64) (*repository.OwnershipShareDetails, error) {
	query := `
		SELECT 
			os.id, os.owner_id, os.apartment_id, os.share_numerator, os.share_denominator,
			os.ownership_type, os.start_date, os.end_date, os.document_type,
			os.document_number, os.document_date, os.notes, os.is_active,
			os.deleted_at, os.created_at, os.updated_at,
			o.last_name || ' ' || o.first_name || COALESCE(' ' || o.middle_name, '') as owner_name,
			o.phone as owner_phone,
			o.email as owner_email,
			a.apartment_number,
			a.floor,
			a.entrance
		FROM ownership_shares os
		INNER JOIN owners o ON os.owner_id = o.id
		INNER JOIN apartments a ON os.apartment_id = a.id
		WHERE os.id = ? AND os.deleted_at IS NULL
	`

	details := &repository.OwnershipShareDetails{
		Share: &entity.OwnershipShare{},
	}

	var ownershipType string
	var startDate, endDate, documentDate, deletedAt, createdAt, updatedAt sql.NullInt64
	var documentType, documentNumber, notes sql.NullString
	var ownerPhone, ownerEmail sql.NullString
	var apartmentEntrance sql.NullInt64
	var isActive int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&details.Share.ID,
		&details.Share.OwnerID,
		&details.Share.ApartmentID,
		&details.Share.ShareNumerator,
		&details.Share.ShareDenominator,
		&ownershipType,
		&startDate,
		&endDate,
		&documentType,
		&documentNumber,
		&documentDate,
		&notes,
		&isActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
		&details.OwnerName,
		&ownerPhone,
		&ownerEmail,
		&details.ApartmentNumber,
		&details.ApartmentFloor,
		&apartmentEntrance,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ownership share details: %w", err)
	}

	details.Share.OwnershipType = entity.OwnershipType(ownershipType)
	details.Share.StartDate = time.Unix(startDate.Int64, 0)
	details.Share.EndDate = nullInt64ToTimePtr(endDate)
	details.Share.DocumentType = nullStringToPtr(documentType)
	details.Share.DocumentNumber = nullStringToPtr(documentNumber)
	details.Share.DocumentDate = nullInt64ToTimePtr(documentDate)
	details.Share.Notes = nullStringToPtr(notes)
	details.Share.IsActive = intToBool(isActive)
	details.Share.DeletedAt = nullInt64ToTimePtr(deletedAt)
	details.Share.CreatedAt = time.Unix(createdAt.Int64, 0)
	details.Share.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	details.OwnerPhone = nullStringToPtr(ownerPhone)
	details.OwnerEmail = nullStringToPtr(ownerEmail)

	if apartmentEntrance.Valid {
		e := int(apartmentEntrance.Int64)
		details.ApartmentEntrance = &e
	}

	return details, nil
}

// ListWithDetails отримує список часток з повною інформацією.
func (r *OwnershipShareRepository) ListWithDetails(ctx context.Context, filter repository.OwnershipShareFilter) ([]*repository.OwnershipShareDetails, error) {
	query := `
		SELECT 
			os.id, os.owner_id, os.apartment_id, os.share_numerator, os.share_denominator,
			os.ownership_type, os.start_date, os.end_date, os.document_type,
			os.document_number, os.document_date, os.notes, os.is_active,
			os.deleted_at, os.created_at, os.updated_at,
			o.last_name || ' ' || o.first_name || COALESCE(' ' || o.middle_name, '') as owner_name,
			o.phone as owner_phone,
			o.email as owner_email,
			a.apartment_number,
			a.floor,
			a.entrance
		FROM ownership_shares os
		INNER JOIN owners o ON os.owner_id = o.id
		INNER JOIN apartments a ON os.apartment_id = a.id
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND os.deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND os.is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.OwnerID != nil {
		query += " AND os.owner_id = ?"
		args = append(args, *filter.OwnerID)
	}

	if filter.ApartmentID != nil {
		query += " AND os.apartment_id = ?"
		args = append(args, *filter.ApartmentID)
	}

	query += " ORDER BY os.start_date DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ownership shares with details: %w", err)
	}
	defer rows.Close()

	results := make([]*repository.OwnershipShareDetails, 0)

	for rows.Next() {
		details := &repository.OwnershipShareDetails{
			Share: &entity.OwnershipShare{},
		}

		var ownershipType string
		var startDate, endDate, documentDate, deletedAt, createdAt, updatedAt sql.NullInt64
		var documentType, documentNumber, notes sql.NullString
		var ownerPhone, ownerEmail sql.NullString
		var apartmentEntrance sql.NullInt64
		var isActive int

		err := rows.Scan(
			&details.Share.ID,
			&details.Share.OwnerID,
			&details.Share.ApartmentID,
			&details.Share.ShareNumerator,
			&details.Share.ShareDenominator,
			&ownershipType,
			&startDate,
			&endDate,
			&documentType,
			&documentNumber,
			&documentDate,
			&notes,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
			&details.OwnerName,
			&ownerPhone,
			&ownerEmail,
			&details.ApartmentNumber,
			&details.ApartmentFloor,
			&apartmentEntrance,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan ownership share details: %w", err)
		}

		details.Share.OwnershipType = entity.OwnershipType(ownershipType)
		details.Share.StartDate = time.Unix(startDate.Int64, 0)
		details.Share.EndDate = nullInt64ToTimePtr(endDate)
		details.Share.DocumentType = nullStringToPtr(documentType)
		details.Share.DocumentNumber = nullStringToPtr(documentNumber)
		details.Share.DocumentDate = nullInt64ToTimePtr(documentDate)
		details.Share.Notes = nullStringToPtr(notes)
		details.Share.IsActive = intToBool(isActive)
		details.Share.DeletedAt = nullInt64ToTimePtr(deletedAt)
		details.Share.CreatedAt = time.Unix(createdAt.Int64, 0)
		details.Share.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		details.OwnerPhone = nullStringToPtr(ownerPhone)
		details.OwnerEmail = nullStringToPtr(ownerEmail)

		if apartmentEntrance.Valid {
			e := int(apartmentEntrance.Int64)
			details.ApartmentEntrance = &e
		}

		results = append(results, details)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ownership shares: %w", err)
	}

	return results, nil
}

// Helper функції
func (r *OwnershipShareRepository) scanShare(ctx context.Context, query string, args ...interface{}) (*entity.OwnershipShare, error) {
	share := &entity.OwnershipShare{}
	var ownershipType string
	var startDate, endDate, documentDate, deletedAt, createdAt, updatedAt sql.NullInt64
	var documentType, documentNumber, notes sql.NullString
	var isActive int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&share.ID,
		&share.OwnerID,
		&share.ApartmentID,
		&share.ShareNumerator,
		&share.ShareDenominator,
		&ownershipType,
		&startDate,
		&endDate,
		&documentType,
		&documentNumber,
		&documentDate,
		&notes,
		&isActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan ownership share: %w", err)
	}

	share.OwnershipType = entity.OwnershipType(ownershipType)
	share.StartDate = time.Unix(startDate.Int64, 0)
	share.EndDate = nullInt64ToTimePtr(endDate)
	share.DocumentType = nullStringToPtr(documentType)
	share.DocumentNumber = nullStringToPtr(documentNumber)
	share.DocumentDate = nullInt64ToTimePtr(documentDate)
	share.Notes = nullStringToPtr(notes)
	share.IsActive = intToBool(isActive)
	share.DeletedAt = nullInt64ToTimePtr(deletedAt)
	share.CreatedAt = time.Unix(createdAt.Int64, 0)
	share.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return share, nil
}

func (r *OwnershipShareRepository) scanShares(rows *sql.Rows) ([]*entity.OwnershipShare, error) {
	shares := make([]*entity.OwnershipShare, 0)

	for rows.Next() {
		share := &entity.OwnershipShare{}
		var ownershipType string
		var startDate, endDate, documentDate, deletedAt, createdAt, updatedAt sql.NullInt64
		var documentType, documentNumber, notes sql.NullString
		var isActive int

		err := rows.Scan(
			&share.ID,
			&share.OwnerID,
			&share.ApartmentID,
			&share.ShareNumerator,
			&share.ShareDenominator,
			&ownershipType,
			&startDate,
			&endDate,
			&documentType,
			&documentNumber,
			&documentDate,
			&notes,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan ownership share: %w", err)
		}

		share.OwnershipType = entity.OwnershipType(ownershipType)
		share.StartDate = time.Unix(startDate.Int64, 0)
		share.EndDate = nullInt64ToTimePtr(endDate)
		share.DocumentType = nullStringToPtr(documentType)
		share.DocumentNumber = nullStringToPtr(documentNumber)
		share.DocumentDate = nullInt64ToTimePtr(documentDate)
		share.Notes = nullStringToPtr(notes)
		share.IsActive = intToBool(isActive)
		share.DeletedAt = nullInt64ToTimePtr(deletedAt)
		share.CreatedAt = time.Unix(createdAt.Int64, 0)
		share.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ownership shares: %w", err)
	}

	return shares, nil
}

func timeToNullInt64(t *time.Time) sql.NullInt64 {
	if t != nil {
		return sql.NullInt64{Int64: t.Unix(), Valid: true}
	}
	return sql.NullInt64{Valid: false}
}
