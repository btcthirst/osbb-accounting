// infrastructure/persistence/sqlite/owner_repository.go
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

// OwnerRepository реалізує repository.OwnerRepository для SQLite.
type OwnerRepository struct {
	db *sql.DB
}

// NewOwnerRepository створює новий OwnerRepository.
func NewOwnerRepository(db *sql.DB) *OwnerRepository {
	return &OwnerRepository{db: db}
}

// Create створює нового власника.
func (r *OwnerRepository) Create(ctx context.Context, owner *entity.Owner) error {
	query := `
		INSERT INTO owners (
			first_name, last_name, middle_name, phone, email, tax_number,
			passport_series, passport_number, registered_address, actual_address,
			notes, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		owner.FirstName,
		owner.LastName,
		owner.MiddleName,
		owner.Phone,
		owner.Email,
		owner.TaxNumber,
		owner.PassportSeries,
		owner.PassportNumber,
		owner.RegisteredAddress,
		owner.ActualAddress,
		owner.Notes,
		boolToInt(owner.IsActive),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "tax_number") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"owner with this tax number already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "tax_number")
			}
		}
		return fmt.Errorf("failed to create owner: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	owner.ID = id
	owner.CreatedAt = time.Unix(now, 0)
	owner.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує власника за ID.
func (r *OwnerRepository) GetByID(ctx context.Context, id int64) (*entity.Owner, error) {
	query := `
		SELECT 
			id, first_name, last_name, middle_name, phone, email, tax_number,
			passport_series, passport_number, registered_address, actual_address,
			notes, is_active, deleted_at, created_at, updated_at
		FROM owners
		WHERE id = ? AND deleted_at IS NULL
	`

	return r.scanOwner(ctx, query, id)
}

// Update оновлює дані власника.
func (r *OwnerRepository) Update(ctx context.Context, owner *entity.Owner) error {
	query := `
		UPDATE owners
		SET first_name = ?,
		    last_name = ?,
		    middle_name = ?,
		    phone = ?,
		    email = ?,
		    tax_number = ?,
		    passport_series = ?,
		    passport_number = ?,
		    registered_address = ?,
		    actual_address = ?,
		    notes = ?,
		    is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		owner.FirstName,
		owner.LastName,
		owner.MiddleName,
		owner.Phone,
		owner.Email,
		owner.TaxNumber,
		owner.PassportSeries,
		owner.PassportNumber,
		owner.RegisteredAddress,
		owner.ActualAddress,
		owner.Notes,
		boolToInt(owner.IsActive),
		now,
		owner.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "tax_number") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"owner with this tax number already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "tax_number")
			}
		}
		return fmt.Errorf("failed to update owner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	owner.UpdatedAt = time.Unix(now, 0)

	return nil
}

// SoftDelete виконує м'яке видалення власника.
func (r *OwnerRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE owners
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete owner: %w", err)
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

// Restore відновлює видаленого власника.
func (r *OwnerRepository) Restore(ctx context.Context, id int64) error {
	query := `
		UPDATE owners
		SET deleted_at = NULL, is_active = 1, updated_at = ?
		WHERE id = ? AND deleted_at IS NOT NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore owner: %w", err)
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

// List отримує список власників з фільтрацією.
func (r *OwnerRepository) List(ctx context.Context, filter repository.OwnerFilter) ([]*entity.Owner, error) {
	query := `
		SELECT 
			id, first_name, last_name, middle_name, phone, email, tax_number,
			passport_series, passport_number, registered_address, actual_address,
			notes, is_active, deleted_at, created_at, updated_at
		FROM owners
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	// Фільтр по видаленим
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	// Фільтр по активності
	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	// Пошук
	if filter.SearchQuery != "" {
		query += ` AND (
			first_name LIKE ? OR 
			last_name LIKE ? OR 
			middle_name LIKE ? OR
			phone LIKE ? OR 
			email LIKE ? OR
			tax_number LIKE ?
		)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern,
			searchPattern, searchPattern, searchPattern)
	}

	// Фільтр по наявності ІПН
	if filter.HasTaxNumber != nil {
		if *filter.HasTaxNumber {
			query += " AND tax_number IS NOT NULL AND tax_number != ''"
		} else {
			query += " AND (tax_number IS NULL OR tax_number = '')"
		}
	}

	// Фільтр по наявності контактів
	if filter.HasContact != nil {
		if *filter.HasContact {
			query += " AND ((phone IS NOT NULL AND phone != '') OR (email IS NOT NULL AND email != ''))"
		} else {
			query += " AND (phone IS NULL OR phone = '') AND (email IS NULL OR email = '')"
		}
	}

	// Сортування
	orderBy := "last_name, first_name"
	if filter.OrderBy != "" {
		switch filter.OrderBy {
		case "name":
			orderBy = "last_name, first_name"
		case "created_at":
			orderBy = "created_at"
		case "updated_at":
			orderBy = "updated_at"
		}
	}

	if filter.OrderDesc {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderBy)
	} else {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderBy)
	}

	// Pagination
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
		return nil, fmt.Errorf("failed to list owners: %w", err)
	}
	defer rows.Close()

	return r.scanOwners(rows)
}

// Count отримує загальну кількість власників.
func (r *OwnerRepository) Count(ctx context.Context, filter repository.OwnerFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM owners WHERE 1=1`
	args := make([]interface{}, 0)

	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}

	if filter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, boolToInt(*filter.IsActive))
	}

	if filter.SearchQuery != "" {
		query += ` AND (
			first_name LIKE ? OR 
			last_name LIKE ? OR 
			middle_name LIKE ? OR
			phone LIKE ? OR 
			email LIKE ? OR
			tax_number LIKE ?
		)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern,
			searchPattern, searchPattern, searchPattern)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count owners: %w", err)
	}

	return count, nil
}

// GetByTaxNumber отримує власника за ІПН.
func (r *OwnerRepository) GetByTaxNumber(ctx context.Context, taxNumber string) (*entity.Owner, error) {
	query := `
		SELECT 
			id, first_name, last_name, middle_name, phone, email, tax_number,
			passport_series, passport_number, registered_address, actual_address,
			notes, is_active, deleted_at, created_at, updated_at
		FROM owners
		WHERE tax_number = ? AND deleted_at IS NULL
	`

	return r.scanOwner(ctx, query, taxNumber)
}

// ExistsByTaxNumber перевіряє чи існує власник з таким ІПН.
func (r *OwnerRepository) ExistsByTaxNumber(ctx context.Context, taxNumber string) (bool, error) {
	query := `SELECT COUNT(*) FROM owners WHERE tax_number = ? AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, taxNumber).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check tax number existence: %w", err)
	}

	return count > 0, nil
}

// GetByApartmentID отримує всіх власників квартири.
func (r *OwnerRepository) GetByApartmentID(ctx context.Context, apartmentID int64) ([]*entity.Owner, error) {
	query := `
		SELECT DISTINCT
			o.id, o.first_name, o.last_name, o.middle_name, o.phone, o.email, o.tax_number,
			o.passport_series, o.passport_number, o.registered_address, o.actual_address,
			o.notes, o.is_active, o.deleted_at, o.created_at, o.updated_at
		FROM owners o
		INNER JOIN ownership_shares os ON o.id = os.owner_id
		WHERE os.apartment_id = ? 
		  AND os.is_active = 1
		  AND os.deleted_at IS NULL
		  AND o.deleted_at IS NULL
		ORDER BY o.last_name, o.first_name
	`

	rows, err := r.db.QueryContext(ctx, query, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owners by apartment: %w", err)
	}
	defer rows.Close()

	return r.scanOwners(rows)
}

// Search шукає власників за запитом.
func (r *OwnerRepository) Search(ctx context.Context, query string, limit int) ([]*entity.Owner, error) {
	sqlQuery := `
		SELECT 
			id, first_name, last_name, middle_name, phone, email, tax_number,
			passport_series, passport_number, registered_address, actual_address,
			notes, is_active, deleted_at, created_at, updated_at
		FROM owners
		WHERE deleted_at IS NULL
		  AND (
		      first_name LIKE ? OR 
		      last_name LIKE ? OR 
		      middle_name LIKE ? OR
		      phone LIKE ? OR 
		      email LIKE ?
		  )
		ORDER BY last_name, first_name
		LIMIT ?
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery,
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search owners: %w", err)
	}
	defer rows.Close()

	return r.scanOwners(rows)
}

// GetByName шукає власника за повним ім'ям (ПІБ).
func (r *OwnerRepository) GetByName(ctx context.Context, name string) (*entity.Owner, error) {
	// Спроба знайти за точним співпадінням ПІБ
	// Припускаємо формат "Прізвище Ім'я По-батькові" або "Прізвище Ім'я"
	query := `
		SELECT 
			id, first_name, last_name, middle_name, phone, email, tax_number,
			passport_series, passport_number, registered_address, actual_address,
			notes, is_active, deleted_at, created_at, updated_at
		FROM owners
		WHERE deleted_at IS NULL
		  AND (
		      last_name || ' ' || first_name || ' ' || middle_name = ? OR
		      last_name || ' ' || first_name = ?
		  )
		LIMIT 1
	`

	return r.scanOwner(ctx, query, name, name)
}

// scanOwner - helper для сканування одного власника.
func (r *OwnerRepository) scanOwner(ctx context.Context, query string, args ...interface{}) (*entity.Owner, error) {
	owner := &entity.Owner{}
	var middleName, phone, email, taxNumber sql.NullString
	var passportSeries, passportNumber sql.NullString
	var registeredAddress, actualAddress, notes sql.NullString
	var deletedAt, createdAt, updatedAt sql.NullInt64
	var isActive int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&owner.ID,
		&owner.FirstName,
		&owner.LastName,
		&middleName,
		&phone,
		&email,
		&taxNumber,
		&passportSeries,
		&passportNumber,
		&registeredAddress,
		&actualAddress,
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
		return nil, fmt.Errorf("failed to scan owner: %w", err)
	}

	owner.MiddleName = nullStringToPtr(middleName)
	owner.Phone = nullStringToPtr(phone)
	owner.Email = nullStringToPtr(email)
	owner.TaxNumber = nullStringToPtr(taxNumber)
	owner.PassportSeries = nullStringToPtr(passportSeries)
	owner.PassportNumber = nullStringToPtr(passportNumber)
	owner.RegisteredAddress = nullStringToPtr(registeredAddress)
	owner.ActualAddress = nullStringToPtr(actualAddress)
	owner.Notes = nullStringToPtr(notes)
	owner.IsActive = intToBool(isActive)
	owner.DeletedAt = nullInt64ToTimePtr(deletedAt)
	owner.CreatedAt = time.Unix(createdAt.Int64, 0)
	owner.UpdatedAt = time.Unix(updatedAt.Int64, 0)

	return owner, nil
}

// scanOwners - helper для сканування списку власників.
func (r *OwnerRepository) scanOwners(rows *sql.Rows) ([]*entity.Owner, error) {
	owners := make([]*entity.Owner, 0)

	for rows.Next() {
		owner := &entity.Owner{}
		var middleName, phone, email, taxNumber sql.NullString
		var passportSeries, passportNumber sql.NullString
		var registeredAddress, actualAddress, notes sql.NullString
		var deletedAt, createdAt, updatedAt sql.NullInt64
		var isActive int

		err := rows.Scan(
			&owner.ID,
			&owner.FirstName,
			&owner.LastName,
			&middleName,
			&phone,
			&email,
			&taxNumber,
			&passportSeries,
			&passportNumber,
			&registeredAddress,
			&actualAddress,
			&notes,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan owner: %w", err)
		}

		owner.MiddleName = nullStringToPtr(middleName)
		owner.Phone = nullStringToPtr(phone)
		owner.Email = nullStringToPtr(email)
		owner.TaxNumber = nullStringToPtr(taxNumber)
		owner.PassportSeries = nullStringToPtr(passportSeries)
		owner.PassportNumber = nullStringToPtr(passportNumber)
		owner.RegisteredAddress = nullStringToPtr(registeredAddress)
		owner.ActualAddress = nullStringToPtr(actualAddress)
		owner.Notes = nullStringToPtr(notes)
		owner.IsActive = intToBool(isActive)
		owner.DeletedAt = nullInt64ToTimePtr(deletedAt)
		owner.CreatedAt = time.Unix(createdAt.Int64, 0)
		owner.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		owners = append(owners, owner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating owners: %w", err)
	}

	return owners, nil
}
