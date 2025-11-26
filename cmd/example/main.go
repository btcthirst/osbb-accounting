// cmd/example/main.go
package main

import (
	"context"
	"log"
	"time"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/application/usecase/auth"

	// "osbb-accounting/application/usecase/charge"
	"osbb-accounting/application/usecase/contractor"
	"osbb-accounting/application/usecase/expense"
	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/application/usecase/osbb"
	"osbb-accounting/application/usecase/owner"
	"osbb-accounting/application/usecase/ownership"

	// "osbb-accounting/application/usecase/payment"
	"osbb-accounting/config"
	"osbb-accounting/domain/entity"
	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/security"
)

func main() {
	ctx := context.Background()

	// ========================================================================
	// 1. ІНІЦІАЛІЗАЦІЯ
	// ========================================================================

	// Підключення до БД
	// 0. Завантаження конфігурації
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("Config file not found or invalid, using defaults: %v", err)
		cfg = config.Default()
	}
	dbConfig := &sqlite.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	}
	db, err := sqlite.Connect(dbConfig)
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
	apartmentRepo := sqlite.NewApartmentRepository(db)
	ownershipShareRepo := sqlite.NewOwnershipShareRepository(db)
	// chargeRepo := sqlite.NewChargeRepository(db)
	// paymentRepo := sqlite.NewPaymentRepository(db)
	expenseRepo := sqlite.NewExpenseRepository(db)
	expenseCategoryRepo := sqlite.NewExpenseCategoryRepository(db)
	contractorRepo := sqlite.NewContractorRepository(db)

	// Security
	passwordHasher := security.DefaultBCryptHasher()

	// Services
	authService := service.NewAuthService(
		userRepo, sessionRepo, roleRepo, permissionRepo, passwordHasher,
	)
	osbbService := service.NewOSBBService(osbbRepo, permissionRepo)
	ownerService := service.NewOwnerService(ownerRepo, permissionRepo)
	apartmentService := service.NewApartmentService(apartmentRepo, permissionRepo)
	ownershipService := service.NewOwnershipService(ownershipShareRepo, ownerRepo, apartmentRepo, permissionRepo)
	// chargeService := service.NewChargeService(chargeRepo, permissionRepo)
	// paymentService := service.NewPaymentService(paymentRepo, permissionRepo)
	expenseService := service.NewExpenseService(expenseRepo, permissionRepo, expenseCategoryRepo)
	expenseCategoryService := service.NewExpenseCategoryService(expenseCategoryRepo, permissionRepo)
	contractorService := service.NewContractorService(contractorRepo, permissionRepo)

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
	// 10. СТВОРЕННЯ КВАРТИР
	// ========================================================================

	log.Println("\n=== Creating Apartments ===")

	entrance1 := 1
	rooms1 := 2
	apt1, err := apartmentService.Create(ctx, apartment.CreateApartmentInput{
		CurrentUserID:   currentUserID,
		ApartmentNumber: "1",
		Floor:           1,
		Entrance:        &entrance1,
		AreaTotal:       50.5,
		AreaLiving:      float64Ptr(30.0),
		RoomsCount:      &rooms1,
	})
	if err != nil {
		log.Printf("Failed to create apartment 1: %v\n", err)
	} else {
		log.Printf("Apartment 1 created: %s (ID: %d)\n", apt1.ApartmentNumber, apt1.ID)
	}

	entrance2 := 1
	rooms2 := 3
	apt2, err := apartmentService.Create(ctx, apartment.CreateApartmentInput{
		CurrentUserID:   currentUserID,
		ApartmentNumber: "2",
		Floor:           1,
		Entrance:        &entrance2,
		AreaTotal:       75.0,
		AreaLiving:      float64Ptr(45.0),
		RoomsCount:      &rooms2,
	})
	if err != nil {
		log.Printf("Failed to create apartment 2: %v\n", err)
	} else {
		log.Printf("Apartment 2 created: %s (ID: %d)\n", apt2.ApartmentNumber, apt2.ID)
	}

	// Квартира 3 - на 2 поверсі
	entrance3 := 1
	rooms3 := 1
	apt3, err := apartmentService.Create(ctx, apartment.CreateApartmentInput{
		CurrentUserID:   currentUserID,
		ApartmentNumber: "3",
		Floor:           2,
		Entrance:        &entrance3,
		AreaTotal:       35.0,
		AreaLiving:      float64Ptr(20.0),
		RoomsCount:      &rooms3,
	})
	if err != nil {
		log.Printf("Failed to create apartment 3: %v\n", err)
	} else {
		log.Printf("Apartment 3 created: %s (ID: %d)\n", apt3.ApartmentNumber, apt3.ID)
	}

	// Квартира 4 - на 2 поверсі в іншому під'їзді
	entrance4 := 2
	rooms4 := 3
	cadastral4 := "UA12345678901234567"
	apt4, err := apartmentService.Create(ctx, apartment.CreateApartmentInput{
		CurrentUserID:   currentUserID,
		ApartmentNumber: "4",
		Floor:           2,
		Entrance:        &entrance4,
		AreaTotal:       68.5,
		AreaLiving:      float64Ptr(42.0),
		RoomsCount:      &rooms4,
		CadastralNumber: &cadastral4,
	})
	if err != nil {
		log.Printf("Failed to create apartment 4: %v\n", err)
	} else {
		log.Printf("Apartment 4 created: %s (ID: %d)\n", apt4.ApartmentNumber, apt4.ID)
	}

	// Квартира 5 - на 3 поверсі, без кімнат (комерційне приміщення)
	entrance5 := 1
	apt5, err := apartmentService.Create(ctx, apartment.CreateApartmentInput{
		CurrentUserID:   currentUserID,
		ApartmentNumber: "5",
		Floor:           3,
		Entrance:        &entrance5,
		AreaTotal:       120.0,
		AreaLiving:      nil, // Комерційне приміщення, без житлової площі
		RoomsCount:      nil,
	})
	if err != nil {
		log.Printf("Failed to create apartment 5: %v\n", err)
	} else {
		log.Printf("Apartment 5 created: %s (ID: %d)\n", apt5.ApartmentNumber, apt5.ID)
	}

	// ========================================================================
	// 11. СТВОРЕННЯ ЧАСТОК ВЛАСНОСТІ
	// ========================================================================

	log.Println("\n=== Creating Ownership Shares ===")

	if owner1 != nil && apt1 != nil {
		share1, err := ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
			CurrentUserID:    currentUserID,
			OwnerID:          owner1.ID,
			ApartmentID:      apt1.ID,
			ShareNumerator:   1,
			ShareDenominator: 1,
			OwnershipType:    entity.OwnershipTypeFull,
			StartDate:        time.Now().AddDate(-1, 0, 0), // Рік тому
		})
		if err != nil {
			log.Printf("Failed to create share for owner 1: %v\n", err)
		} else {
			log.Printf("Share created for Owner 1 in Apt 1: %.0f%% (ID: %d)\n", share1.SharePercentage, share1.ID)
		}
	}

	if owner2 != nil && apt2 != nil {
		share2, err := ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
			CurrentUserID:    currentUserID,
			OwnerID:          owner2.ID,
			ApartmentID:      apt2.ID,
			ShareNumerator:   1,
			ShareDenominator: 2,
			OwnershipType:    entity.OwnershipTypeShared,
			StartDate:        time.Now().AddDate(0, -6, 0), // Півроку тому
		})
		if err != nil {
			log.Printf("Failed to create share for owner 2: %v\n", err)
		} else {
			log.Printf("Share created for Owner 2 in Apt 2: %.0f%% (ID: %d)\n", share2.SharePercentage, share2.ID)
		}
	}

	// Частка для квартири 3 - власник 1 здає в оренду
	if owner1 != nil && apt3 != nil {
		share3, err := ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
			CurrentUserID:    currentUserID,
			OwnerID:          owner1.ID,
			ApartmentID:      apt3.ID,
			ShareNumerator:   1,
			ShareDenominator: 1,
			OwnershipType:    entity.OwnershipTypeRent,
			StartDate:        time.Now().AddDate(0, -3, 0), // 3 місяці тому
		})
		if err != nil {
			log.Printf("Failed to create share for owner 1 (apt 3): %v\n", err)
		} else {
			log.Printf("Share created for Owner 1 in Apt 3 (rent): %.0f%% (ID: %d)\n", share3.SharePercentage, share3.ID)
		}
	}

	// Частка для квартири 4 - власник 2 (50%)
	if owner2 != nil && apt4 != nil {
		share4, err := ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
			CurrentUserID:    currentUserID,
			OwnerID:          owner2.ID,
			ApartmentID:      apt4.ID,
			ShareNumerator:   1,
			ShareDenominator: 2,
			OwnershipType:    entity.OwnershipTypeShared,
			StartDate:        time.Now().AddDate(-2, 0, 0), // 2 роки тому
		})
		if err != nil {
			log.Printf("Failed to create share for owner 2 (apt 4, 50%%): %v\n", err)
		} else {
			log.Printf("Share created for Owner 2 in Apt 4: %.0f%% (ID: %d)\n", share4.SharePercentage, share4.ID)
		}

		// Друга частка для квартири 4 - власник 1 (50%)
		if owner1 != nil {
			share5, err := ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
				CurrentUserID:    currentUserID,
				OwnerID:          owner1.ID,
				ApartmentID:      apt4.ID,
				ShareNumerator:   1,
				ShareDenominator: 2,
				OwnershipType:    entity.OwnershipTypeShared,
				StartDate:        time.Now().AddDate(-2, 0, 0), // 2 роки тому
			})
			if err != nil {
				log.Printf("Failed to create share for owner 1 (apt 4, 50%%): %v\n", err)
			} else {
				log.Printf("Share created for Owner 1 in Apt 4: %.0f%% (ID: %d)\n", share5.SharePercentage, share5.ID)
			}
		}
	}

	// Частка для квартири 5 (комерційне приміщення) - власник 2
	if owner2 != nil && apt5 != nil {
		share6, err := ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
			CurrentUserID:    currentUserID,
			OwnerID:          owner2.ID,
			ApartmentID:      apt5.ID,
			ShareNumerator:   1,
			ShareDenominator: 1,
			OwnershipType:    entity.OwnershipTypeFull,
			StartDate:        time.Now().AddDate(-3, 0, 0), // 3 роки тому
		})
		if err != nil {
			log.Printf("Failed to create share for owner 2 (apt 5): %v\n", err)
		} else {
			log.Printf("Share created for Owner 2 in Apt 5 (commercial): %.0f%% (ID: %d)\n", share6.SharePercentage, share6.ID)
		}
	}

	// ========================================================================
	// 12. СТВОРЕННЯ КОНТРАГЕНТІВ
	// ========================================================================

	log.Println("\n=== Creating Contractors ===")

	contractor1, err := contractorService.Create(ctx, contractor.CreateContractorInput{
		CurrentUserID:  currentUserID,
		Name:           "Київенерго",
		EDRPOU:         strPtr("12345678"),
		ContractorType: entity.ContractorTypeUtility,
		Phone:          strPtr("+380441234567"),
	})
	if err != nil {
		log.Printf("Failed to create contractor 1: %v\n", err)
	} else {
		log.Printf("Contractor 1 created: %s (ID: %d)\n", contractor1.Name, contractor1.ID)
	}

	contractor2, err := contractorService.Create(ctx, contractor.CreateContractorInput{
		CurrentUserID:  currentUserID,
		Name:           "ФОП Ремонтник",
		EDRPOU:         strPtr("87654321"),
		ContractorType: entity.ContractorTypeService,
		Phone:          strPtr("+380505556677"),
	})
	if err != nil {
		log.Printf("Failed to create contractor 2: %v\n", err)
	} else {
		log.Printf("Contractor 2 created: %s (ID: %d)\n", contractor2.Name, contractor2.ID)
	}

	contractor3, err := contractorService.Create(ctx, contractor.CreateContractorInput{
		CurrentUserID:  currentUserID,
		Name:           "Київводоканал",
		EDRPOU:         strPtr("11112222"),
		ContractorType: entity.ContractorTypeUtility,
		Phone:          strPtr("+380449876543"),
	})
	if err != nil {
		log.Printf("Failed to create contractor 3: %v\n", err)
	} else {
		log.Printf("Contractor 3 created: %s (ID: %d)\n", contractor3.Name, contractor3.ID)
	}

	contractor4, err := contractorService.Create(ctx, contractor.CreateContractorInput{
		CurrentUserID:  currentUserID,
		Name:           "ТОВ 'Чистоту місту'",
		EDRPOU:         strPtr("33334444"),
		ContractorType: entity.ContractorTypeUtility,
		Phone:          strPtr("+380501112233"),
	})
	if err != nil {
		log.Printf("Failed to create contractor 4: %v\n", err)
	} else {
		log.Printf("Contractor 4 created: %s (ID: %d)\n", contractor4.Name, contractor4.ID)
	}

	contractor5, err := contractorService.Create(ctx, contractor.CreateContractorInput{
		CurrentUserID:  currentUserID,
		Name:           "ПП 'Управління будинком'",
		EDRPOU:         strPtr("55556666"),
		ContractorType: entity.ContractorTypeOther,
		Phone:          strPtr("+380445557788"),
	})
	if err != nil {
		log.Printf("Failed to create contractor 5: %v\n", err)
	} else {
		log.Printf("Contractor 5 created: %s (ID: %d)\n", contractor5.Name, contractor5.ID)
	}

	// ========================================================================
	// 13. СТВОРЕННЯ КАТЕГОРІЙ ВИТРАТ
	// ========================================================================

	log.Println("\n=== Creating Expense Categories ===")

	cat1, err := expenseCategoryService.Create(ctx, expense_category.CreateExpenseCategoryInput{
		CurrentUserID: currentUserID,
		Name:          "Комунальні послуги",
		CategoryType:  entity.CategoryTypeUtility,
		Description:   strPtr("Оплата за світло, воду, тепло"),
	})
	if err != nil {
		log.Printf("Failed to create category 1: %v\n", err)
	} else {
		log.Printf("Category 1 created: %s (ID: %d)\n", cat1.Name, cat1.ID)
	}

	cat2, err := expenseCategoryService.Create(ctx, expense_category.CreateExpenseCategoryInput{
		CurrentUserID: currentUserID,
		Name:          "Ремонтний фонд",
		CategoryType:  entity.CategoryTypeRepair,
		Description:   strPtr("Витрати на поточний ремонт"),
	})
	if err != nil {
		log.Printf("Failed to create category 2: %v\n", err)
	} else {
		log.Printf("Category 2 created: %s (ID: %d)\n", cat2.Name, cat2.ID)
	}

	cat3, err := expenseCategoryService.Create(ctx, expense_category.CreateExpenseCategoryInput{
		CurrentUserID: currentUserID,
		Name:          "Вивіз сміття",
		CategoryType:  entity.CategoryTypeUtility,
		Description:   strPtr("Послуги з вивозу твердих побутових відходів"),
	})
	if err != nil {
		log.Printf("Failed to create category 3: %v\n", err)
	} else {
		log.Printf("Category 3 created: %s (ID: %d)\n", cat3.Name, cat3.ID)
	}

	cat4, err := expenseCategoryService.Create(ctx, expense_category.CreateExpenseCategoryInput{
		CurrentUserID: currentUserID,
		Name:          "Утримання будинку",
		CategoryType:  entity.CategoryTypeRepair,
		Description:   strPtr("Загальні витрати на утримання будинку та прибудинкової території"),
	})
	if err != nil {
		log.Printf("Failed to create category 4: %v\n", err)
	} else {
		log.Printf("Category 4 created: %s (ID: %d)\n", cat4.Name, cat4.ID)
	}

	cat5, err := expenseCategoryService.Create(ctx, expense_category.CreateExpenseCategoryInput{
		CurrentUserID: currentUserID,
		Name:          "Адміністративні витрати",
		CategoryType:  entity.CategoryTypeOther,
		Description:   strPtr("Канцтовари, зв'язок, інші адміністративні витрати"),
	})
	if err != nil {
		log.Printf("Failed to create category 5: %v\n", err)
	} else {
		log.Printf("Category 5 created: %s (ID: %d)\n", cat5.Name, cat5.ID)
	}

	// ========================================================================
	// 14. СТВОРЕННЯ ВИТРАТ
	// ========================================================================

	log.Println("\n=== Creating Expenses ===")

	if cat1 != nil && contractor1 != nil {
		exp1, err := expenseService.Create(ctx, expense.CreateExpenseInput{
			CurrentUserID: currentUserID,
			CategoryID:    cat1.ID,
			ContractorID:  &contractor1.ID,
			ExpenseDate:   time.Now(),
			Amount:        5000.00,
			Description:   "Оплата за електроенергію (Жовтень)",
		})
		if err != nil {
			log.Printf("Failed to create expense 1: %v\n", err)
		} else {
			log.Printf("Expense 1 created: %.2f грн (ID: %d)\n", exp1.Amount, exp1.ID)
		}
	}

	// Витрата 2 - вода
	if cat1 != nil && contractor3 != nil {
		exp2, err := expenseService.Create(ctx, expense.CreateExpenseInput{
			CurrentUserID: currentUserID,
			CategoryID:    cat1.ID,
			ContractorID:  &contractor3.ID,
			ExpenseDate:   time.Now().AddDate(0, 0, -5),
			Amount:        3500.00,
			Description:   "Оплата за водопостачання та водовідведення (Жовтень)",
		})
		if err != nil {
			log.Printf("Failed to create expense 2: %v\n", err)
		} else {
			log.Printf("Expense 2 created: %.2f грн (ID: %d)\n", exp2.Amount, exp2.ID)
		}
	}

	// Витрата 3 - ремонт
	if cat2 != nil && contractor2 != nil {
		exp3, err := expenseService.Create(ctx, expense.CreateExpenseInput{
			CurrentUserID: currentUserID,
			CategoryID:    cat2.ID,
			ContractorID:  &contractor2.ID,
			ExpenseDate:   time.Now().AddDate(0, 0, -15),
			Amount:        12000.00,
			Description:   "Ремонт під'їзду №1",
		})
		if err != nil {
			log.Printf("Failed to create expense 3: %v\n", err)
		} else {
			log.Printf("Expense 3 created: %.2f грн (ID: %d)\n", exp3.Amount, exp3.ID)
		}
	}

	// Витрата 4 - вивіз сміття
	if cat3 != nil && contractor4 != nil {
		exp4, err := expenseService.Create(ctx, expense.CreateExpenseInput{
			CurrentUserID: currentUserID,
			CategoryID:    cat3.ID,
			ContractorID:  &contractor4.ID,
			ExpenseDate:   time.Now().AddDate(0, 0, -1),
			Amount:        4200.00,
			Description:   "Послуги з вивозу сміття (Листопад)",
		})
		if err != nil {
			log.Printf("Failed to create expense 4: %v\n", err)
		} else {
			log.Printf("Expense 4 created: %.2f грн (ID: %d)\n", exp4.Amount, exp4.ID)
		}
	}

	// Витрата 5 - утримання будинку
	if cat4 != nil && contractor5 != nil {
		exp5, err := expenseService.Create(ctx, expense.CreateExpenseInput{
			CurrentUserID: currentUserID,
			CategoryID:    cat4.ID,
			ContractorID:  &contractor5.ID,
			ExpenseDate:   time.Now(),
			Amount:        8500.00,
			Description:   "Послуги з управління будинком (Листопад)",
		})
		if err != nil {
			log.Printf("Failed to create expense 5: %v\n", err)
		} else {
			log.Printf("Expense 5 created: %.2f грн (ID: %d)\n", exp5.Amount, exp5.ID)
		}
	}

	// Витрата 6 - адміністративні витрати
	if cat5 != nil {
		exp6, err := expenseService.Create(ctx, expense.CreateExpenseInput{
			CurrentUserID: currentUserID,
			CategoryID:    cat5.ID,
			ContractorID:  nil, // Без контрагента
			ExpenseDate:   time.Now().AddDate(0, 0, -10),
			Amount:        1500.00,
			Description:   "Канцтовари та матеріали для офісу",
		})
		if err != nil {
			log.Printf("Failed to create expense 6: %v\n", err)
		} else {
			log.Printf("Expense 6 created: %.2f грн (ID: %d)\n", exp6.Amount, exp6.ID)
		}
	}

	// ========================================================================
	// 15. СТВОРЕННЯ НАРАХУВАНЬ
	// ========================================================================

	log.Println("\n=== Creating Charges ===")

	// TODO: Додати нарахування використовуючи OwnershipShareID
	// Приклад:
	// if owner1 != nil && apt1 != nil && share1 != nil {
	// 	charge1, err := chargeService.Create(ctx, charge.CreateChargeInput{
	// 		CurrentUserID:    currentUserID,
	// 		OwnershipShareID: share1.ID,  // Використовуй ID частки власності!
	// 		ChargeType:       entity.ChargeTypeMaintenance,
	// 		ChargeDate:       time.Now(),
	// 		Amount:           1000.00,
	// 		Description:      strPtr("Внесок за утримання будинку (Листопад)"),
	// 		PeriodMonth:      int(time.Now().Month()),
	// 		PeriodYear:       time.Now().Year(),
	// 	})
	// }

	// ========================================================================
	// 16. СТВОРЕННЯ ПЛАТЕЖІВ
	// ========================================================================

	log.Println("\n=== Creating Payments ===")

	// TODO: Додати платежі використовуючи OwnershipShareID та PaymentMethod
	// Приклад:
	// if owner1 != nil && share1 != nil {
	// 	pay1, err := paymentService.Create(ctx, payment.CreatePaymentInput{
	// 		CurrentUserID:    currentUserID,
	// 		OwnershipShareID: share1.ID,  // Використовуй ID частки власності!
	// 		Amount:           1000.00,
	// 		PaymentMethod:    entity.PaymentMethodBankTransfer,
	// 		PaymentPurpose:   "Оплата внеску за Nov",
	//		PaymentDate:      time.Now(),
	// 	})
	// }

	// ========================================================================
	// 17. СТАТИСТИКА
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

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
