package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"time"
)

// SQLiteApartmentRepository - SQLite реалізація ApartmentRepository.
type SQLiteApartmentRepository struct {
	db *sql.DB
}

// NewSQLiteApartmentRepository створює новий екземпляр SQLite репозиторію квартир.
func NewSQLiteApartmentRepository(db *sql.DB) repository.ApartmentRepository {
	return &SQLiteApartmentRepository{db: db}
}

// Save реалізує збереження нової квартири.
func (r *SQLiteApartmentRepository) Save(apartment *domain.Apartment) error {
	// Валідація
	if err := apartment.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO apartments (
			apartment_number, floor, entrance, area, rooms,
			owner_name, owner_phone, owner_email,
			residents_count, notes, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	apartment.CreatedAt = now
	apartment.UpdatedAt = now
	apartment.IsActive = true

	result, err := r.db.Exec(
		query,
		apartment.ApartmentNumber,
		apartment.Floor,
		apartment.Entrance,
		apartment.Area,
		apartment.Rooms,
		apartment.OwnerName,
		apartment.OwnerPhone,
		apartment.OwnerEmail,
		apartment.ResidentsCount,
		apartment.Notes,
		apartment.IsActive,
		now,
		now,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrApartmentAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	apartment.ID = int(id)
	return nil
}

// FindByID реалізує пошук квартири за ID.
func (r *SQLiteApartmentRepository) FindByID(id int) (*domain.Apartment, error) {
	query := `
		SELECT id, apartment_number, floor, entrance, area, rooms,
		       owner_name, owner_phone, owner_email,
		       residents_count, notes, is_active,
		       created_at, updated_at
		FROM apartments
		WHERE id = ? AND is_active = 1
	`

	apartment := &domain.Apartment{}
	err := r.db.QueryRow(query, id).Scan(
		&apartment.ID,
		&apartment.ApartmentNumber,
		&apartment.Floor,
		&apartment.Entrance,
		&apartment.Area,
		&apartment.Rooms,
		&apartment.OwnerName,
		&apartment.OwnerPhone,
		&apartment.OwnerEmail,
		&apartment.ResidentsCount,
		&apartment.Notes,
		&apartment.IsActive,
		&apartment.CreatedAt,
		&apartment.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApartmentNotFound
		}
		return nil, err
	}

	return apartment, nil
}

// FindByNumber реалізує пошук квартири за номером.
func (r *SQLiteApartmentRepository) FindByNumber(apartmentNumber string) (*domain.Apartment, error) {
	query := `
		SELECT id, apartment_number, floor, entrance, area, rooms,
		       owner_name, owner_phone, owner_email,
		       residents_count, notes, is_active,
		       created_at, updated_at
		FROM apartments
		WHERE apartment_number = ? AND is_active = 1
	`

	apartment := &domain.Apartment{}
	err := r.db.QueryRow(query, apartmentNumber).Scan(
		&apartment.ID,
		&apartment.ApartmentNumber,
		&apartment.Floor,
		&apartment.Entrance,
		&apartment.Area,
		&apartment.Rooms,
		&apartment.OwnerName,
		&apartment.OwnerPhone,
		&apartment.OwnerEmail,
		&apartment.ResidentsCount,
		&apartment.Notes,
		&apartment.IsActive,
		&apartment.CreatedAt,
		&apartment.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApartmentNotFound
		}
		return nil, err
	}

	return apartment, nil
}

// FindAll реалізує отримання всіх активних квартир.
func (r *SQLiteApartmentRepository) FindAll() ([]*domain.Apartment, error) {
	query := `
		SELECT id, apartment_number, floor, entrance, area, rooms,
		       owner_name, owner_phone, owner_email,
		       residents_count, notes, is_active,
		       created_at, updated_at
		FROM apartments
		WHERE is_active = 1
		ORDER BY CAST(apartment_number AS INTEGER), apartment_number
	`

	return r.queryApartments(query)
}

// FindByFilter реалізує пошук з фільтрами.
func (r *SQLiteApartmentRepository) FindByFilter(filter *domain.ApartmentFilter) ([]*domain.Apartment, error) {
	if filter == nil {
		filter = domain.DefaultApartmentFilter()
	}

	// Базовий запит
	query := `
		SELECT id, apartment_number, floor, entrance, area, rooms,
		       owner_name, owner_phone, owner_email,
		       residents_count, notes, is_active,
		       created_at, updated_at
		FROM apartments
		WHERE 1=1
	`
	args := []interface{}{}

	// Додаємо умови фільтрації
	if filter.SearchQuery != "" {
		query += ` AND (apartment_number LIKE ? OR owner_name LIKE ?)`
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if filter.Floor != nil {
		query += ` AND floor = ?`
		args = append(args, *filter.Floor)
	}

	if filter.Entrance != nil {
		query += ` AND entrance = ?`
		args = append(args, *filter.Entrance)
	}

	if filter.MinArea != nil {
		query += ` AND area >= ?`
		args = append(args, *filter.MinArea)
	}

	if filter.MaxArea != nil {
		query += ` AND area <= ?`
		args = append(args, *filter.MaxArea)
	}

	if filter.Rooms != nil {
		query += ` AND rooms = ?`
		args = append(args, *filter.Rooms)
	}

	if filter.IsActive != nil {
		query += ` AND is_active = ?`
		args = append(args, *filter.IsActive)
	}

	// Сортування
	sortBy := "apartment_number"
	if filter.SortBy != "" {
		// Валідуємо поле сортування
		validFields := map[string]bool{
			"apartment_number": true,
			"floor":            true,
			"area":             true,
			"owner_name":       true,
			"residents_count":  true,
		}
		if validFields[filter.SortBy] {
			sortBy = filter.SortBy
		}
	}

	query += fmt.Sprintf(" ORDER BY %s", sortBy)
	if filter.SortDesc {
		query += " DESC"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// FindByFloor реалізує пошук квартир по поверху.
func (r *SQLiteApartmentRepository) FindByFloor(floor int) ([]*domain.Apartment, error) {
	query := `
		SELECT id, apartment_number, floor, entrance, area, rooms,
		       owner_name, owner_phone, owner_email,
		       residents_count, notes, is_active,
		       created_at, updated_at
		FROM apartments
		WHERE floor = ? AND is_active = 1
		ORDER BY apartment_number
	`

	rows, err := r.db.Query(query, floor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// FindByEntrance реалізує пошук квартир по під'їзду.
func (r *SQLiteApartmentRepository) FindByEntrance(entrance int) ([]*domain.Apartment, error) {
	query := `
		SELECT id, apartment_number, floor, entrance, area, rooms,
		       owner_name, owner_phone, owner_email,
		       residents_count, notes, is_active,
		       created_at, updated_at
		FROM apartments
		WHERE entrance = ? AND is_active = 1
		ORDER BY floor, apartment_number
	`

	rows, err := r.db.Query(query, entrance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// Update реалізує оновлення квартири.
func (r *SQLiteApartmentRepository) Update(apartment *domain.Apartment) error {
	if err := apartment.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE apartments
		SET apartment_number = ?, floor = ?, entrance = ?, area = ?, rooms = ?,
		    owner_name = ?, owner_phone = ?, owner_email = ?,
		    residents_count = ?, notes = ?, updated_at = ?
		WHERE id = ? AND is_active = 1
	`

	apartment.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		apartment.ApartmentNumber,
		apartment.Floor,
		apartment.Entrance,
		apartment.Area,
		apartment.Rooms,
		apartment.OwnerName,
		apartment.OwnerPhone,
		apartment.OwnerEmail,
		apartment.ResidentsCount,
		apartment.Notes,
		apartment.UpdatedAt,
		apartment.ID,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrApartmentAlreadyExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrApartmentNotFound
	}

	return nil
}

// Deactivate реалізує м'яке видалення квартири.
func (r *SQLiteApartmentRepository) Deactivate(id int) error {
	query := `UPDATE apartments SET is_active = 0, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrApartmentNotFound
	}

	return nil
}

// Activate реалізує активацію квартири.
func (r *SQLiteApartmentRepository) Activate(id int) error {
	query := `UPDATE apartments SET is_active = 1, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrApartmentNotFound
	}

	return nil
}

// Delete реалізує фізичне видалення квартири.
func (r *SQLiteApartmentRepository) Delete(id int) error {
	query := `DELETE FROM apartments WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrApartmentNotFound
	}

	return nil
}

// Count реалізує підрахунок квартир.
func (r *SQLiteApartmentRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM apartments WHERE is_active = 1").Scan(&count)
	return count, err
}

// GetStatistics реалізує отримання статистики.
func (r *SQLiteApartmentRepository) GetStatistics() (*repository.ApartmentStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(area), 0) as total_area,
			COALESCE(AVG(area), 0) as avg_area,
			COALESCE(SUM(residents_count), 0) as total_residents,
			COALESCE(MIN(floor), 0) as min_floor,
			COALESCE(MAX(floor), 0) as max_floor,
			COUNT(DISTINCT entrance) as entrances_count
		FROM apartments
		WHERE is_active = 1
	`

	stats := &repository.ApartmentStatistics{}
	err := r.db.QueryRow(query).Scan(
		&stats.TotalApartments,
		&stats.TotalArea,
		&stats.AverageArea,
		&stats.TotalResidents,
		&stats.MinFloor,
		&stats.MaxFloor,
		&stats.EntrancesCount,
	)

	return stats, err
}

// queryApartments - допоміжна функція для виконання запиту без параметрів.
func (r *SQLiteApartmentRepository) queryApartments(query string) ([]*domain.Apartment, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanApartments(rows)
}

// scanApartments - допоміжна функція для сканування результатів запиту.
func (r *SQLiteApartmentRepository) scanApartments(rows *sql.Rows) ([]*domain.Apartment, error) {
	var apartments []*domain.Apartment

	for rows.Next() {
		apartment := &domain.Apartment{}
		err := rows.Scan(
			&apartment.ID,
			&apartment.ApartmentNumber,
			&apartment.Floor,
			&apartment.Entrance,
			&apartment.Area,
			&apartment.Rooms,
			&apartment.OwnerName,
			&apartment.OwnerPhone,
			&apartment.OwnerEmail,
			&apartment.ResidentsCount,
			&apartment.Notes,
			&apartment.IsActive,
			&apartment.CreatedAt,
			&apartment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apartments = append(apartments, apartment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apartments, nil
}
