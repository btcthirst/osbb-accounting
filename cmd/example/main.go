// cmd/example/main.go
package main

import (
	"context"
	"log"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/auth"
	"osbb-accounting/application/usecase/osbb"
	"osbb-accounting/application/usecase/owner"
	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/security"
)

func main() {
	ctx := context.Background()

	// ========================================================================
	// 1. ІНІЦІАЛІЗАЦІЯ
	// ========================================================================

	// Підключення до БД
	db, err := sqlite.Connect(sqlite.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer sqlite.Close(db)

	// Repositories
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	roleRepo := sqlite.NewRoleRepository(db)
	permissionRepo := sqlite.NewPermissionRepository(db)
	osbbRepo := sqlite.NewOSBBRepository(db)
	ownerRepo := sqlite.NewOwnerRepository(db)

	// Security
	passwordHasher := security.DefaultBCryptHasher()

	// Services
	authService := service.NewAuthService(
		userRepo, sessionRepo, roleRepo, permissionRepo, passwordHasher,
	)
	osbbService := service.NewOSBBService(osbbRepo, permissionRepo)
	ownerService := service.NewOwnerService(ownerRepo, permissionRepo)

	// ========================================================================
	// 2. СТВОРЕННЯ ADMIN КОРИСТУВАЧА (якщо не існує)
	// ========================================================================

	_, err = userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		// Користувача не існує - створюємо
		log.Println("Creating admin user...")

		registerOutput, err := authService.Register(ctx, auth.RegisterUserInput{
			Username:  "admin",
			Email:     "admin@osbb.local",
			Password:  "Admin123!",
			FirstName: "Адміністратор",
			LastName:  "Системи",
			RoleName:  "admin",
		})
		if err != nil {
			log.Fatal("Failed to create admin:", err)
		}

		log.Printf("Admin user created: %s (ID: %d)\n",
			registerOutput.Username, registerOutput.UserID)
	}

	// Вхід як admin
	loginOutput, err := authService.Login(ctx, auth.LoginUserInput{
		UsernameOrEmail: "admin",
		Password:        "Admin123!",
	})
	if err != nil {
		log.Fatal("Failed to login:", err)
	}

	currentUserID := loginOutput.User.ID
	log.Printf("Logged in as: %s\n", loginOutput.User.FullName)

	// ========================================================================
	// 3. СТВОРЕННЯ/ОТРИМАННЯ ОСББ
	// ========================================================================

	// Спробуємо отримати ОСББ
	osbbData, err := osbbService.Get(ctx, osbb.GetOSBBInput{
		CurrentUserID: currentUserID,
	})

	if err != nil {
		// ОСББ не існує - створюємо
		log.Println("Creating OSBB organization...")

		phone := "+380442345678"
		email := "info@osbb-example.com"
		website := "https://osbb-example.com"

		osbbData, err = osbbService.Create(ctx, osbb.CreateOSBBInput{
			CurrentUserID: currentUserID,
			Name:          "ОСББ 'Сонячний'",
			EDRPOU:        "12345678",
			LegalAddress:  "м. Київ, вул. Центральна, 1",
			Phone:         &phone,
			Email:         &email,
			Website:       &website,
			ChairmanName:  "Іванов Іван Іванович",
		})
		if err != nil {
			log.Fatal("Failed to create OSBB:", err)
		}

		log.Printf("OSBB created: %s (ЄДРПОУ: %s)\n", osbbData.Name, osbbData.EDRPOU)
	} else {
		log.Printf("OSBB loaded: %s (ЄДРПОУ: %s)\n", osbbData.Name, osbbData.EDRPOU)
	}

	// ========================================================================
	// 4. СТВОРЕННЯ ВЛАСНИКІВ
	// ========================================================================

	log.Println("\n=== Creating Owners ===")

	// Власник 1
	phone1 := "+380501234567"
	email1 := "petrenko@example.com"
	taxNumber1 := "1234567890"
	middleName1 := "Олександрович"

	owner1, err := ownerService.Create(ctx, owner.CreateOwnerInput{
		CurrentUserID: currentUserID,
		FirstName:     "Іван",
		LastName:      "Петренко",
		MiddleName:    &middleName1,
		Phone:         &phone1,
		Email:         &email1,
		TaxNumber:     &taxNumber1,
	})
	if err != nil {
		log.Printf("Failed to create owner 1: %v\n", err)
	} else {
		log.Printf("Owner 1 created: %s (ID: %d)\n", owner1.FullName, owner1.ID)
	}

	// Власник 2
	phone2 := "+380509876543"
	email2 := "kovalenko@example.com"
	taxNumber2 := "9876543210"
	middleName2 := "Петрівна"

	owner2, err := ownerService.Create(ctx, owner.CreateOwnerInput{
		CurrentUserID: currentUserID,
		FirstName:     "Марія",
		LastName:      "Коваленко",
		MiddleName:    &middleName2,
		Phone:         &phone2,
		Email:         &email2,
		TaxNumber:     &taxNumber2,
	})
	if err != nil {
		log.Printf("Failed to create owner 2: %v\n", err)
	} else {
		log.Printf("Owner 2 created: %s (ID: %d)\n", owner2.FullName, owner2.ID)
	}

	// Власник 3 (без ІПН та email)
	phone3 := "+380631112233"

	owner3, err := ownerService.Create(ctx, owner.CreateOwnerInput{
		CurrentUserID: currentUserID,
		FirstName:     "Олександр",
		LastName:      "Сидоренко",
		Phone:         &phone3,
	})
	if err != nil {
		log.Printf("Failed to create owner 3: %v\n", err)
	} else {
		log.Printf("Owner 3 created: %s (ID: %d)\n", owner3.FullName, owner3.ID)
	}

	// ========================================================================
	// 5. ОТРИМАННЯ СПИСКУ ВЛАСНИКІВ
	// ========================================================================

	log.Println("\n=== Listing All Owners ===")

	listOutput, err := ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: currentUserID,
		Limit:         10,
		Offset:        0,
		OrderBy:       "name",
	})
	if err != nil {
		log.Fatal("Failed to list owners:", err)
	}

	log.Printf("Total owners: %d\n", listOutput.Total)
	for i, o := range listOutput.Owners {
		contact := "No contact"
		if o.Phone != nil || o.Email != nil {
			contact = ""
			if o.Phone != nil {
				contact += *o.Phone
			}
			if o.Email != nil {
				if contact != "" {
					contact += ", "
				}
				contact += *o.Email
			}
		}

		taxInfo := "No tax number"
		if o.TaxNumber != nil {
			taxInfo = "ІПН: " + *o.TaxNumber
		}

		log.Printf("%d. %s | %s | %s\n",
			i+1, o.FullName, contact, taxInfo)
	}

	// ========================================================================
	// 6. ПОШУК ВЛАСНИКІВ
	// ========================================================================

	log.Println("\n=== Searching Owners ===")

	searchResults, err := ownerService.Search(ctx, owner.SearchOwnersInput{
		CurrentUserID: currentUserID,
		Query:         "Кова",
		Limit:         5,
	})
	if err != nil {
		log.Fatal("Failed to search owners:", err)
	}

	log.Printf("Found %d owners matching 'Кова':\n", len(searchResults))
	for i, o := range searchResults {
		log.Printf("%d. %s (%s)\n", i+1, o.FullName, o.ShortName)
	}

	// ========================================================================
	// 7. ОНОВЛЕННЯ ВЛАСНИКА
	// ========================================================================

	if owner1 != nil {
		log.Println("\n=== Updating Owner ===")

		newPhone := "+380509999999"
		newEmail := "petrenko.new@example.com"
		notes := "Власник квартири №15, 3-й поверх"

		updatedOwner, err := ownerService.Update(ctx, owner.UpdateOwnerInput{
			CurrentUserID: currentUserID,
			OwnerID:       owner1.ID,
			FirstName:     owner1.FirstName,
			LastName:      owner1.LastName,
			MiddleName:    owner1.MiddleName,
			Phone:         &newPhone,
			Email:         &newEmail,
			TaxNumber:     owner1.TaxNumber,
			Notes:         &notes,
		})
		if err != nil {
			log.Printf("Failed to update owner: %v\n", err)
		} else {
			log.Printf("Owner updated: %s\n", updatedOwner.FullName)
			log.Printf("  New phone: %s\n", *updatedOwner.Phone)
			log.Printf("  New email: %s\n", *updatedOwner.Email)
		}
	}

	// ========================================================================
	// 8. ФІЛЬТРАЦІЯ ВЛАСНИКІВ
	// ========================================================================

	log.Println("\n=== Filtering Owners ===")

	// Тільки з ІПН
	hasTaxNumber := true
	filteredOutput, err := ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: currentUserID,
		HasTaxNumber:  &hasTaxNumber,
		Limit:         10,
	})
	if err != nil {
		log.Fatal("Failed to filter owners:", err)
	}

	log.Printf("Owners with tax number: %d\n", filteredOutput.Total)
	for _, o := range filteredOutput.Owners {
		if o.TaxNumber != nil {
			log.Printf("  - %s (ІПН: %s)\n", o.FullName, *o.TaxNumber)
		}
	}

	// Тільки з контактами
	hasContact := true
	contactsOutput, err := ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: currentUserID,
		HasContact:    &hasContact,
		Limit:         10,
	})
	if err != nil {
		log.Fatal("Failed to filter owners:", err)
	}

	log.Printf("\nOwners with contacts: %d\n", contactsOutput.Total)
	for _, o := range contactsOutput.Owners {
		log.Printf("  - %s\n", o.FullName)
	}

	// ========================================================================
	// 9. ВИДАЛЕННЯ ВЛАСНИКА (SOFT DELETE)
	// ========================================================================

	if owner3 != nil {
		log.Println("\n=== Deleting Owner ===")

		deleteOutput, err := ownerService.Delete(ctx, owner.DeleteOwnerInput{
			CurrentUserID: currentUserID,
			OwnerID:       owner3.ID,
		})
		if err != nil {
			log.Printf("Failed to delete owner: %v\n", err)
		} else {
			log.Printf("Delete result: %s\n", deleteOutput.Message)
		}

		// Перевірка - owner3 не повинен з'являтися в списку
		afterDeleteOutput, err := ownerService.List(ctx, owner.ListOwnersInput{
			CurrentUserID: currentUserID,
			Limit:         10,
		})
		if err == nil {
			log.Printf("Owners count after delete: %d (was %d)\n",
				afterDeleteOutput.Total, listOutput.Total)
		}
	}

	// ========================================================================
	// 10. СТАТИСТИКА
	// ========================================================================

	log.Println("\n=== Statistics ===")

	allOwners, _ := ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: currentUserID,
		Limit:         1000,
	})

	ownersWithTax, _ := ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: currentUserID,
		HasTaxNumber:  &hasTaxNumber,
		Limit:         1000,
	})

	ownersWithContacts, _ := ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: currentUserID,
		HasContact:    &hasContact,
		Limit:         1000,
	})

	log.Printf("Total owners: %d\n", allOwners.Total)
	log.Printf("Owners with tax number: %d (%.1f%%)\n",
		ownersWithTax.Total,
		float64(ownersWithTax.Total)/float64(allOwners.Total)*100)
	log.Printf("Owners with contacts: %d (%.1f%%)\n",
		ownersWithContacts.Total,
		float64(ownersWithContacts.Total)/float64(allOwners.Total)*100)

	log.Println("\n=== Done ===")
}

// Helper function
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
