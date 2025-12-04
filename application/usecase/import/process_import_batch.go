// application/usecase/import/process_import_batch.go
package importusecase

import (
	"context"
	"fmt"
	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"strconv"
	"strings"
	"time"
)

// ProcessImportBatchUseCase обробляє імпортований батч і створює сутності.
type ProcessImportBatchUseCase struct {
	batchRepo          repository.ImportBatchRepository
	recordRepo         repository.ImportedRecordRepository
	apartmentRepo      repository.ApartmentRepository
	ownerRepo          repository.OwnerRepository
	ownershipShareRepo repository.OwnershipShareRepository
	chargeRepo         repository.ChargeRepository
	paymentRepo        repository.PaymentRepository
}

// NewProcessImportBatchUseCase створює новий use case.
func NewProcessImportBatchUseCase(
	batchRepo repository.ImportBatchRepository,
	recordRepo repository.ImportedRecordRepository,
	apartmentRepo repository.ApartmentRepository,
	ownerRepo repository.OwnerRepository,
	ownershipShareRepo repository.OwnershipShareRepository,
	chargeRepo repository.ChargeRepository,
	paymentRepo repository.PaymentRepository,
) *ProcessImportBatchUseCase {
	return &ProcessImportBatchUseCase{
		batchRepo:          batchRepo,
		recordRepo:         recordRepo,
		apartmentRepo:      apartmentRepo,
		ownerRepo:          ownerRepo,
		ownershipShareRepo: ownershipShareRepo,
		chargeRepo:         chargeRepo,
		paymentRepo:        paymentRepo,
	}
}

// ProcessResult результат обробки батчу.
type ProcessResult struct {
	SuccessCount int
	FailCount    int
	Errors       []string
}

// Execute виконує обробку батчу.
func (uc *ProcessImportBatchUseCase) Execute(ctx context.Context, batchID int64) (*ProcessResult, error) {
	// 1. Отримуємо записи, які ще не мігровані
	records, err := uc.recordRepo.FindByBatchID(batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	result := &ProcessResult{
		Errors: make([]string, 0),
	}

	// Check if DB is empty (no owners and no charges)
	// If DB is not empty, we disable auto-creation of structural entities (Apartments, Owners)
	// to prevent polluting the database with duplicates or incorrect data during regular imports.
	ownerCount, err := uc.ownerRepo.Count(ctx, repository.OwnerFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to count owners: %w", err)
	}
	chargeCount, err := uc.chargeRepo.Count(ctx, repository.ChargeFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to count charges: %w", err)
	}

	allowCreation := ownerCount == 0 && chargeCount == 0

	if allowCreation {
		fmt.Println("Database is empty. Using MASS CREATION strategy (2-Pass).")
		// Pass 1: Analyze
		dtos := uc.analyzeBatch(records)
		// Pass 2: Execute
		return uc.executeMassCreation(ctx, dtos)
	} else {
		fmt.Println("Database is not empty. Using INCREMENTAL strategy (Restricted).")
		// Existing logic (Incremental)
		for _, record := range records {
			if record.IsMigrated {
				continue
			}

			if err := uc.processRecord(ctx, record, false); err != nil {
				// Зберігаємо помилку
				errMsg := fmt.Sprintf("Record %d (Apt %s): %v", record.ID, record.ApartmentNumber, err)
				fmt.Println(errMsg) // Log to stdout as well
				result.Errors = append(result.Errors, errMsg)
				result.FailCount++
			} else {
				result.SuccessCount++
			}
		}
		return result, nil
	}

}

func (uc *ProcessImportBatchUseCase) processRecord(ctx context.Context, record *entity.ImportedMonthlyRecord, allowCreation bool) error {
	// 1. Знайти квартиру
	apartment, err := uc.apartmentRepo.GetByNumber(ctx, record.ApartmentNumber)
	if err != nil {
		// Якщо квартира не знайдена
		if !allowCreation {
			return fmt.Errorf("apartment %s not found (auto-creation disabled)", record.ApartmentNumber)
		}

		// Створюємо квартиру

		// Parse cadastral number for entrance and floor
		var entrance *int
		var floor int
		if record.AccountNumber != nil {
			ent, apt, err := parseCadastralNumber(*record.AccountNumber)
			if err == nil {
				entrance = &ent
				// Calculate floor
				floor = calculateFloor(ent, apt)
			}
		}

		newApartment := &entity.Apartment{
			ApartmentNumber: record.ApartmentNumber,
			AreaTotal:       coalesceFloat(record.TotalArea, 0),
			AreaLiving:      nil,
			RoomsCount:      nil,
			Entrance:        entrance,
			Floor:           floor,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := uc.apartmentRepo.Create(ctx, newApartment); err != nil {
			return fmt.Errorf("failed to create apartment: %w", err)
		}
		apartment = newApartment

		// Check if owner already exists (e.g. same owner for multiple apartments in import)
		owner, err := uc.ownerRepo.GetByName(ctx, record.OwnerName)
		if err != nil {
			// Owner not found, create new
			owner, err = uc.createOwner(ctx, record.OwnerName)
			if err != nil {
				return fmt.Errorf("failed to create owner: %w", err)
			}
		}

		// Створюємо частку
		share, err := uc.createOwnershipShare(ctx, apartment.ID, owner.ID)
		if err != nil {
			return fmt.Errorf("failed to create share: %w", err)
		}

		return uc.createChargeAndPayment(ctx, record, apartment, owner, share)
	}

	// Квартира існує. Шукаємо власника за іменем.
	owner, err := uc.ownerRepo.GetByName(ctx, record.OwnerName)
	if err != nil {
		// Власника за іменем не знайдено.

		// Перевіряємо, чи є у квартири інші власники.
		shares, err := uc.ownershipShareRepo.GetActiveByApartment(ctx, apartment.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing shares: %w", err)
		}

		if len(shares) > 0 {
			// Квартира вже має власників, але ім'я в імпорті не співпадає.
			// Враховуючи вимогу про співвласників (напр. кв 5), ми додаємо нового власника як співвласника.
			// АЛЕ тільки якщо дозволено створення!
			if !allowCreation {
				// Якщо створення заборонено, ми не можемо додати нового співвласника автоматично.
				// Але тут є нюанс: якщо це "регулярний імпорт", можливо ми хочемо дозволити додавання співвласників?
				// Користувач сказав: "масове створення власників ... лише коли немає жодного запису".
				// Це суворо. Тому повертаємо помилку.
				return fmt.Errorf("owner '%s' not found for apartment %s (auto-creation disabled)",
					record.OwnerName, record.ApartmentNumber)
			}

			fmt.Printf("Adding new co-owner '%s' to apartment %s which already has %d owners\n",
				record.OwnerName, apartment.ApartmentNumber, len(shares))
		} else {
			// Власників немає взагалі.
			if !allowCreation {
				return fmt.Errorf("no owners found for apartment %s and auto-creation is disabled", record.ApartmentNumber)
			}
		}

		// Створюємо нового власника (як першого або як співвласника)
		owner, err = uc.createOwner(ctx, record.OwnerName)
		if err != nil {
			return fmt.Errorf("failed to create owner: %w", err)
		}

		// Створюємо частку
		share, err := uc.createOwnershipShare(ctx, apartment.ID, owner.ID)
		if err != nil {
			return fmt.Errorf("failed to create share: %w", err)
		}

		return uc.createChargeAndPayment(ctx, record, apartment, owner, share)
	}

	// Власник знайдений. Перевіряємо чи є у нього частка в цій квартирі.
	share, err := uc.ownershipShareRepo.GetByApartmentAndOwner(ctx, apartment.ID, owner.ID)
	if err != nil {
		// Частки немає - створюємо
		share, err = uc.createOwnershipShare(ctx, apartment.ID, owner.ID)
		if err != nil {
			return fmt.Errorf("failed to create share: %w", err)
		}
	}

	return uc.createChargeAndPayment(ctx, record, apartment, owner, share)
}

// Helper methods

func (uc *ProcessImportBatchUseCase) createOwner(ctx context.Context, name string) (*entity.Owner, error) {
	// Parse name: "Сидоренко М.А." -> Last: "Сидоренко", First: "М.xxx", Middle: "А.xxx"
	parts := strings.Fields(name)
	lastName := name
	var firstName string
	var middleName string

	if len(parts) > 0 {
		lastName = parts[0]
		if len(parts) > 1 {
			// Check if second part contains initials like "М.М." or "М.А."
			initials := parts[1]
			// Simple heuristic: if it contains dots, treat as initials/firstname placeholder
			if strings.Contains(initials, ".") {
				// Split by dot
				initialParts := strings.Split(initials, ".")

				// First initial
				if len(initialParts) > 0 && initialParts[0] != "" {
					firstName = initialParts[0] + ".xxx"
				}

				// Second initial (Middle Name)
				if len(initialParts) > 1 && initialParts[1] != "" {
					middleName = initialParts[1] + ".xxx"
				}
			} else {
				// No dots, assume full names
				firstName = parts[1]
				if len(parts) > 2 {
					middleName = parts[2]
				}
			}
		}
	}

	owner := &entity.Owner{
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: &middleName,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := uc.ownerRepo.Create(ctx, owner); err != nil {
		return nil, err
	}
	return owner, nil
}

func (uc *ProcessImportBatchUseCase) createOwnershipShare(ctx context.Context, apartmentID, ownerID int64) (*entity.OwnershipShare, error) {
	share := &entity.OwnershipShare{
		OwnerID:          ownerID,
		ApartmentID:      apartmentID,
		ShareNumerator:   100,
		ShareDenominator: 100,
		OwnershipType:    entity.OwnershipTypeFull,
		StartDate:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := uc.ownershipShareRepo.Create(ctx, share); err != nil {
		return nil, err
	}
	return share, nil
}

func (uc *ProcessImportBatchUseCase) createChargeAndPayment(
	ctx context.Context,
	record *entity.ImportedMonthlyRecord,
	apartment *entity.Apartment,
	owner *entity.Owner,
	share *entity.OwnershipShare,
) error {
	var chargeID *int64

	// 4. Створити нарахування (Charge)
	if record.ChargeAmount > 0 {
		charge, err := entity.NewCharge(
			share.ID,
			entity.ChargeTypeMaintenance,
			time.Now(),
			record.PeriodMonth,
			record.PeriodYear,
			record.ChargeAmount,
			nil, nil,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create charge entity: %w", err)
		}

		if err := uc.chargeRepo.Create(ctx, charge); err != nil {
			return fmt.Errorf("failed to save charge: %w", err)
		}
		chargeID = &charge.ID
	}

	// 5. Створити платіж (Payment) - REMOVED per user request
	// Payments should not be created from import records automatically.
	// if record.AmountPaid > 0 { ... }

	// 6. Оновити запис як мігрований
	record.MarkAsMigrated(&apartment.ID, &owner.ID, &share.ID, chargeID, nil)
	if err := uc.recordRepo.Update(record); err != nil {
		return fmt.Errorf("failed to update record status: %w", err)
	}

	return nil
}

func coalesceFloat(val *float64, def float64) float64 {
	if val != nil {
		return *val
	}
	return def
}

// Parsing Logic

func parseCadastralNumber(cadastral string) (int, int, error) {
	// Format: [Entrance][0...0][Apartment]
	// Example: 1005 -> Entrance 1, Apt 5
	// Example: 1005-a -> Entrance 1, Apt 5 (suffix ignored for parsing numbers)

	// Remove suffix if present
	clean := cadastral
	if idx := strings.Index(cadastral, "-"); idx != -1 {
		clean = cadastral[:idx]
	}

	// Parse as int
	val, err := strconv.Atoi(clean)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid cadastral number format: %s", cadastral)
	}

	// Heuristic: First digit is entrance?
	// If val is 1005, entrance is 1, apt is 5.
	// If val is 10005 (unlikely for 9 floors), entrance is 10?
	// Let's assume standard format where apartment is last 3 digits? Or just removing leading digit?
	// User said: "first digit - entrance, second digit - apartment, zeros between - filler"
	// This implies 1005 -> 1 (entrance), 5 (apartment).
	// 2025 -> 2 (entrance), 25 (apartment).

	s := strconv.Itoa(val)
	if len(s) < 2 {
		return 0, 0, fmt.Errorf("cadastral number too short: %s", cadastral)
	}

	entranceStr := s[:1]
	entrance, _ := strconv.Atoi(entranceStr)

	aptStr := s[1:]
	apartment, _ := strconv.Atoi(aptStr)

	return entrance, apartment, nil
}

func calculateFloor(entrance, apartment int) int {
	// Building: 9 floors
	// Standard: 4 apts per floor
	// Entrances 1-6
	// Exceptions:
	// Ent 4, Fl 1: 3 apts
	// Ent 6, Fl 1: 3 apts
	// Ent 6, Fl 2: 3 apts

	// We need to find which floor this apartment is on, RELATIVE to the entrance.
	// Usually apartment numbers are continuous across the building.
	// But here the input is just "Apartment 5" in "Entrance 1".
	// Wait, if cadastral is 1005, does it mean Apartment 5 GLOBALLY or Apartment 5 in Entrance 1?
	// Usually apartment numbers are unique per building.
	// If "1005" means Entrance 1, Apt 5, then Apt 5 is likely on Floor 2 (if 4 per floor).

	// Let's assume the apartment number extracted IS the global apartment number.
	// We need to check if this apartment number belongs to this entrance.
	// Actually, the user requirement says "first digit entrance, second digit apartment".
	// This strongly suggests the apartment number is extracted from the cadastral number.

	// Let's calculate floor based on apartment number and entrance configuration.
	// We need to know the starting apartment number for each entrance to be precise,
	// OR we assume the apartment number provided IS the global number and we just need to find its floor.
	// But the floor depends on the layout.

	// Let's implement a simulation:
	// Iterate through entrances 1 to 6.
	// For current entrance, iterate floors 1 to 9.
	// Determine apts on this floor.
	// Check if our apartment falls in this range.

	// BUT, we are given Entrance explicitly from Cadastral.
	// So we only need to simulate THIS entrance?
	// No, apartment numbers are usually continuous. Apt 1 is in Ent 1. Apt 36 is in Ent 1 (9*4=36). Apt 37 is in Ent 2.
	// If Cadastral is 2037 -> Entrance 2, Apt 37.

	// Let's simulate the whole building to find the floor.

	currentApt := 1
	for e := 1; e <= 6; e++ {
		for f := 1; f <= 9; f++ {
			aptsOnFloor := 4

			// Exceptions
			if e == 4 && f == 1 {
				aptsOnFloor = 3
			}
			if e == 6 && (f == 1 || f == 2) {
				aptsOnFloor = 3
			}

			// Check if our apartment is on this floor
			// Range: [currentApt, currentApt + aptsOnFloor - 1]
			if apartment >= currentApt && apartment < currentApt+aptsOnFloor {
				return f
			}

			currentApt += aptsOnFloor
		}
	}

	return 0 // Not found or out of range
}

// ============================================================================
// Mass Import Strategy (DTOs & Analysis)
// ============================================================================

type ApartmentImportDTO struct {
	ApartmentNumber string
	Entrance        *int
	Floor           int
	TotalArea       float64
	Owners          map[string]*OwnerImportDTO // Key: OwnerName
}

type OwnerImportDTO struct {
	RawName string
	Records []*entity.ImportedMonthlyRecord
}

func (uc *ProcessImportBatchUseCase) analyzeBatch(records []*entity.ImportedMonthlyRecord) map[string]*ApartmentImportDTO {
	dtos := make(map[string]*ApartmentImportDTO)

	for _, record := range records {
		dto, exists := dtos[record.ApartmentNumber]
		if !exists {
			// Parse entrance/floor
			var entrance *int
			var floor int
			if record.AccountNumber != nil {
				ent, apt, err := parseCadastralNumber(*record.AccountNumber)
				if err == nil {
					entrance = &ent
					floor = calculateFloor(ent, apt)
				}
			}

			dto = &ApartmentImportDTO{
				ApartmentNumber: record.ApartmentNumber,
				Entrance:        entrance,
				Floor:           floor,
				TotalArea:       coalesceFloat(record.TotalArea, 0),
				Owners:          make(map[string]*OwnerImportDTO),
			}
			dtos[record.ApartmentNumber] = dto
		}

		ownerDTO, ownerExists := dto.Owners[record.OwnerName]
		if !ownerExists {
			ownerDTO = &OwnerImportDTO{
				RawName: record.OwnerName,
				Records: make([]*entity.ImportedMonthlyRecord, 0),
			}
			dto.Owners[record.OwnerName] = ownerDTO
		}
		ownerDTO.Records = append(ownerDTO.Records, record)
	}

	return dtos
}

func (uc *ProcessImportBatchUseCase) executeMassCreation(ctx context.Context, dtos map[string]*ApartmentImportDTO) (*ProcessResult, error) {
	result := &ProcessResult{
		Errors: make([]string, 0),
	}

	// Cache for created owners to handle duplicates across apartments
	// Key: OwnerName, Value: Owner Entity
	ownerCache := make(map[string]*entity.Owner)

	for _, aptDTO := range dtos {
		// 1. Create or Get Apartment
		var apartment *entity.Apartment
		existingApartment, err := uc.apartmentRepo.GetByNumber(ctx, aptDTO.ApartmentNumber)
		if err == nil {
			// Apartment exists, use it
			apartment = existingApartment
		} else {
			// Apartment not found (or error), try to create
			newApartment := &entity.Apartment{
				ApartmentNumber: aptDTO.ApartmentNumber,
				AreaTotal:       aptDTO.TotalArea,
				AreaLiving:      nil,
				RoomsCount:      nil,
				Entrance:        aptDTO.Entrance,
				Floor:           aptDTO.Floor,
				IsActive:        true,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := uc.apartmentRepo.Create(ctx, newApartment); err != nil {
				errMsg := fmt.Sprintf("Failed to create apartment %s: %v", aptDTO.ApartmentNumber, err)
				result.Errors = append(result.Errors, errMsg)
				// Mark all records for this apartment as failed
				for _, ownerDTO := range aptDTO.Owners {
					result.FailCount += len(ownerDTO.Records)
				}
				continue
			}
			apartment = newApartment
		}

		// 2. Calculate Share
		totalOwners := len(aptDTO.Owners)
		shareNum := 1
		shareDenom := totalOwners
		if totalOwners == 0 {
			shareDenom = 1
		}

		// 3. Process Owners
		for _, ownerDTO := range aptDTO.Owners {
			// Check cache first
			owner, exists := ownerCache[ownerDTO.RawName]
			if !exists {
				// Check DB
				dbOwner, err := uc.ownerRepo.GetByName(ctx, ownerDTO.RawName)
				if err == nil {
					owner = dbOwner
				} else {
					// Create new
					owner, err = uc.createOwner(ctx, ownerDTO.RawName)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to create owner %s: %v", ownerDTO.RawName, err)
						result.Errors = append(result.Errors, errMsg)
						result.FailCount += len(ownerDTO.Records)
						continue
					}
				}
				// Add to cache
				ownerCache[ownerDTO.RawName] = owner
			}

			// Create Share
			share := &entity.OwnershipShare{
				OwnerID:          owner.ID,
				ApartmentID:      apartment.ID,
				ShareNumerator:   shareNum,
				ShareDenominator: shareDenom,
				OwnershipType:    entity.OwnershipTypeFull,
				StartDate:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				IsActive:         true,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if shareDenom > 1 {
				share.OwnershipType = entity.OwnershipTypeShared
			}

			if err := uc.ownershipShareRepo.Create(ctx, share); err != nil {
				errMsg := fmt.Sprintf("Failed to create share for %s in apt %s: %v", ownerDTO.RawName, aptDTO.ApartmentNumber, err)
				result.Errors = append(result.Errors, errMsg)
				result.FailCount += len(ownerDTO.Records)
				continue
			}

			// 4. Create Charges for each record
			for _, record := range ownerDTO.Records {
				if err := uc.createChargeAndPayment(ctx, record, apartment, owner, share); err != nil {
					errMsg := fmt.Sprintf("Record %d (Apt %s): %v", record.ID, record.ApartmentNumber, err)
					result.Errors = append(result.Errors, errMsg)
					result.FailCount++
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	return result, nil
}
