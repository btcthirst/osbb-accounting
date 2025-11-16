// infrastructure/persistence/sqlite/contractor_repository.go
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

// ContractorRepository реалізує repository.ContractorRepository для SQLite.
type ContractorRepository struct {
	db *sql.DB
}

// NewContractorRepository створює новий ContractorRepository.
func NewContractorRepository(db *sql.DB) *ContractorRepository {
	return &ContractorRepository{db: db}
}

// Create створює нового контрагента.
func (r *ContractorRepository) Create(ctx context.Context, contractor *entity.Contractor) error {
	query := `
		INSERT INTO contractors (
			name, edrpou, contractor_type, contact_person, phone, email, address,
			bank_account, bank_name, bank_mfo, contract_number, contract_date,
			notes, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		contractor.Name,
		contractor.EDRPOU,
		string(contractor.ContractorType),
		contractor.ContactPerson,
		contractor.Phone,
		contractor.Email,
		contractor.Address,
		contractor.BankAccount,
		contractor.BankName,
		contractor.BankMFO,
		contractor.ContractNumber,
		timeToNullInt64(contractor.ContractDate),
		contractor.Notes,
		boolToInt(contractor.IsActive),
		now,
		now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "edrpou") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"contractor with this EDRPOU already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "edrpou")
			}
		}
		return fmt.Errorf("failed to create contractor: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	contractor.ID = id
	contractor.CreatedAt = time.Unix(now, 0)
	contractor.UpdatedAt = time.Unix(now, 0)

	return nil
}

// GetByID отримує контрагента за ID.
func (r *ContractorRepository) GetByID(ctx context.Context, id int64) (*entity.Contractor, error) {
	query := `
		SELECT 
			id, name, edrpou, contractor_type, contact_person, phone, email, address,
			bank_account, bank_name, bank_mfo, contract_number, contract_date,
			notes, is_active, deleted_at, created_at, updated_at
		FROM contractors
		WHERE id = ? AND deleted_at IS NULL
	`

	return r.scanContractor(ctx, query, id)
}

// GetByEDRPOU отримує контрагента за ЄДРПОУ.
func (r *ContractorRepository) GetByEDRPOU(ctx context.Context, edrpou string) (*entity.Contractor, error) {
	query := `
		SELECT 
			id, name, edrpou, contractor_type, contact_person, phone, email, address,
			bank_account, bank_name, bank_mfo, contract_number, contract_date,
			notes, is_active, deleted_at, created_at, updated_at
		FROM contractors
		WHERE edrpou = ? AND deleted_at IS NULL
	`

	return r.scanContractor(ctx, query, edrpou)
}

// Update оновлює дані контрагента.
func (r *ContractorRepository) Update(ctx context.Context, contractor *entity.Contractor) error {
	query := `
		UPDATE contractors
		SET name = ?,
		    contractor_type = ?,
		    contact_person = ?,
		    phone = ?,
		    email = ?,
		    address = ?,
		    bank_account = ?,
		    bank_name = ?,
		    bank_mfo = ?,
		    contract_number = ?,
		    contract_date = ?,
		    notes = ?,
		    is_active = ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query,
		contractor.Name,
		string(contractor.ContractorType),
		contractor.ContactPerson,
		contractor.Phone,
		contractor.Email,
		contractor.Address,
		contractor.BankAccount,
		contractor.BankName,
		contractor.BankMFO,
		contractor.ContractNumber,
		timeToNullInt64(contractor.ContractDate),
		contractor.Notes,
		boolToInt(contractor.IsActive),
		now,
		contractor.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "edrpou") {
				return domainErrors.NewDomainError(
					domainErrors.CodeDuplicateEntry,
					"contractor with this EDRPOU already exists",
					domainErrors.ErrAlreadyExists,
				).WithDetails("field", "edrpou")
			}
		}
		return fmt.Errorf("failed to update contractor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	contractor.UpdatedAt = time.Unix(now, 0)

	return nil
}

// SoftDelete виконує м'яке видалення контрагента.
func (r *ContractorRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE contractors
		SET deleted_at = ?, is_active = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete contractor: %w", err)
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

// Restore відновлює видаленого контрагента.
func (r *ContractorRepository) Restore(ctx context.Context, id int64) error {
	query := `
		UPDATE contractors
		SET deleted_at = NULL, is_active = 1, updated_at = ?
		WHERE id = ? AND deleted_at IS NOT NULL
	`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore contractor: %w", err)
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

// List отримує список контрагентів з фільтрацією.
func (r *ContractorRepository) List(ctx context.Context, filter repository.ContractorFilter) ([]*entity.Contractor, error) {
	query := `
		SELECT 
			id, name, edrpou, contractor_type, contact_person, phone, email, address,
			bank_account, bank_name, bank_mfo, contract_number, contract_date,
			notes, is_active, deleted_at, created_at, updated_at
		FROM contractors
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
		query += ` AND (
			name LIKE ? OR 
			edrpou LIKE ? OR 
			contact_person LIKE ? OR
			phone LIKE ? OR
			email LIKE ?
		)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if filter.ContractorType != nil {
		query += " AND contractor_type = ?"
		args = append(args, string(*filter.ContractorType))
	}

	if filter.HasBankDetails != nil {
		if *filter.HasBankDetails {
			query += " AND (bank_account IS NOT NULL AND bank_account != '')"
		} else {
			query += " AND (bank_account IS NULL OR bank_account = '')"
		}
	}

	if filter.HasContract != nil {
		if *filter.HasContract {
			query += " AND (contract_number IS NOT NULL AND contract_number != '')"
		} else {
			query += " AND (contract_number IS NULL OR contract_number = '')"
		}
	}

	// Сортування
	orderBy := "name"
	if filter.OrderBy != "" {
		switch filter.OrderBy {
		case "name":
			orderBy = "name"
		case "type":
			orderBy = "contractor_type, name"
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
		return nil, fmt.Errorf("failed to list contractors: %w", err)
	}
	defer rows.Close()

	return r.scanContractors(rows)
}

// Count отримує загальну кількість контрагентів.
func (r *ContractorRepository) Count(ctx context.Context, filter repository.ContractorFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM contractors WHERE 1=1`
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
			name LIKE ? OR 
			edrpou LIKE ? OR 
			contact_person LIKE ?
		)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if filter.ContractorType != nil {
		query += " AND contractor_type = ?"
		args = append(args, string(*filter.ContractorType))
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count contractors: %w", err)
	}

	return count, nil
}

// ExistsByEDRPOU перевіряє чи існує контрагент з таким ЄДРПОУ.
func (r *ContractorRepository) ExistsByEDRPOU(ctx context.Context, edrpou string) (bool, error) {
	query := `SELECT COUNT(*) FROM contractors WHERE edrpou = ? AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, edrpou).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check EDRPOU existence: %w", err)
	}

	return count > 0, nil
}

// Search шукає контрагентів за запитом.
func (r *ContractorRepository) Search(ctx context.Context, query string, limit int) ([]*entity.Contractor, error) {
	sqlQuery := `
		SELECT 
			id, name, edrpou, contractor_type, contact_person, phone, email, address,
			bank_account, bank_name, bank_mfo, contract_number, contract_date,
			notes, is_active, deleted_at, created_at, updated_at
		FROM contractors
		WHERE deleted_at IS NULL
		  AND (
		      name LIKE ? OR 
		      edrpou LIKE ? OR 
		      contact_person LIKE ?
		  )
		ORDER BY name
		LIMIT ?
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery,
		searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search contractors: %w", err)
	}
	defer rows.Close()

	return r.scanContractors(rows)
}

// GetByType отримує всіх контрагентів певного типу.
func (r *ContractorRepository) GetByType(ctx context.Context, contractorType entity.ContractorType) ([]*entity.Contractor, error) {
	query := `
		SELECT 
			id, name, edrpou, contractor_type, contact_person, phone, email, address,
			bank_account, bank_name, bank_mfo, contract_number, contract_date,
			notes, is_active, deleted_at, created_at, updated_at
		FROM contractors
		WHERE contractor_type = ? AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, string(contractorType))
	if err != nil {
		return nil, fmt.Errorf("failed to get contractors by type: %w", err)
	}
	defer rows.Close()

	return r.scanContractors(rows)
}

// scanContractor - helper для сканування одного контрагента.
func (r *ContractorRepository) scanContractor(ctx context.Context, query string, args ...interface{}) (*entity.Contractor, error) {
	contractor := &entity.Contractor{}
	var edrpou, contactPerson, phone, email, address sql.NullString
	var bankAccount, bankName, bankMFO, contractNumber, notes sql.NullString
	var contractDate, deletedAt, createdAt, updatedAt sql.NullInt64
	var contractorType string
	var isActive int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&contractor.ID,
		&contractor.Name,
		&edrpou,
		&contractorType,
		&contactPerson,
		&phone,
		&email,
		&address,
		&bankAccount,
		&bankName,
		&bankMFO,
		&contractNumber,
		&contractDate,
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
		return nil, fmt.Errorf("failed to scan contractor: %w", err)
	}

	contractor.EDRPOU = nullStringToPtr(edrpou)
	contractor.ContractorType = entity.ContractorType(contractorType)
	contractor.ContactPerson = nullStringToPtr(contactPerson)
	contractor.Phone = nullStringToPtr(phone)
	contractor.Email = nullStringToPtr(email)
	contractor.Address = nullStringToPtr(address)
	contractor.BankAccount = nullStringToPtr(bankAccount)
	contractor.BankName = nullStringToPtr(bankName)
	contractor.BankMFO = nullStringToPtr(bankMFO)
	contractor.ContractNumber = nullStringToPtr(contractNumber)
	contractor.ContractDate = nullInt64ToTimePtr(contractDate)
	contractor.Notes = nullStringToPtr(notes)
	contractor.IsActive = intToBool(isActive)
	contractor.DeletedAt = nullInt64ToTimePtr(deletedAt)
	contractor.CreatedAt = time.Unix(createdAt.Int64, 0)
	contractor.UpdatedAt = time.Unix(updatedAt.Int64, 0)
	return contractor, nil
}

// scanContractors - helper для сканування списку контрагентів.
func (r *ContractorRepository) scanContractors(rows *sql.Rows) ([]*entity.Contractor, error) {
	contractors := make([]*entity.Contractor, 0)

	for rows.Next() {
		contractor := &entity.Contractor{}
		var edrpou, contactPerson, phone, email, address sql.NullString
		var bankAccount, bankName, bankMFO, contractNumber, notes sql.NullString
		var contractDate, deletedAt, createdAt, updatedAt sql.NullInt64
		var contractorType string
		var isActive int

		err := rows.Scan(
			&contractor.ID,
			&contractor.Name,
			&edrpou,
			&contractorType,
			&contactPerson,
			&phone,
			&email,
			&address,
			&bankAccount,
			&bankName,
			&bankMFO,
			&contractNumber,
			&contractDate,
			&notes,
			&isActive,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contractor: %w", err)
		}

		contractor.EDRPOU = nullStringToPtr(edrpou)
		contractor.ContractorType = entity.ContractorType(contractorType)
		contractor.ContactPerson = nullStringToPtr(contactPerson)
		contractor.Phone = nullStringToPtr(phone)
		contractor.Email = nullStringToPtr(email)
		contractor.Address = nullStringToPtr(address)
		contractor.BankAccount = nullStringToPtr(bankAccount)
		contractor.BankName = nullStringToPtr(bankName)
		contractor.BankMFO = nullStringToPtr(bankMFO)
		contractor.ContractNumber = nullStringToPtr(contractNumber)
		contractor.ContractDate = nullInt64ToTimePtr(contractDate)
		contractor.Notes = nullStringToPtr(notes)
		contractor.IsActive = intToBool(isActive)
		contractor.DeletedAt = nullInt64ToTimePtr(deletedAt)
		contractor.CreatedAt = time.Unix(createdAt.Int64, 0)
		contractor.UpdatedAt = time.Unix(updatedAt.Int64, 0)

		contractors = append(contractors, contractor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contractors: %w", err)
	}

	return contractors, nil
}
