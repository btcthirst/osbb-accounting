// presentation/fyne/screens/expense_category_form_dialog.go
package dialogs

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/domain/entity"
)

// ExpenseCategoryFormDialog - діалог створення/редагування категорії витрат
type ExpenseCategoryFormDialog struct {
	window                 fyne.Window
	expenseCategoryService *service.ExpenseCategoryService
	authManager            interface {
		GetCurrentUserID() int64
	}

	// Data
	existingCategory *expense_category.ExpenseCategoryOutput

	// UI Elements
	nameEntry        *widget.Entry
	typeSelect       *widget.Select
	descriptionEntry *widget.Entry

	// Callbacks
	onSaved func()
}

// NewExpenseCategoryFormDialog створює новий діалог
func NewExpenseCategoryFormDialog(
	window fyne.Window,
	expenseCategoryService *service.ExpenseCategoryService,
	authManager interface {
		GetCurrentUserID() int64
	},
	existingCategory *expense_category.ExpenseCategoryOutput,
) *ExpenseCategoryFormDialog {
	return &ExpenseCategoryFormDialog{
		window:                 window,
		expenseCategoryService: expenseCategoryService,
		authManager:            authManager,
		existingCategory:       existingCategory,
	}
}

// Show показує діалог
func (d *ExpenseCategoryFormDialog) Show() {
	title := "Нова категорія витрат"
	if d.existingCategory != nil {
		title = "Редагування категорії витрат"
	}

	// Name
	d.nameEntry = widget.NewEntry()
	d.nameEntry.PlaceHolder = "Назва категорії"

	// Type
	types := []string{
		entity.CategoryTypeUtility.GetDisplayName(),
		entity.CategoryTypeRepair.GetDisplayName(),
		entity.CategoryTypeSalary.GetDisplayName(),
		entity.CategoryTypeService.GetDisplayName(),
		entity.CategoryTypeOther.GetDisplayName(),
	}
	d.typeSelect = widget.NewSelect(types, nil)
	d.typeSelect.PlaceHolder = "Тип категорії"

	// Description
	d.descriptionEntry = widget.NewMultiLineEntry()
	d.descriptionEntry.PlaceHolder = "Опис (необов'язково)..."
	d.descriptionEntry.SetMinRowsVisible(3)

	// Pre-fill if editing
	if d.existingCategory != nil {
		d.nameEntry.SetText(d.existingCategory.Name)
		d.typeSelect.SetSelected(d.existingCategory.TypeDisplayName)
		if d.existingCategory.Description != nil {
			d.descriptionEntry.SetText(*d.existingCategory.Description)
		}
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Назва", d.nameEntry),
		widget.NewFormItem("Тип", d.typeSelect),
		widget.NewFormItem("Опис", d.descriptionEntry),
	}

	dialog.ShowForm(title, "Зберегти", "Скасувати", formItems, func(confirm bool) {
		if !confirm {
			return
		}

		// Validation & Save

		// 1. Name
		if d.nameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("назва обов'язкова"), d.window)
			return
		}

		// 2. Type
		var cType entity.CategoryType
		switch d.typeSelect.Selected {
		case entity.CategoryTypeUtility.GetDisplayName():
			cType = entity.CategoryTypeUtility
		case entity.CategoryTypeRepair.GetDisplayName():
			cType = entity.CategoryTypeRepair
		case entity.CategoryTypeSalary.GetDisplayName():
			cType = entity.CategoryTypeSalary
		case entity.CategoryTypeService.GetDisplayName():
			cType = entity.CategoryTypeService
		case entity.CategoryTypeOther.GetDisplayName():
			cType = entity.CategoryTypeOther
		default:
			dialog.ShowError(fmt.Errorf("оберіть тип категорії"), d.window)
			return
		}

		// Optional fields helper
		toPtr := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}

		// Execute
		ctx := context.Background()
		userID := d.authManager.GetCurrentUserID()

		if d.existingCategory == nil {
			// Create
			input := expense_category.CreateExpenseCategoryInput{
				CurrentUserID: userID,
				Name:          d.nameEntry.Text,
				CategoryType:  cType,
				Description:   toPtr(d.descriptionEntry.Text),
			}
			_, err := d.expenseCategoryService.Create(ctx, input)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
		} else {
			// Update
			input := expense_category.UpdateExpenseCategoryInput{
				CurrentUserID: userID,
				CategoryID:    d.existingCategory.ID,
				Name:          d.nameEntry.Text,
				CategoryType:  cType,
				Description:   toPtr(d.descriptionEntry.Text),
			}
			_, err := d.expenseCategoryService.Update(ctx, input)
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
func (d *ExpenseCategoryFormDialog) SetOnSaved(f func()) {
	d.onSaved = f
}
