package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"osbb-accounting/application/service"
	authUseCase "osbb-accounting/application/usecase/auth"
	"osbb-accounting/config"
	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/security"
)

// App holds all application dependencies
type App struct {
	Config   *config.Config
	DB       *sql.DB
	Services *service.ServiceContainer
}

// Initialize sets up the database, repositories, and services
func Initialize(cfg *config.Config) (*App, error) {
	// 1. Initialize Database
	dbConfig := &sqlite.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	}

	db, err := sqlite.Connect(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 2. Initialize Repositories (Infrastructure Layer)
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
	osbbRepo := sqlite.NewOSBBRepository(db)

	// Import
	importBatchRepo := sqlite.NewImportBatchDAO(db)
	importRecordRepo := sqlite.NewImportedRecordDAO(db)

	// 3. Initialize Services (Application Layer)
	passwordHasher := security.DefaultBCryptHasher()

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		roleRepo,
		permissionRepo,
		passwordHasher,
	)

	// 3.1. Ensure Default Admin Exists
	ctx := context.Background()
	if err := ensureDefaultAdmin(ctx, userRepo, authService); err != nil {
		log.Printf("Warning: Failed to ensure default admin: %v", err)
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

	expenseCategoryService := service.NewExpenseCategoryService(
		expenseCategoryRepo,
		permissionRepo,
	)

	contractorService := service.NewContractorService(
		contractorRepo,
		permissionRepo,
	)

	osbbService := service.NewOSBBService(
		osbbRepo,
		permissionRepo,
	)

	importService := service.NewImportService(
		importBatchRepo,
		importRecordRepo,
	)

	// 4. Create Service Container
	services := &service.ServiceContainer{
		AuthService:            authService,
		OwnerService:           ownerService,
		ApartmentService:       apartmentService,
		OwnershipService:       ownershipService,
		ChargeService:          chargeService,
		PaymentService:         paymentService,
		ExpenseService:         expenseService,
		ExpenseCategoryService: expenseCategoryService,
		ContractorService:      contractorService,
		OSBBService:            osbbService,
		ImportService:          importService,
	}

	return &App{
		Config:   cfg,
		DB:       db,
		Services: services,
	}, nil
}

// Close closes resources
func (app *App) Close() {
	if app.DB != nil {
		sqlite.Close(app.DB)
	}
}

func ensureDefaultAdmin(ctx context.Context, userRepo *sqlite.UserRepository, authService *service.AuthService) error {
	_, err := userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		// User not found (or DB error) - try to create
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
			return err
		}
		log.Println("Default admin created successfully: admin / Admin123!")
	}
	return nil
}
