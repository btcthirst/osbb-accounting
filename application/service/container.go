package service

// ServiceContainer holds all application services
type ServiceContainer struct {
	AuthService              *AuthService
	OwnerService             *OwnerService
	ApartmentService         *ApartmentService
	OwnershipService         *OwnershipService
	ChargeService            *ChargeService
	PaymentService           *PaymentService
	ContractorPaymentService *ContractorPaymentService
	ExpenseService           *ExpenseService
	ExpenseCategoryService   *ExpenseCategoryService
	ContractorService        *ContractorService
	CashFlowService          *CashFlowService
	OSBBService              *OSBBService
	ImportService            ImportServiceInterface
}
