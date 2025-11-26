// presentation/fyne/screens/expense_category_screen.go
package screens

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense_category"
)

// ExpenseCategoryScreen - екран списку категорій витрат
type ExpenseCategoryScreen struct {
	window                 fyne.Window
	expenseCategoryService *service.ExpenseCategoryService
	authManager            interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	}

	// UI
	content *fyne.Container
	table   *widget.Table

	// Data
	categories []*expense_category.ExpenseCategoryOutput

	// Filters
	searchEntry *widget.Entry
}

// NewExpenseCategoryScreen створює новий екран
func NewExpenseCategoryScreen(
	window fyne.Window,
	expenseCategoryService *service.ExpenseCategoryService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *ExpenseCategoryScreen {
	s := &ExpenseCategoryScreen{
		window:                 window,
		expenseCategoryService: expenseCategoryService,
		authManager:            authManager,
	}
	s.buildUI()
	return s
}

func (s *ExpenseCategoryScreen) buildUI() {
	// Buttons
	createButton := widget.NewButtonWithIcon("Додати категорію", theme.ContentAddIcon(), func() {
		s.showCategoryDialog(nil)
	})
	createButton.Importance = widget.HighImportance

	refreshButton := widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadCategories()
	})

	// Filters
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("Пошук (Назва)...")
	s.searchEntry.OnSubmitted = func(_ string) { s.loadCategories() }

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.categories) + 1, 5 // +1 for header
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell content")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				headers := []string{"Назва", "Тип", "Опис", "Статус", "Дії"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			// Data
			if id.Row-1 >= len(s.categories) {
				label.SetText("")
				return
			}
			c := s.categories[id.Row-1]

			// Reset style
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0: // Name
				label.SetText(c.Name)
			case 1: // Type
				label.SetText(c.TypeDisplayName)
			case 2: // Description
				if c.Description != nil {
					label.SetText(*c.Description)
				} else {
					label.SetText("-")
				}
			case 3: // Status
				if c.IsActive {
					label.SetText("Активна")
				} else {
					label.SetText("Неактивна")
				}
			case 4: // Actions
				label.SetText("⚙️")
			}
		},
	)

	// Column widths
	s.table.SetColumnWidth(0, 200) // Name
	s.table.SetColumnWidth(1, 150) // Type
	s.table.SetColumnWidth(2, 300) // Description
	s.table.SetColumnWidth(3, 100) // Status
	s.table.SetColumnWidth(4, 50)  // Actions

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id)
		if id.Row == 0 {
			return
		}
		if id.Col == 4 {
			s.showActionsMenu(id.Row - 1)
		}
	}

	// Toolbar container
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(createButton, refreshButton),
		nil,
		s.searchEntry,
	)

	// Layout
	s.content = container.NewBorder(
		toolbar,
		nil, nil, nil,
		s.table,
	)

	s.loadCategories()
}

func (s *ExpenseCategoryScreen) Render() fyne.CanvasObject {
	return s.content
}

func (s *ExpenseCategoryScreen) loadCategories() {
	ctx := context.Background()
	input := expense_category.ListExpenseCategoriesInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
	}

	output, err := s.expenseCategoryService.List(ctx, input)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.categories = output.Categories
	s.table.Refresh()
}

func (s *ExpenseCategoryScreen) showCategoryDialog(existing *expense_category.ExpenseCategoryOutput) {
	d := NewExpenseCategoryFormDialog(
		s.window,
		s.expenseCategoryService,
		s.authManager,
		existing,
	)
	d.SetOnSaved(func() {
		s.loadCategories()
	})
	d.Show()
}

func (s *ExpenseCategoryScreen) showActionsMenu(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(s.categories) {
		return
	}
	c := s.categories[rowIndex]

	menu := fyne.NewMenu("Дії",
		fyne.NewMenuItem("Редагувати", func() {
			s.showCategoryDialog(c)
		}),
		fyne.NewMenuItem("Видалити", func() {
			dialog.ShowConfirm("Видалення", "Ви впевнені, що хочете видалити цю категорію?", func(ok bool) {
				if ok {
					s.deleteCategory(c.ID)
				}
			}, s.window)
		}),
	)

	widget.ShowPopUpMenuAtPosition(menu, s.window.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(s.table))
}

func (s *ExpenseCategoryScreen) deleteCategory(id int64) {
	ctx := context.Background()
	_, err := s.expenseCategoryService.Delete(ctx, expense_category.DeleteExpenseCategoryInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		CategoryID:    id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadCategories()
}
