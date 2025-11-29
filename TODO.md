# Cash Flow Screen Enhancement with Contractor Payments

Implement `ContractorPayment` entity for rental income from contractors (e.g., Vodafone, Life) and update the Cash Flow screen with a proper 12-column accounting format showing detailed transaction breakdowns by expense categories.

## User Review Required

> [!IMPORTANT]
> **New Entity**: Creating `ContractorPayment` as a separate entity (not linked to apartments/owners) for tracking rental payments from contractors like Vodafone and Life. This is a cleaner domain model than trying to reuse existing Payment/OwnershipShare entities.

> [!IMPORTANT]
> **UI Change**: Cash Flow table format will change from 7 columns to 12 columns with expense category breakdown (313-Картка, 63-Електроенергія, 641-ПДФО, 641.1-Військовий збір, 651-ЄСВ, 94-Комісія). Expenses not matching these categories will only appear in "Оборот по кт" (total credit).

> [!WARNING]
> **Existing Data**: The 6 standard expense categories will be created with specific codes (313, 63, 641, 641.1, 651, 94). If these codes already exist in your database, migration will fail. Manual intervention may be required.

## Proposed Changes

### Domain Layer

#### [NEW] [contractor_payment.go](file:///home/min/git-workspace/osbb-accounting/domain/entity/contractor_payment.go)
New entity representing rental/usage payments from contractors. Fields:
- `ID`, `ContractorID` (FK to contractors table)
- `PaymentDate`, `Amount`, `PaymentMethod`, `Purpose`
- `PeriodMonth`, `PeriodYear` (optional period link)
- `ReceiptNumber`, `Notes`
- Soft delete pattern, timestamps
- **No approval fields** (per user requirement)

#### [NEW] [contractor_payment_repository.go](file:///home/min/git-workspace/osbb-accounting/domain/repository/contractor_payment_repository.go)
Repository interface with standard CRUD operations:
- `Create`, `GetByID`, `Update`, `SoftDelete`, `Restore`
- `List` (with filter by contractor, date range, period, search)
- `Count`, `GetByContractorID`, `GetTotalByPeriod`

---

### Infrastructure Layer

#### [NEW] [contractor_payment_repository.go](file:///home/min/git-workspace/osbb-accounting/infrastructure/persistence/sqlite/contractor_payment_repository.go)
SQLite implementation following existing Payment repository pattern with full filtering, pagination, and search support.

#### [NEW] [004_contractor_payments.sql](file:///home/min/git-workspace/osbb-accounting/migrations/004_contractor_payments.sql)
Migration creating `contractor_payments` table with:
- Standard columns following existing pattern
- Indexes on contractor_id, payment_date, period
- Updated_at trigger
- **6 standard expense categories** with codes: 313, 63, 641, 641.1, 651, 94

---

### Application Layer

#### [NEW] [contractor_payment_usecase.go](file:///home/min/git-workspace/osbb-accounting/application/usecase/contractor_payment/contractor_payment_usecase.go)
Use cases: `Create`, `Update`, `GetByID`, `List`, `Delete` with business validation.

#### [NEW] [contractor_payment_service.go](file:///home/min/git-workspace/osbb-accounting/application/service/contractor_payment_service.go)
Service layer enriching contractor payment data with contractor names and providing high-level operations.

#### [MODIFY] [cashflow_usecase.go](file:///home/min/git-workspace/osbb-accounting/application/usecase/cashflow/cashflow_usecase.go)
- Add `ContractorPaymentRepository` dependency
- Update `List()` to fetch and merge 3 transaction types: Payment, ContractorPayment, Expense
- Keep chronological sorting by date

#### [MODIFY] [CashFlowEntry](file:///home/min/git-workspace/osbb-accounting/application/usecase/cashflow/cashflow_usecase.go#L12-L23)
Add fields for 12-column breakdown:
- New `Type` value: `"contractor_payment"`
- New fields: `Credit313`, `Credit63`, `Credit641`, `Credit6411`, `Credit651`, `Credit94` (category-specific amounts)
- Existing total credit/debit preserved

#### [MODIFY] [cashflow_service.go](file:///home/min/git-workspace/osbb-accounting/application/service/cashflow_service.go)
- Update `enrichEntries()` to handle ContractorPayment type
- Map expenses to category columns by category code
- Update `getCounterparty()` logic:
  - Payment → Owner.LastName (capitalized)
  - ContractorPayment → Contractor.Name
  - Expense → Category.Name (for taxes) or Contractor.Name

---

### Presentation Layer

#### [MODIFY] [cashflow_screen.go](file:///home/min/git-workspace/osbb-accounting/presentation/fyne/screens/cashflow_screen.go)
Update table to 12 columns:
1. № п/п (row number)
2. Контрагент (payer/contractor/category)
3. Дата (transaction date)
4. Дт рах. 311 (debit amount)
5. Оборот по дт (sum of col 4, with subtotal)
6. 313 - Кошти на картку
7. 63 - Електроенергія
8. 641 - ПДФО 18%
9. 641.1 - Військовий збір
10. 651 - ЄСВ 22%
11. 94 - Комісія банку
12. Оборот по кт (sum of cols 6-11, with subtotal)

Update column widths, rendering logic, and stats label.

#### [MODIFY] [export_xlsx.go](file:///home/min/git-workspace/osbb-accounting/application/usecase/cashflow/export_xlsx.go)
Update XLSX export with new 12-column format, category columns, and proper subtotals.

#### [NEW] [contractor_payment_screen.go](file:///home/min/git-workspace/osbb-accounting/presentation/fyne/screens/contractor_payment_screen.go)
New UI screen for managing contractor payments with:
- Table showing contractor name, date, amount, period, purpose
- Add/Edit/Delete buttons
- Period filters (month/year)
- Search by contractor name or purpose

#### [NEW] [contractor_payment_form.go](file:///home/min/git-workspace/osbb-accounting/presentation/fyne/dialogs/contractor_payment_form.go)
Dialog for creating/editing contractor payments with fields:
- Contractor selection (dropdown)
- Date picker, amount input
- Payment method select
- Purpose text field
- Period (optional month/year)
- Receipt number, notes

#### [MODIFY] [main_screen.go](file:///home/min/git-workspace/osbb-accounting/presentation/fyne/screens/main_screen.go)
Add "Платежі контрагентів" menu item in Фінанси section.

---

### Bootstrap & Integration

#### [MODIFY] [bootstrap.go](file:///home/min/git-workspace/osbb-accounting/cmd/bootstrap/bootstrap.go)
- Initialize ContractorPaymentRepository
- Initialize ContractorPaymentService
- Add to ServiceContainer
- Pass ContractorPaymentRepository to CashFlowUseCase

## Verification Plan

### Automated Tests

```bash
# Run all tests to ensure no regressions
go test ./...
```

Tests to add:
1. `domain/entity/contractor_payment_test.go` - entity validation tests
2. `infrastructure/persistence/sqlite/contractor_payment_repository_test.go` - CRUD tests
3. `application/usecase/contractor_payment/contractor_payment_usecase_test.go` - use case tests

### Manual Verification

1. **Database Migration**
   ```bash
   # Run application to trigger migration
   go run cmd/main.go
   # Verify 6 expense categories created with correct codes
   ```

2. **Contractor Payment CRUD**
   - Navigate to "Платежі контрагентів" screen
   - Create new contractor payment (ensure contractor exists first)
   - Edit existing payment
   - Delete payment (soft delete)
   - Verify all fields save correctly

3. **Cash Flow Screen**
   - Navigate to "Рух коштів" screen
   - Verify 12 columns display correctly
   - Create test data:
     - Regular owner payment → shows in Дебет
     - Contractor payment → shows in Дебет with contractor name
     - Expense with category 63 (Електroенергія) → shows in column 7
     - Expense with category 641 (ПДФО) → shows in column 8
   - Verify subtotals calculate correctly
   - Export XLSX and verify format

4. **User Approval Required**
   - User should test with real data after implementation
   - Verify category mapping works for actual expenses
   - Confirm 12-column format meets accounting requirements
