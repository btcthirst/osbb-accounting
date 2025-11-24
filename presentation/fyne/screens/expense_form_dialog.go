// presentation/fyne/screens/expense_form_dialog.go
package screens

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense"
	"osbb-accounting/domain/entity"
)

// ExpenseFormDialog - діалог створення/редагування витрати
type ExpenseFormDialog struct {
	window         fyne.Window
	expenseService *service.ExpenseService
	authManager    interface {
		GetCurrentUserID() int64
	}

	// Data
	existingExpense *expense.ExpenseOutput

	// UI Elements
	categoryEntry    *widget.Entry // Placeholder for Select
	amountEntry      *widget.Entry
	expenseDateEntry *widget.Entry // Format: DD.MM.YYYY
	descriptionEntry *widget.Entry

	// Document
	docTypeSelect  *widget.Select
	docNumberEntry *widget.Entry
	docDateEntry   *widget.Entry

	notesEntry *widget.Entry

	// Callbacks
	onSaved func()
}

// NewExpenseFormDialog створює новий діалог
func NewExpenseFormDialog(
	window fyne.Window,
	expenseService *service.ExpenseService,
	authManager interface {
		GetCurrentUserID() int64
	},
	existingExpense *expense.ExpenseOutput,
) *ExpenseFormDialog {
	return &ExpenseFormDialog{
		window:          window,
		expenseService:  expenseService,
		authManager:     authManager,
		existingExpense: existingExpense,
	}
}

// Show показує діалог
func (d *ExpenseFormDialog) Show() {
	title := "Нова витрата"
	if d.existingExpense != nil {
		title = "Редагування витрати"
	}

	// Category (TODO: Use Select with CategoryService)
	d.categoryEntry = widget.NewEntry()
	d.categoryEntry.PlaceHolder = "ID категорії (тимчасово)"

	// Amount
	d.amountEntry = widget.NewEntry()
	d.amountEntry.PlaceHolder = "0.00"

	// Date
	d.expenseDateEntry = widget.NewEntry()
	d.expenseDateEntry.PlaceHolder = "DD.MM.YYYY"
	d.expenseDateEntry.SetText(time.Now().Format("02.01.2006"))

	// Description
	d.descriptionEntry = widget.NewEntry()
	d.descriptionEntry.PlaceHolder = "Опис витрати"

	// Document
	docTypes := []string{
		entity.DocumentTypeInvoice.GetDisplayName(),
		entity.DocumentTypeAct.GetDisplayName(),
		entity.DocumentTypeReceipt.GetDisplayName(),
		entity.DocumentTypeOrder.GetDisplayName(),
		entity.DocumentTypeOther.GetDisplayName(),
	}
	d.docTypeSelect = widget.NewSelect(docTypes, nil)
	d.docTypeSelect.PlaceHolder = "Тип документа"

	d.docNumberEntry = widget.NewEntry()
	d.docNumberEntry.PlaceHolder = "Номер документа"

	d.docDateEntry = widget.NewEntry()
	d.docDateEntry.PlaceHolder = "Дата документа (DD.MM.YYYY)"

	// Notes
	d.notesEntry = widget.NewMultiLineEntry()
	d.notesEntry.PlaceHolder = "Примітки..."
	d.notesEntry.SetMinRowsVisible(3)

	// Pre-fill if editing
	if d.existingExpense != nil {
		d.categoryEntry.SetText(fmt.Sprintf("%d", d.existingExpense.CategoryID))
		d.amountEntry.SetText(fmt.Sprintf("%.2f", d.existingExpense.Amount))
		d.expenseDateEntry.SetText(d.existingExpense.ExpenseDate.Format("02.01.2006"))
		d.descriptionEntry.SetText(d.existingExpense.Description)

		if d.existingExpense.DocumentType != nil {
			// Find display name
			dt := entity.DocumentType(*d.existingExpense.DocumentType)
			d.docTypeSelect.SetSelected(dt.GetDisplayName())
		}
		if d.existingExpense.DocumentNumber != nil {
			d.docNumberEntry.SetText(*d.existingExpense.DocumentNumber)
		}
		if d.existingExpense.DocumentDate != nil {
			d.docDateEntry.SetText(d.existingExpense.DocumentDate.Format("02.01.2006"))
		}
		if d.existingExpense.Notes != nil {
			d.notesEntry.SetText(*d.existingExpense.Notes)
		}
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Категорія ID", d.categoryEntry),
		widget.NewFormItem("Сума (грн)", d.amountEntry),
		widget.NewFormItem("Дата витрати", d.expenseDateEntry),
		widget.NewFormItem("Опис", d.descriptionEntry),
		widget.NewFormItem("Тип документа", d.docTypeSelect),
		widget.NewFormItem("№ Документа", d.docNumberEntry),
		widget.NewFormItem("Дата документа", d.docDateEntry),
		widget.NewFormItem("Примітки", d.notesEntry),
	}

	dialog.ShowForm(title, "Зберегти", "Скасувати", formItems, func(confirm bool) {
		if !confirm {
			return
		}

		// Validation & Save

		// 1. Category
		catID, err := strconv.ParseInt(d.categoryEntry.Text, 10, 64)
		if err != nil || catID <= 0 {
			dialog.ShowError(fmt.Errorf("некоректний ID категорії"), d.window)
			return
		}

		// 2. Amount
		amount, err := strconv.ParseFloat(d.amountEntry.Text, 64)
		if err != nil || amount <= 0 {
			dialog.ShowError(fmt.Errorf("некоректна сума"), d.window)
			return
		}

		// 3. Date
		eDate, err := time.Parse("02.01.2006", d.expenseDateEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("некоректна дата (DD.MM.YYYY)"), d.window)
			return
		}

		// 4. Description
		if d.descriptionEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("опис обов'язковий"), d.window)
			return
		}

		// 5. Document (Optional)
		var docType *entity.DocumentType
		var docNum *string
		var docDate *time.Time

		if d.docTypeSelect.Selected != "" {
			var dt entity.DocumentType
			switch d.docTypeSelect.Selected {
			case entity.DocumentTypeInvoice.GetDisplayName():
				dt = entity.DocumentTypeInvoice
			case entity.DocumentTypeAct.GetDisplayName():
				dt = entity.DocumentTypeAct
			case entity.DocumentTypeReceipt.GetDisplayName():
				dt = entity.DocumentTypeReceipt
			case entity.DocumentTypeOrder.GetDisplayName():
				dt = entity.DocumentTypeOrder
			case entity.DocumentTypeOther.GetDisplayName():
				dt = entity.DocumentTypeOther
			}
			docType = &dt
		}

		if d.docNumberEntry.Text != "" {
			val := d.docNumberEntry.Text
			docNum = &val
		}

		if d.docDateEntry.Text != "" {
			t, err := time.Parse("02.01.2006", d.docDateEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("некоректна дата документа"), d.window)
				return
			}
			docDate = &t
		}

		var notes *string
		if d.notesEntry.Text != "" {
			val := d.notesEntry.Text
			notes = &val
		}

		// Execute
		ctx := context.Background()
		userID := d.authManager.GetCurrentUserID()

		if d.existingExpense == nil {
			// Create
			input := expense.CreateExpenseInput{
				CurrentUserID:  userID,
				CategoryID:     catID,
				ExpenseDate:    eDate,
				Amount:         amount,
				Description:    d.descriptionEntry.Text,
				DocumentType:   docType,
				DocumentNumber: docNum,
				DocumentDate:   docDate,
				Notes:          notes,
			}
			_, err := d.expenseService.Create(ctx, input)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
		} else {
			// Update
			input := expense.UpdateExpenseInput{
				CurrentUserID: userID,
				ExpenseID:     d.existingExpense.ID,
				CategoryID:    catID,
				ExpenseDate:   eDate,
				Amount:        amount,
				Description:   d.descriptionEntry.Text,
				Notes:         notes,
			}
			_, err := d.expenseService.Update(ctx, input)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
		}

		if d.onSaved != nil {
			d.onSaved()
		}
	}, d.window)
}

// SetOnSaved sets the callback
func (d *ExpenseFormDialog) SetOnSaved(f func()) {
	d.onSaved = f
}
