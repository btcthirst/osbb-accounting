// cmd/main.go
package main

import (
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"osbb-accounting/application/service"
	authUseCase "osbb-accounting/application/usecase/auth"
	"osbb-accounting/config"
	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/security"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/screens"
	appTheme "osbb-accounting/presentation/fyne/theme"
)

func main() {
	// 0. Завантаження конфігурації
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("Config file not found or invalid, using defaults: %v", err)
		cfg = config.Default()
	}

	// 1. Ініціалізація Fyne додатку
	myApp := app.NewWithID("com.osbb.management")
	myApp.Settings().SetTheme(&appTheme.CustomTheme{})

	mainWindow := myApp.NewWindow(cfg.App.WindowTitle)
	mainWindow.Resize(fyne.NewSize(1200, 800))
	mainWindow.CenterOnScreen()

	// 2. Ініціалізація бази даних
	dbConfig := &sqlite.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	}

	db, err := sqlite.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlite.Close(db)

	// 3. Створення repositories (Infrastructure Layer)
	// Auth
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	roleRepo := sqlite.NewRoleRepository(db)
	permissionRepo := sqlite.NewPermissionRepository(db)

	// Domain
	ownerRepo := sqlite.NewOwnerRepository(db)
	apartmentRepo := sqlite.NewApartmentRepository(db)
	ownershipRepo := sqlite.NewOwnershipShareRepository(db)
	chargeRepo := sqlite.NewChargeRepository(db)
	paymentRepo := sqlite.NewPaymentRepository(db)
	expenseRepo := sqlite.NewExpenseRepository(db)
	expenseCategoryRepo := sqlite.NewExpenseCategoryRepository(db)
	contractorRepo := sqlite.NewContractorRepository(db)

	// 4. Створення services (Application Layer)
	passwordHasher := security.DefaultBCryptHasher()

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		roleRepo,
		permissionRepo,
		passwordHasher,
	)

	// 4.1. Перевірка та створення дефолтного адміна
	ctx := context.Background()
	_, err = userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		// Користувача не існує (або помилка БД) - спробуємо створити
		log.Println("Default admin not found, creating...")

		_, err = authService.Register(ctx, authUseCase.RegisterUserInput{
			Username:  "admin",
			Email:     "admin@osbb.local",
			Password:  "Admin123!",
			FirstName: "System",
			LastName:  "Administrator",
			RoleName:  "admin",
		})
		if err != nil {
			log.Printf("Failed to create default admin: %v", err)
		} else {
			log.Println("Default admin created successfully: admin / Admin123!")
		}
	}

	ownerService := service.NewOwnerService(
		ownerRepo,
		permissionRepo,
	)

	apartmentService := service.NewApartmentService(
		apartmentRepo,
		permissionRepo,
	)

	ownershipService := service.NewOwnershipService(
		ownershipRepo,
		ownerRepo,
		apartmentRepo,
		permissionRepo,
	)

	chargeService := service.NewChargeService(
		chargeRepo,
		permissionRepo,
	)

	paymentService := service.NewPaymentService(
		paymentRepo,
		permissionRepo,
	)

	expenseService := service.NewExpenseService(
		expenseRepo,
		permissionRepo,
		expenseCategoryRepo,
	)

	// Note: expenseCategoryService is initialized for managing expense categories in the UI
	expenseCategoryService := service.NewExpenseCategoryService(
		expenseCategoryRepo,
		permissionRepo,
	)

	contractorService := service.NewContractorService(
		contractorRepo,
		permissionRepo,
	)

	// 5. Створення AuthManager (Presentation Layer)
	authManager := auth.NewAuthManager(myApp, authService)

	// 6. Налаштування callbacks навігації
	authManager.OnLoginSuccess(func() {
		// Передаємо всі необхідні сервіси у головний екран
		screens.ShowMainScreen(mainWindow, authManager, ownerService, apartmentService, ownershipService, chargeService, paymentService, expenseService, expenseCategoryService, contractorService)
	})

	authManager.OnLogout(func() {
		authManager.ShowLoginScreen(mainWindow)
	})

	// 7. Старт
	authManager.ShowLoginScreen(mainWindow)

	mainWindow.SetOnClosed(func() {
		authManager.Cleanup()
	})

	mainWindow.ShowAndRun()
}
