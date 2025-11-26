package service

// ServiceContainer holds all application services
type ServiceContainer struct {
	AuthService            *AuthService
	OwnerService           *OwnerService
	ApartmentService       *ApartmentService
	OwnershipService       *OwnershipService
	ChargeService          *ChargeService
	PaymentService         *PaymentService
	ExpenseService         *ExpenseService
	ExpenseCategoryService *ExpenseCategoryService
	ContractorService      *ContractorService
	OSBBService            *OSBBService
}
