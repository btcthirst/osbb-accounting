// presentation/fyne/screens/expenses_screen.go
package screens

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense"
)

// ExpensesScreen - екран списку витрат
type ExpensesScreen struct {
	window         fyne.Window
	expenseService *service.ExpenseService
	authManager    interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	}

	// UI
	content *fyne.Container
	table   *widget.Table

	// Data
	expenses []*expense.ExpenseOutput

	// Filters
	searchEntry *widget.Entry
}

// NewExpensesScreen створює новий екран
func NewExpensesScreen(
	window fyne.Window,
	expenseService *service.ExpenseService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *ExpensesScreen {
	s := &ExpensesScreen{
		window:         window,
		expenseService: expenseService,
		authManager:    authManager,
	}
	s.buildUI()
	return s
}

func (s *ExpensesScreen) buildUI() {
	// Buttons
	createButton := widget.NewButtonWithIcon("Додати витрату", theme.ContentAddIcon(), func() {
		s.showExpenseDialog(nil)
	})
	createButton.Importance = widget.HighImportance

	refreshButton := widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadExpenses()
	})

	// Filters
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("Пошук...")
	s.searchEntry.OnSubmitted = func(_ string) { s.loadExpenses() }

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.expenses) + 1, 7 // +1 for header
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell content")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				headers := []string{"ID", "Дата", "Категорія", "Сума", "Статус", "Підтверджено", "Дії"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			// Data
			if id.Row-1 >= len(s.expenses) {
				label.SetText("")
				return
			}
			e := s.expenses[id.Row-1]

			// Reset style
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0: // ID
				label.SetText(fmt.Sprintf("%d", e.ID))
			case 1: // Date
				label.SetText(e.ExpenseDate.Format("02.01.2006"))
			case 2: // Category
				label.SetText(e.CategoryName)
			case 3: // Amount
				label.SetText(fmt.Sprintf("%.2f грн", e.Amount))
			case 4: // Status
				label.SetText(e.PaymentStatusName)
			case 5: // Approved
				if e.IsApproved {
					label.SetText("Так")
				} else {
					label.SetText("Ні")
				}
			case 6: // Actions
				label.SetText("⚙️")
			}
		},
	)

	// Column widths
	s.table.SetColumnWidth(0, 50)  // ID
	s.table.SetColumnWidth(1, 100) // Date
	s.table.SetColumnWidth(2, 150) // Category
	s.table.SetColumnWidth(3, 100) // Amount
	s.table.SetColumnWidth(4, 120) // Status
	s.table.SetColumnWidth(5, 80)  // Approved
	s.table.SetColumnWidth(6, 50)  // Actions

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id)
		if id.Row == 0 {
			return
		}
		if id.Col == 6 {
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

	s.loadExpenses()
}

func (s *ExpensesScreen) Render() fyne.CanvasObject {
	return s.content
}

func (s *ExpensesScreen) loadExpenses() {
	ctx := context.Background()
	input := expense.ListExpensesInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
	}

	output, err := s.expenseService.List(ctx, input)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.expenses = output.Expenses
	s.table.Refresh()
}

func (s *ExpensesScreen) showExpenseDialog(existing *expense.ExpenseOutput) {
	d := NewExpenseFormDialog(
		s.window,
		s.expenseService,
		s.authManager,
		existing,
	)
	d.SetOnSaved(func() {
		s.loadExpenses()
	})
	d.Show()
}

func (s *ExpensesScreen) showActionsMenu(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(s.expenses) {
		return
	}
	e := s.expenses[rowIndex]

	menu := fyne.NewMenu("Дії",
		fyne.NewMenuItem("Редагувати", func() {
			if e.IsApproved {
				dialog.ShowInformation("Інформація", "Не можна редагувати затверджену витрату", s.window)
				return
			}
			s.showExpenseDialog(e)
		}),
		fyne.NewMenuItem("Видалити", func() {
			if e.IsApproved {
				dialog.ShowInformation("Інформація", "Не можна видалити затверджену витрату", s.window)
				return
			}
			dialog.ShowConfirm("Видалення", "Ви впевнені, що хочете видалити цю витрату?", func(ok bool) {
				if ok {
					s.deleteExpense(e.ID)
				}
			}, s.window)
		}),
		fyne.NewMenuItemSeparator(),
	)

	if e.IsApproved {
		menu.Items = append(menu.Items, fyne.NewMenuItem("Скасувати затвердження", func() {
			s.unapproveExpense(e.ID)
		}))
	} else {
		menu.Items = append(menu.Items, fyne.NewMenuItem("Затвердити", func() {
			s.approveExpense(e.ID)
		}))
	}

	widget.ShowPopUpMenuAtPosition(menu, s.window.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(s.table))
}

func (s *ExpensesScreen) deleteExpense(id int64) {
	ctx := context.Background()
	_, err := s.expenseService.Delete(ctx, expense.DeleteExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadExpenses()
}

func (s *ExpensesScreen) approveExpense(id int64) {
	ctx := context.Background()
	_, err := s.expenseService.Approve(ctx, expense.ApproveExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadExpenses()
}

func (s *ExpensesScreen) unapproveExpense(id int64) {
	ctx := context.Background()
	_, err := s.expenseService.Unapprove(ctx, expense.UnapproveExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadExpenses()
}
