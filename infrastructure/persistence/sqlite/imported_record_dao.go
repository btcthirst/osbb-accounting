// infrastructure/persistence/sqlite/imported_record_dao.go
package sqlite

import (
	"database/sql"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
)

// ImportedRecordDAO implements ImportedRecordRepository.
type ImportedRecordDAO struct {
	db *sql.DB
}

// NewImportedRecordDAO creates a new DAO.
func NewImportedRecordDAO(db *sql.DB) repository.ImportedRecordRepository {
	return &ImportedRecordDAO{db: db}
}

// Create створює новий запис.
func (dao *ImportedRecordDAO) Create(record *entity.ImportedMonthlyRecord) (int64, error) {
	query := `
		INSERT INTO imported_monthly_records (
			import_batch_id, period_month, period_year, sheet_name,
			apartment_number, owner_name, account_number,
			opening_debit, opening_credit, closing_debit, closing_credit,
			total_area, discount_area, discount_percent,
			tariff, charge_amount, discount_amount, amount_due,
			amount_paid, corrections, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := dao.db.Exec(
		query,
		record.ImportBatchID,
		record.PeriodMonth,
		record.PeriodYear,
		record.SheetName,
		record.ApartmentNumber,
		record.OwnerName,
		record.AccountNumber,
		record.OpeningDebit,
		record.OpeningCredit,
		record.ClosingDebit,
		record.ClosingCredit,
		record.TotalArea,
		record.DiscountArea,
		record.DiscountPercent,
		record.Tariff,
		record.ChargeAmount,
		record.DiscountAmount,
		record.AmountDue,
		record.AmountPaid,
		record.Corrections,
		record.Notes,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// BulkCreate створює багато записів ефективно.
func (dao *ImportedRecordDAO) BulkCreate(records []*entity.ImportedMonthlyRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Use transaction for bulk insert
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare statement
	stmt, err := tx.Prepare(`
		INSERT INTO imported_monthly_records (
			import_batch_id, period_month, period_year, sheet_name,
			apartment_number, owner_name, account_number,
			opening_debit, opening_credit, closing_debit, closing_credit,
			total_area, discount_area, discount_percent,
			tariff, charge_amount, discount_amount, amount_due,
			amount_paid, corrections, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert all records
	for _, record := range records {
		_, err := stmt.Exec(
			record.ImportBatchID,
			record.PeriodMonth,
			record.PeriodYear,
			record.SheetName,
			record.ApartmentNumber,
			record.OwnerName,
			record.AccountNumber,
			record.OpeningDebit,
			record.OpeningCredit,
			record.ClosingDebit,
			record.ClosingCredit,
			record.TotalArea,
			record.DiscountArea,
			record.DiscountPercent,
			record.Tariff,
			record.ChargeAmount,
			record.DiscountAmount,
			record.AmountDue,
			record.AmountPaid,
			record.Corrections,
			record.Notes,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FindByID знаходить запис за ID.
func (dao *ImportedRecordDAO) FindByID(id int64) (*entity.ImportedMonthlyRecord, error) {
	query := `
		SELECT id, import_batch_id, period_month, period_year, sheet_name,
		       apartment_number, owner_name, account_number,
		       opening_debit, opening_credit, closing_debit, closing_credit,
		       total_area, discount_area, discount_percent,
		       tariff, charge_amount, discount_amount, amount_due,
		       amount_paid, corrections, notes,
		       apartment_id, owner_id, ownership_share_id, charge_id, payment_id,
		       is_migrated, migrated_at, created_at, updated_at
		FROM imported_monthly_records
		WHERE id = ?
	`

	row := dao.db.QueryRow(query, id)
	return dao.scanRecord(row)
}

// FindByBatchID знаходить всі записи батчу.
func (dao *ImportedRecordDAO) FindByBatchID(batchID int64) ([]*entity.ImportedMonthlyRecord, error) {
	query := `
		SELECT id, import_batch_id, period_month, period_year, sheet_name,
		       apartment_number, owner_name, account_number,
		       opening_debit, opening_credit, closing_debit, closing_credit,
		       total_area, discount_area, discount_percent,
		       tariff, charge_amount, discount_amount, amount_due,
		       amount_paid, corrections, notes,
		       apartment_id, owner_id, ownership_share_id, charge_id, payment_id,
		       is_migrated, migrated_at, created_at, updated_at
		FROM imported_monthly_records
		WHERE import_batch_id = ?
		ORDER BY period_month, apartment_number
	`

	rows, err := dao.db.Query(query, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dao.scanRecords(rows)
}

// FindByPeriod знаходить записи за період.
func (dao *ImportedRecordDAO) FindByPeriod(month, year int) ([]*entity.ImportedMonthlyRecord, error) {
	query := `
		SELECT id, import_batch_id, period_month, period_year, sheet_name,
		       apartment_number, owner_name, account_number,
		       opening_debit, opening_credit, closing_debit, closing_credit,
		       total_area, discount_area, discount_percent,
		       tariff, charge_amount, discount_amount, amount_due,
		       amount_paid, corrections, notes,
		       apartment_id, owner_id, ownership_share_id, charge_id, payment_id,
		       is_migrated, migrated_at, created_at, updated_at
		FROM imported_monthly_records
		WHERE period_month = ? AND period_year = ?
		ORDER BY apartment_number
	`

	rows, err := dao.db.Query(query, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dao.scanRecords(rows)
}

// FindByApartmentNumber знаходить записи за номером квартири.
func (dao *ImportedRecordDAO) FindByApartmentNumber(apartmentNumber string) ([]*entity.ImportedMonthlyRecord, error) {
	query := `
		SELECT id, import_batch_id, period_month, period_year, sheet_name,
		       apartment_number, owner_name, account_number,
		       opening_debit, opening_credit, closing_debit, closing_credit,
		       total_area, discount_area, discount_percent,
		       tariff, charge_amount, discount_amount, amount_due,
		       amount_paid, corrections, notes,
		       apartment_id, owner_id, ownership_share_id, charge_id, payment_id,
		       is_migrated, migrated_at, created_at, updated_at
		FROM imported_monthly_records
		WHERE apartment_number = ?
		ORDER BY period_year DESC, period_month DESC
	`

	rows, err := dao.db.Query(query, apartmentNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dao.scanRecords(rows)
}

// FindUnmigrated знаходить незміґровані записи.
func (dao *ImportedRecordDAO) FindUnmigrated(batchID *int64) ([]*entity.ImportedMonthlyRecord, error) {
	var query string
	var args []interface{}

	if batchID != nil {
		query = `
			SELECT id, import_batch_id, period_month, period_year, sheet_name,
			       apartment_number, owner_name, account_number,
			       opening_debit, opening_credit, closing_debit, closing_credit,
			       total_area, discount_area, discount_percent,
			       tariff, charge_amount, discount_amount, amount_due,
			       amount_paid, corrections, notes,
			       apartment_id, owner_id, ownership_share_id, charge_id, payment_id,
			       is_migrated, migrated_at, created_at, updated_at
			FROM imported_monthly_records
			WHERE is_migrated = 0 AND import_batch_id = ?
			ORDER BY period_month, apartment_number
		`
		args = []interface{}{*batchID}
	} else {
		query = `
			SELECT id, import_batch_id, period_month, period_year, sheet_name,
			       apartment_number, owner_name, account_number,
			       opening_debit, opening_credit, closing_debit, closing_credit,
			       total_area, discount_area, discount_percent,
			       tariff, charge_amount, discount_amount, amount_due,
			       amount_paid, corrections, notes,
			       apartment_id, owner_id, ownership_share_id, charge_id, payment_id,
			       is_migrated, migrated_at, created_at, updated_at
			FROM imported_monthly_records
			WHERE is_migrated = 0
			ORDER BY import_batch_id, period_month, apartment_number
		`
	}

	rows, err := dao.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dao.scanRecords(rows)
}

// Update оновлює запис.
func (dao *ImportedRecordDAO) Update(record *entity.ImportedMonthlyRecord) error {
	query := `
		UPDATE imported_monthly_records
		SET period_month = ?, period_year = ?, sheet_name = ?,
		    apartment_number = ?, owner_name = ?, account_number = ?,
		    opening_debit = ?, opening_credit = ?, closing_debit = ?, closing_credit = ?,
		    total_area = ?, discount_area = ?, discount_percent = ?,
		    tariff = ?, charge_amount = ?, discount_amount = ?, amount_due = ?,
		    amount_paid = ?, corrections = ?, notes = ?,
		    apartment_id = ?, owner_id = ?, ownership_share_id = ?,
		    charge_id = ?, payment_id = ?,
		    is_migrated = ?, migrated_at = ?
		WHERE id = ?
	`

	var migratedAt sql.NullInt64
	if record.MigratedAt != nil {
		migratedAt = sql.NullInt64{Int64: record.MigratedAt.Unix(), Valid: true}
	}

	_, err := dao.db.Exec(
		query,
		record.PeriodMonth,
		record.PeriodYear,
		record.SheetName,
		record.ApartmentNumber,
		record.OwnerName,
		record.AccountNumber,
		record.OpeningDebit,
		record.OpeningCredit,
		record.ClosingDebit,
		record.ClosingCredit,
		record.TotalArea,
		record.DiscountArea,
		record.DiscountPercent,
		record.Tariff,
		record.ChargeAmount,
		record.DiscountAmount,
		record.AmountDue,
		record.AmountPaid,
		record.Corrections,
		record.Notes,
		record.ApartmentID,
		record.OwnerID,
		record.OwnershipShareID,
		record.ChargeID,
		record.PaymentID,
		record.IsMigrated,
		migratedAt,
		record.ID,
	)

	return err
}

// MarkAsMigrated позначає запис як міґрований.
func (dao *ImportedRecordDAO) MarkAsMigrated(
	id int64,
	apartmentID, ownerID, ownershipShareID, chargeID, paymentID *int64,
) error {
	query := `
		UPDATE imported_monthly_records
		SET is_migrated = 1,
		    migrated_at = ?,
		    apartment_id = ?,
		    owner_id = ?,
		    ownership_share_id = ?,
		    charge_id = ?,
		    payment_id = ?
		WHERE id = ?
	`

	_, err := dao.db.Exec(
		query,
		time.Now().Unix(),
		apartmentID,
		ownerID,
		ownershipShareID,
		chargeID,
		paymentID,
		id,
	)

	return err
}

// Delete видаляє запис.
func (dao *ImportedRecordDAO) Delete(id int64) error {
	query := `DELETE FROM imported_monthly_records WHERE id = ?`
	_, err := dao.db.Exec(query, id)
	return err
}

// DeleteByBatchID видаляє всі записи батчу.
func (dao *ImportedRecordDAO) DeleteByBatchID(batchID int64) error {
	query := `DELETE FROM imported_monthly_records WHERE import_batch_id = ?`
	_, err := dao.db.Exec(query, batchID)
	return err
}

// CountByBatchID рахує записи в батчі.
func (dao *ImportedRecordDAO) CountByBatchID(batchID int64) (int, error) {
	query := `SELECT COUNT(*) FROM imported_monthly_records WHERE import_batch_id = ?`
	var count int
	err := dao.db.QueryRow(query, batchID).Scan(&count)
	return count, err
}

// scanRecord helper to scan a single record from Row.
func (dao *ImportedRecordDAO) scanRecord(row *sql.Row) (*entity.ImportedMonthlyRecord, error) {
	var record entity.ImportedMonthlyRecord
	var sheetName, accountNumber, notes sql.NullString
	var totalArea, tariff, corrections sql.NullFloat64
	var apartmentID, ownerID, ownershipShareID, chargeID, paymentID sql.NullInt64
	var migratedAt, createdAtUnix, updatedAtUnix sql.NullInt64
	var isMigrated int

	err := row.Scan(
		&record.ID,
		&record.ImportBatchID,
		&record.PeriodMonth,
		&record.PeriodYear,
		&sheetName,
		&record.ApartmentNumber,
		&record.OwnerName,
		&accountNumber,
		&record.OpeningDebit,
		&record.OpeningCredit,
		&record.ClosingDebit,
		&record.ClosingCredit,
		&totalArea,
		&record.DiscountArea,
		&record.DiscountPercent,
		&tariff,
		&record.ChargeAmount,
		&record.DiscountAmount,
		&record.AmountDue,
		&record.AmountPaid,
		&corrections,
		&notes,
		&apartmentID,
		&ownerID,
		&ownershipShareID,
		&chargeID,
		&paymentID,
		&isMigrated,
		&migratedAt,
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
	if sheetName.Valid {
		record.SheetName = &sheetName.String
	}
	if accountNumber.Valid {
		record.AccountNumber = &accountNumber.String
	}
	if notes.Valid {
		record.Notes = &notes.String
	}
	if totalArea.Valid {
		record.TotalArea = &totalArea.Float64
	}
	if tariff.Valid {
		record.Tariff = &tariff.Float64
	}
	if corrections.Valid {
		record.Corrections = &corrections.Float64
	}
	if apartmentID.Valid {
		record.ApartmentID = &apartmentID.Int64
	}
	if ownerID.Valid {
		record.OwnerID = &ownerID.Int64
	}
	if ownershipShareID.Valid {
		record.OwnershipShareID = &ownershipShareID.Int64
	}
	if chargeID.Valid {
		record.ChargeID = &chargeID.Int64
	}
	if paymentID.Valid {
		record.PaymentID = &paymentID.Int64
	}
	if migratedAt.Valid {
		t := time.Unix(migratedAt.Int64, 0)
		record.MigratedAt = &t
	}
	if createdAtUnix.Valid {
		record.CreatedAt = time.Unix(createdAtUnix.Int64, 0)
	}
	if updatedAtUnix.Valid {
		record.UpdatedAt = time.Unix(updatedAtUnix.Int64, 0)
	}

	record.IsMigrated = isMigrated == 1

	return &record, nil
}

// scanRecords helper to scan multiple records from Rows.
func (dao *ImportedRecordDAO) scanRecords(rows *sql.Rows) ([]*entity.ImportedMonthlyRecord, error) {
	var records []*entity.ImportedMonthlyRecord

	for rows.Next() {
		var record entity.ImportedMonthlyRecord
		var sheetName, accountNumber, notes sql.NullString
		var totalArea, tariff, corrections sql.NullFloat64
		var apartmentID, ownerID, ownershipShareID, chargeID, paymentID sql.NullInt64
		var migratedAt, createdAtUnix, updatedAtUnix sql.NullInt64
		var isMigrated int

		err := rows.Scan(
			&record.ID,
			&record.ImportBatchID,
			&record.PeriodMonth,
			&record.PeriodYear,
			&sheetName,
			&record.ApartmentNumber,
			&record.OwnerName,
			&accountNumber,
			&record.OpeningDebit,
			&record.OpeningCredit,
			&record.ClosingDebit,
			&record.ClosingCredit,
			&totalArea,
			&record.DiscountArea,
			&record.DiscountPercent,
			&tariff,
			&record.ChargeAmount,
			&record.DiscountAmount,
			&record.AmountDue,
			&record.AmountPaid,
			&corrections,
			&notes,
			&apartmentID,
			&ownerID,
			&ownershipShareID,
			&chargeID,
			&paymentID,
			&isMigrated,
			&migratedAt,
			&createdAtUnix,
			&updatedAtUnix,
		)

		if err != nil {
			return nil, err
		}

		// Convert NULLs
		if sheetName.Valid {
			record.SheetName = &sheetName.String
		}
		if accountNumber.Valid {
			record.AccountNumber = &accountNumber.String
		}
		if notes.Valid {
			record.Notes = &notes.String
		}
		if totalArea.Valid {
			record.TotalArea = &totalArea.Float64
		}
		if tariff.Valid {
			record.Tariff = &tariff.Float64
		}
		if corrections.Valid {
			record.Corrections = &corrections.Float64
		}
		if apartmentID.Valid {
			record.ApartmentID = &apartmentID.Int64
		}
		if ownerID.Valid {
			record.OwnerID = &ownerID.Int64
		}
		if ownershipShareID.Valid {
			record.OwnershipShareID = &ownershipShareID.Int64
		}
		if chargeID.Valid {
			record.ChargeID = &chargeID.Int64
		}
		if paymentID.Valid {
			record.PaymentID = &paymentID.Int64
		}
		if migratedAt.Valid {
			t := time.Unix(migratedAt.Int64, 0)
			record.MigratedAt = &t
		}
		if createdAtUnix.Valid {
			record.CreatedAt = time.Unix(createdAtUnix.Int64, 0)
		}
		if updatedAtUnix.Valid {
			record.UpdatedAt = time.Unix(updatedAtUnix.Int64, 0)
		}

		record.IsMigrated = isMigrated == 1

		records = append(records, &record)
	}

	return records, rows.Err()
}
