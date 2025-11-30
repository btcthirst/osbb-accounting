package text

// Table Headers
const (
	// Apartments
	TableHeaderNumber          = "№"
	TableHeaderDisplayName     = "ПІБ"
	TableHeaderApartmentNumber = "Номер"
	TableHeaderAreaLiving      = "Жила площа"
	TableHeaderEntrance        = "Підїзд"
	TableHeaderActions         = "Дії"
	// CashFlow
	TableHeaderCashFlowNumber       = "№ п/п"
	TableHeaderCashFlowCounterparty = "Контрагент"
	TableHeaderCashFlowDate         = "Дата"
	TableHeaderCashFlowDebit311     = "Дт рах. 311"
	TableHeaderCashFlowDebitTotal   = "Оборот по дт"
	TableHeaderCashFlowCredit313    = "313"
	TableHeaderCashFlowCredit63     = "63"
	TableHeaderCashFlowCredit641    = "641"
	TableHeaderCashFlowCredit6411   = "641.1"
	TableHeaderCashFlowCredit651    = "651"
	TableHeaderCashFlowCredit94     = "94"
	TableHeaderCashFlowCreditTotal  = "Оборот по кт"
	// Charges

	// Payments "ID", "Контрагент", "Дата", "Період", "Метод", "Сума", "Дії"

	// ExpenseCategories
	TableHeaderExpenseCategoryNumber  = TableHeaderNumber
	TableHeaderExpenseCategoryName    = "Категорія витрат"
	TableHeaderExpenseCategoryActions = TableHeaderActions
	// Contractors []string{"Назва", "Тип", "ЄДРПОУ", "Контакт", "Статус", "Дії"}
	TableHeaderContractorName    = "Назва"
	TableHeaderContractorType    = "Тип"
	TableHeaderContractorEDRPOU  = "ЄДРПОУ"
	TableHeaderContractorContact = "Контакт"
	TableHeaderContractorStatus  = "Статус"
	TableHeaderContractorActions = TableHeaderActions
	// Expenses
	TableHeaderExpenseID       = "ID"
	TableHeaderExpenseDate     = TableHeaderCashFlowDate
	TableHeaderExpenseCategory = "Категорія"
	TableHeaderExpenseAmount   = TableHeaderChargeAmount
	TableHeaderExpenseStatus   = "Статус"
	TableHeaderExpenseApproved = "Підтверджено"
	TableHeaderExpenseActions  = TableHeaderActions
	// Owners
	TableHeaderOwnerName    = TableHeaderDisplayName
	TableHeaderOwnerPhone   = "Телефон"
	TableHeaderOwnerEmail   = "Email"
	TableHeaderOwnerTax     = "ІПН"
	TableHeaderOwnerActions = TableHeaderActions
	// Ownership Shares
	TableHeaderShareOwner     = "Власник"
	TableHeaderShareApartment = "Квартира"
	TableHeaderShareFraction  = "Частка"
	TableHeaderShareType      = "Тип"
	TableHeaderShareDates     = "Дати"
	TableHeaderShareActions   = TableHeaderActions
	// Charges
	TableHeaderChargeDate    = "Дата"
	TableHeaderChargePeriod  = "Період"
	TableHeaderChargeType    = "Тип"
	TableHeaderChargeAmount  = "Сума"
	TableHeaderChargePayer   = "Платник"
	TableHeaderChargeActions = TableHeaderActions
	// Payments
	TableHeaderPaymentID       = "ID"
	TableHeaderPaymentDate     = "Дата"
	TableHeaderPaymentPeriod   = "Період"
	TableHeaderPaymentMethod   = "Метод"
	TableHeaderPaymentAmount   = "Сума"
	TableHeaderPaymentApproved = "Підтверджено"
	TableHeaderPaymentReceipt  = "Квитанція"
	TableHeaderPaymentActions  = TableHeaderActions
	// Expense Categories
	TableHeaderCategoryName        = "Назва"
	TableHeaderCategoryType        = "Тип"
	TableHeaderCategoryDescription = "Опис"
	TableHeaderCategoryStatus      = "Статус"
	TableHeaderCategoryActions     = TableHeaderActions
	// Import Batches
	TableHeaderBatchID      = "ID"
	TableHeaderBatchFile    = "Файл"
	TableHeaderBatchStatus  = "Статус"
	TableHeaderBatchSheets  = "Аркушів"
	TableHeaderBatchRecords = "Записів"
	TableHeaderBatchDate    = "Дата імпорту"
	TableHeaderBatchActions = TableHeaderActions
	// Import Records (Month)
	TableHeaderRecordApartment = "Квартира"
	TableHeaderRecordName      = "ПІБ"
	TableHeaderRecordPeriod    = "Період"
	TableHeaderRecordCharged   = "Нараховано"
	TableHeaderRecordPaid      = "Сплачено"
	TableHeaderRecordDebt      = "Борг"
	// ContractorPayments
	TableHeaderContractorPaymentID         = "ID"
	TableHeaderContractorPaymentContractor = "Контрагент"
	TableHeaderContractorPaymentDate       = "Дата"
	TableHeaderContractorPaymentPeriod     = "Період"
	TableHeaderContractorPaymentMethod     = "Метод"
	TableHeaderContractorPaymentAmount     = "Сума"
	TableHeaderContractorPaymentActions    = TableHeaderActions
)

var (
	ApartmentsTableHeaders = []string{TableHeaderNumber, TableHeaderDisplayName, TableHeaderApartmentNumber,
		TableHeaderAreaLiving, TableHeaderEntrance, TableHeaderActions}
	CashFlowTableHeaders = []string{TableHeaderNumber, TableHeaderDisplayName, TableHeaderApartmentNumber,
		TableHeaderAreaLiving, TableHeaderEntrance, TableHeaderActions}
	ContractorPaymentsTableHeaders = []string{TableHeaderContractorPaymentID, TableHeaderContractorPaymentContractor,
		TableHeaderContractorPaymentDate, TableHeaderContractorPaymentPeriod, TableHeaderContractorPaymentMethod,
		TableHeaderContractorPaymentAmount, TableHeaderContractorPaymentActions}
	ContractorsTableHeaders = []string{TableHeaderNumber, TableHeaderContractorName, TableHeaderContractorType,
		TableHeaderContractorEDRPOU, TableHeaderContractorContact, TableHeaderContractorStatus, TableHeaderContractorActions}
	ExpensesTableHeaders = []string{TableHeaderExpenseID, TableHeaderExpenseDate, TableHeaderExpenseCategory,
		TableHeaderExpenseAmount, TableHeaderExpenseStatus, TableHeaderExpenseApproved, TableHeaderExpenseActions}
	OwnersTableHeaders = []string{TableHeaderNumber, TableHeaderOwnerName, TableHeaderOwnerPhone,
		TableHeaderOwnerEmail, TableHeaderOwnerTax, TableHeaderOwnerActions}
	OwnershipSharesTableHeaders = []string{TableHeaderNumber, TableHeaderShareOwner, TableHeaderShareApartment,
		TableHeaderShareFraction, TableHeaderShareType, TableHeaderShareDates, TableHeaderShareActions}
	ChargesTableHeaders = []string{TableHeaderNumber, TableHeaderChargeDate, TableHeaderChargePeriod,
		TableHeaderChargeType, TableHeaderChargeAmount, TableHeaderChargePayer, TableHeaderChargeActions}
	PaymentsTableHeaders = []string{TableHeaderPaymentID, TableHeaderPaymentDate, TableHeaderPaymentPeriod,
		TableHeaderPaymentMethod, TableHeaderPaymentAmount, TableHeaderPaymentActions}
	ExpenseCategoriesTableHeaders = []string{TableHeaderCategoryName, TableHeaderCategoryType, TableHeaderCategoryDescription,
		TableHeaderCategoryStatus, TableHeaderCategoryActions}
	ImportBatchesTableHeaders = []string{TableHeaderBatchID, TableHeaderBatchFile, TableHeaderBatchStatus,
		TableHeaderBatchSheets, TableHeaderBatchRecords, TableHeaderBatchDate, TableHeaderBatchActions}
	ImportRecordsTableHeaders = []string{TableHeaderRecordApartment, TableHeaderRecordName, TableHeaderRecordPeriod,
		TableHeaderRecordCharged, TableHeaderRecordPaid, TableHeaderRecordDebt}
)
