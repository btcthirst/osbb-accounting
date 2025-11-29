### where we stopped

1. імпорт xlsx файла (поки тільки нарахування*.xlsx) [ok]
2. внесення данних файла в базу
3. вивід таблиці (меню статистика) [ok]
4. екран рух коштів [ok]


###### cash_flow_01
Завершив імплементацію архітектури для екрану "Рух коштів":

1. UI екран з таблицею та фільтрами (add search/change table format)
2. Функція експорту в XLSX (change table format)

# start here #
# Overview 
Implement ContractorPayment entity and update Cash Flow screen with proper accounting format (12 columns) and support for 3 types of transactions.

Tasks
### Domain Layer
- [v] Create ContractorPayment entity
- [v] Create ContractorPaymentRepository interface
- [v] Implement SQLite repository for ContractorPayment
- [v] Add migrations for contractor_payments table
### Application Layer
- [v] Create ContractorPayment use cases (Create, List, Update, Delete)
- [v] Create ContractorPaymentService
- [v] Update CashFlowUseCase to support 3 transaction types
- [v] Update CashFlowEntry structure for 12-column format
### Infrastructure Layer
- [v] Create 6 standard expense categories with codes (313, 63, 641, 641.1, 651, 94)
- [v] Add expense category initialization in bootstrap
### Presentation Layer
- [v] Update CashFlowScreen with 12 columns
- [v] Update column mapping logic for expense categories
- [ ] Update XLSX export with new format
- [ ] Create ContractorPaymentScreen UI
- [ ] Create ContractorPaymentFormDialog
- [ ] Add ContractorPayment menu item
### Integration
- [v] Initialize ContractorPaymentService in bootstrap
- [v] Add ContractorPayment to ServiceContainer
- [ ] Update main screen menu
- [ ] Test all screens compile and run
# end here #