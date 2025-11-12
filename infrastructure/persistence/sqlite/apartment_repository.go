// infrastructure/persistence/sqlite/apartment_repository.go
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
	"osbb-accounting/domain/repository"
)

// ApartmentRepository реалізує repository.ApartmentRepository для SQLite.
type ApartmentRepository struct {
	db *sql.DB
}

// NewApartmentRepository створює новий ApartmentRepository.
func NewApartmentRepository(db *sql.DB) *ApartmentRepository {
	return &ApartmentRepository{db: db}
}

// Create створює нову квартиру.
func (r *ApartmentRepository) Create(ctx context.Context, apartment *entity.Apartment) error {
	query := `
		INSERT INTO apartments (
			apartment_number, floor, entrance, area_total, area_living,
			rooms_count, cadastral_number, notes, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		apartment.ApartmentNumber,
		apartment.Floor,
		apartment.Entrance,
		apartment.AreaTotal,
		apartment.AreaLiving,
		apartment.RoomsCount,
		apartment.CadastralNumber,
		apartment.Notes,
		boolToInt(apartment.IsActive),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domainErrors.NewDomainError(
				domainErrors.CodeDuplicateEntry,
				"apartment with this number already exists",
				domainErrors.ErrAlreadyExists,
			).WithDetails("field", "apartment_number")
		}
		return fmt.Errorf("failed to create apartment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	apartment.ID = id
	apartment.CreatedAt = time.Unix(now, 0)
	apartment.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує квартиру за ID.
func (r *ApartmentRepository) GetByID(ctx context.Context, id int64) (*entity.Apartment, error) {
	query := `
		SELECT 
			id, apartment_number, floor, entrance, area_total, area_living,
			rooms_count, cadastral_number, notes, is_active,
			deleted_at, created_at, updated_at
		FROM apartments
		WHERE id = ? AND deleted_at IS NULL
	`

	return r.scanApartment(ctx, query, id)
}

// GetByNumber отримує квартиру за номером.
func (r *ApartmentRepository) GetByNumber(ctx context.Context, number string) (*entity.Apartment, error) {
	query := `
		SELECT 
			id, apartment_number, floor, entrance, area_total, area_living,
			rooms_count, cadastral_number, notes, is_active,
			deleted_at, created_at, updated_at
		FROM apartments
		WHERE apartment_number = ? AND deleted_at IS NULL
	`

	return r.scanApartment(ctx, query, number)
}

// Update оновлює дані квартири.
func (r *ApartmentRepository) Update(ctx context.Context, apartment *entity.Apartment) error {
	query := `
		UPDATE apartments
		SET floor = ?,
		    entrance = ?,
		    area_total = ?,
		    area_living = ?,
		    rooms_count = ?,
		    cadastral_number = ?,
		    notes = ?,
		    is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		apartment.Floor,
		apartment.Entrance,
		apartment.AreaTotal,
		apartment.AreaLiving,
		apartment.RoomsCount,
		apartment.CadastralNumber,
		apartment.Notes,
		boolToInt(apartment.IsActive),
		now,
		apartment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update apartment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	apartment.UpdatedAt = time.Unix(now, 0)

	return nil
}

// SoftDelete виконує м'яке видалення квартири.
func (r *ApartmentRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE apartments
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete apartment: %w", err)
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

// Restore відновлює видалену квартиру.
func (r *ApartmentRepository) Restore(ctx context.Context, id int64) error {
	query := `
		UPDATE apartments
		SET deleted_at = NULL, is_active = 1, updated_at = ?
		WHERE id = ? AND deleted_at IS NOT NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore apartment: %w", err)
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

// List отримує список квартир з фільтрацією.
func (r *ApartmentRepository) List(ctx context.Context, filter repository.ApartmentFilter) ([]*entity.Apartment, error) {
	query := `
		SELECT 
			id, apartment_number, floor, entrance, area_total, area_living,
			rooms_count, cadastral_number, notes, is_active,
			deleted_at, created_at, updated_at
		FROM apartments
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

	if filter.SearchQuery != "" {
		query += " AND apartment_number LIKE ?"
		args = append(args, "%"+filter.SearchQuery+"%")
	}

	if filter.Floor != nil {
		query += " AND floor = ?"
		args = append(args, *filter.Floor)
	}

	if filter.Entrance != nil {
		query += " AND entrance = ?"
		args = append(args, *filter.Entrance)
	}

	if filter.MinArea != nil {
		query += " AND area_total >= ?"
		args = append(args, *filter.MinArea)
	}

	if filter.MaxArea != nil {
		query += " AND area_total <= ?"
		args = append(args, *filter.MaxArea)
	}

	if filter.HasOwners != nil {
		if *filter.HasOwners {
			query += ` AND EXISTS (
				SELECT 1 FROM ownership_shares 
				WHERE apartment_id = apartments.id 
				AND is_active = 1 
				AND deleted_at IS NULL
			)`
		} else {
			query += ` AND NOT EXISTS (
				SELECT 1 FROM ownership_shares 
				WHERE apartment_id = apartments.id 
				AND is_active = 1 
				AND deleted_at IS NULL
			)`
		}
	}

	// Сортування
	orderBy := "apartment_number"
	if filter.OrderBy != "" {
		switch filter.OrderBy {
		case "number":
			orderBy = "CAST(apartment_number AS INTEGER)"
		case "floor":
			orderBy = "floor, apartment_number"
		case "area":
			orderBy = "area_total"
		case "created_at":
			orderBy = "created_at"
		}
	}

	if filter.OrderDesc {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderBy)
	} else {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderBy)
	}

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
		return nil, fmt.Errorf("failed to list apartments: %w", err)
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// Count отримує загальну кількість квартир.
func (r *ApartmentRepository) Count(ctx context.Context, filter repository.ApartmentFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM apartments WHERE 1=1`
	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.SearchQuery != "" {
		query += " AND apartment_number LIKE ?"
		args = append(args, "%"+filter.SearchQuery+"%")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count apartments: %w", err)
	}

	return count, nil
}

// ExistsByNumber перевіряє чи існує квартира з таким номером.
func (r *ApartmentRepository) ExistsByNumber(ctx context.Context, number string) (bool, error) {
	query := `SELECT COUNT(*) FROM apartments WHERE apartment_number = ? AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, number).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check apartment number: %w", err)
	}

	return count > 0, nil
}

// GetByFloor отримує всі квартири на поверсі.
func (r *ApartmentRepository) GetByFloor(ctx context.Context, floor int) ([]*entity.Apartment, error) {
	query := `
		SELECT 
			id, apartment_number, floor, entrance, area_total, area_living,
			rooms_count, cadastral_number, notes, is_active,
			deleted_at, created_at, updated_at
		FROM apartments
		WHERE floor = ? AND deleted_at IS NULL
		ORDER BY apartment_number
	`

	rows, err := r.db.QueryContext(ctx, query, floor)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments by floor: %w", err)
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// GetByEntrance отримує всі квартири в під'їзді.
func (r *ApartmentRepository) GetByEntrance(ctx context.Context, entrance int) ([]*entity.Apartment, error) {
	query := `
		SELECT 
			id, apartment_number, floor, entrance, area_total, area_living,
			rooms_count, cadastral_number, notes, is_active,
			deleted_at, created_at, updated_at
		FROM apartments
		WHERE entrance = ? AND deleted_at IS NULL
		ORDER BY floor, apartment_number
	`

	rows, err := r.db.QueryContext(ctx, query, entrance)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments by entrance: %w", err)
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// GetStatistics отримує статистику по квартирах.
func (r *ApartmentRepository) GetStatistics(ctx context.Context) (*repository.ApartmentStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(area_total), 0) as total_area,
			COALESCE(AVG(area_total), 0) as avg_area,
			COALESCE(MIN(area_total), 0) as min_area,
			COALESCE(MAX(area_total), 0) as max_area,
			COUNT(DISTINCT floor) as floor_count,
			COUNT(DISTINCT entrance) as entrance_count
		FROM apartments
		WHERE deleted_at IS NULL
	`

	stats := &repository.ApartmentStatistics{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalApartments,
		&stats.TotalArea,
		&stats.AverageArea,
		&stats.MinArea,
		&stats.MaxArea,
		&stats.FloorCount,
		&stats.EntranceCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Підрахунок квартир з/без власників
	queryWithOwners := `
		SELECT COUNT(DISTINCT a.id)
		FROM apartments a
		INNER JOIN ownership_shares os ON a.id = os.apartment_id
		WHERE a.deleted_at IS NULL 
		  AND os.is_active = 1 
		  AND os.deleted_at IS NULL
	`
	err = r.db.QueryRowContext(ctx, queryWithOwners).Scan(&stats.WithOwners)
	if err != nil {
		stats.WithOwners = 0
	}

	stats.WithoutOwners = stats.TotalApartments - stats.WithOwners

	return stats, nil
}

// scanApartment - helper для сканування однієї квартири.
func (r *ApartmentRepository) scanApartment(ctx context.Context, query string, args ...interface{}) (*entity.Apartment, error) {
	apt := &entity.Apartment{}
	var entrance, roomsCount sql.NullInt64
	var areaLiving sql.NullFloat64
	var cadastralNumber, notes sql.NullString
	var deletedAt, createdAt, updatedAt sql.NullInt64
	var isActive int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&apt.ID,
		&apt.ApartmentNumber,
		&apt.Floor,
		&entrance,
		&apt.AreaTotal,
		&areaLiving,
		&roomsCount,
		&cadastralNumber,
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
		return nil, fmt.Errorf("failed to scan apartment: %w", err)
	}

	if entrance.Valid {
		e := int(entrance.Int64)
		apt.Entrance = &e
	}
	if areaLiving.Valid {
		apt.AreaLiving = &areaLiving.Float64
	}
	if roomsCount.Valid {
		r := int(roomsCount.Int64)
		apt.RoomsCount = &r
	}
	apt.CadastralNumber = nullStringToPtr(cadastralNumber)
	apt.Notes = nullStringToPtr(notes)
	apt.IsActive = intToBool(isActive)
	apt.DeletedAt = nullInt64ToTimePtr(deletedAt)
	apt.CreatedAt = time.Unix(createdAt.Int64, 0)
	apt.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return apt, nil
}

// scanApartments - helper для сканування списку квартир.
func (r *ApartmentRepository) scanApartments(rows *sql.Rows) ([]*entity.Apartment, error) {
	apartments := make([]*entity.Apartment, 0)

	for rows.Next() {
		apt := &entity.Apartment{}
		var entrance, roomsCount sql.NullInt64
		var areaLiving sql.NullFloat64
		var cadastralNumber, notes sql.NullString
		var deletedAt, createdAt, updatedAt sql.NullInt64
		var isActive int

		err := rows.Scan(
			&apt.ID,
			&apt.ApartmentNumber,
			&apt.Floor,
			&entrance,
			&apt.AreaTotal,
			&areaLiving,
			&roomsCount,
			&cadastralNumber,
			&notes,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan apartment: %w", err)
		}

		if entrance.Valid {
			e := int(entrance.Int64)
			apt.Entrance = &e
		}
		if areaLiving.Valid {
			apt.AreaLiving = &areaLiving.Float64
		}
		if roomsCount.Valid {
			r := int(roomsCount.Int64)
			apt.RoomsCount = &r
		}
		apt.CadastralNumber = nullStringToPtr(cadastralNumber)
		apt.Notes = nullStringToPtr(notes)
		apt.IsActive = intToBool(isActive)
		apt.DeletedAt = nullInt64ToTimePtr(deletedAt)
		apt.CreatedAt = time.Unix(createdAt.Int64, 0)
		apt.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		apartments = append(apartments, apt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating apartments: %w", err)
	}

	return apartments, nil
}
