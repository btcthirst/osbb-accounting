// presentation/fyne/screens/expenses_screen.go
package screens

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense"
	"osbb-accounting/presentation/fyne/common"
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
	createButton  *widget.Button
	refreshButton *widget.Button
	statsLabel    *widget.Label
	table         *widget.Table
	searchEntry   *widget.Entry
	filterSelect  *widget.Select

	// Data
	expenses         []*expense.ExpenseOutput
	filteredExpenses []*expense.ExpenseOutput
	currentPage      int
	pageSize         int
	totalCount       int64
}

type ExpenseFilterType string

const (
	ExpenseFilterAll         ExpenseFilterType = "Всі витрати"
	ExpenseFilterApproved    ExpenseFilterType = "Затверджені витрати"
	ExpenseFilterNotApproved ExpenseFilterType = "Не затверджені витрати"
)

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
		currentPage:    0,
		authManager:    authManager,
	}
	s.buildUI()
	s.loadExpenses()
	return s
}

func (s *ExpensesScreen) buildUI() {
	// Buttons
	s.createButton = newCreateButton("Додати витрату", func() {
		s.showExpenseDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadExpenses()
	})

	// Filters
	s.searchEntry = newSearchEntry("Пошук...", func(_ string) { s.applyFilters() })
	s.filterSelect = newFilterSelect([]string{string(ExpenseFilterAll), string(ExpenseFilterApproved), string(ExpenseFilterNotApproved)}, func(_ string) {
		s.applyFilters()
	})

	s.statsLabel = widget.NewLabel("")

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.filteredExpenses) + 1, 7 // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				renderTableHeader(label,
					[]string{"ID", "Дата", "Категорія", "Сума", "Статус", "Підтверджено", "Дії"}, id.Col)
				return
			}

			// Data
			if id.Row-1 >= len(s.filteredExpenses) {
				label.SetText("")
				return
			}
			e := s.filteredExpenses[id.Row-1]

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
				renderActionColumn(label)
			}
		},
	)

	// Column widths
	setupTableColumnWidths(s.table, map[int]float32{
		0: 50,  // ID
		1: 100, // Date
		2: 150, // Category
		3: 100, // Amount
		4: 120, // Status
		5: 80,  // Approved
		6: 50,  // Actions
	})

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id)
		if id.Row == 0 {
			return
		}
		if id.Col == 6 {
			s.showActionsMenu(id.Row - 1)
		}
	}

	s.filterSelect.SetSelected(string(ExpenseFilterAll))

	s.loadExpenses()
}

func (s *ExpensesScreen) Render() fyne.CanvasObject {
	toolbar := buildStandardToolbar(s.createButton, s.refreshButton, s.searchEntry, s.filterSelect)
	return buildStandardLayout(toolbar, s.table, s.statsLabel)
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
		common.ShowError(s.window, err)
		return
	}

	s.expenses = output.Expenses
	s.totalCount = output.Total
	s.applyFilters()
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

	actions := buildStandardActions(
		func() {
			if e.IsApproved {
				common.ShowInformation(s.window, "Інформація", "Не можна редагувати затверджену витрату")
				return
			}
			s.showExpenseDialog(e)
		},
		func() {
			if e.IsApproved {
				common.ShowInformation(s.window, "Інформація", "Не можна видалити затверджену витрату")
				return
			}
			s.confirmDelete(e)
		},
		nil,
	)

	if e.IsApproved {
		actions = append(actions, common.Action{
			Label: "↩️ Скасувати затвердження",
			OnTap: func() { s.unapproveExpense(e.ID) },
		})
	} else {
		actions = append(actions, common.Action{
			Label: "✅ Затвердити",
			OnTap: func() { s.approveExpense(e.ID) },
		})
	}

	common.ShowActionsMenu(s.window, fmt.Sprintf("Витрата #%d", e.ID), actions)
}

func (s *ExpensesScreen) confirmDelete(e *expense.ExpenseOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf("витрату #%d", e.ID), func() {
		s.deleteExpense(e.ID)
	})
}

func (s *ExpensesScreen) deleteExpense(id int64) {
	ctx := context.Background()
	_, err := s.expenseService.Delete(ctx, expense.DeleteExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Витрату успішно видалено")
	s.loadExpenses()
}

func (s *ExpensesScreen) approveExpense(id int64) {
	ctx := context.Background()
	_, err := s.expenseService.Approve(ctx, expense.ApproveExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Витрату затверджено")
	s.loadExpenses()
}

func (s *ExpensesScreen) unapproveExpense(id int64) {
	ctx := context.Background()
	_, err := s.expenseService.Unapprove(ctx, expense.UnapproveExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Затвердження витрати скасовано")
	s.loadExpenses()
}

func (s *ExpensesScreen) applyFilters() {
	s.filteredExpenses = make([]*expense.ExpenseOutput, 0)

	searchQuery := ""
	if s.searchEntry != nil {
		searchQuery = s.searchEntry.Text
	}
	filterType := ExpenseFilterAll
	if s.filterSelect != nil {
		filterType = ExpenseFilterType(s.filterSelect.Selected)
	}

	s.filteredExpenses = FilterList(
		s.expenses,
		searchQuery,
		string(filterType),
		func(e *expense.ExpenseOutput, query string) bool {
			return contains(e.CategoryName, query) || contains(fmt.Sprintf("%.2f", e.Amount), query)
		},
		func(e *expense.ExpenseOutput, filter string) bool {
			switch ExpenseFilterType(filter) {
			case ExpenseFilterAll:
				return true
			case ExpenseFilterApproved:
				return e.IsApproved
			case ExpenseFilterNotApproved:
				return !e.IsApproved
			}
			return true
		},
	)
	s.table.Refresh()
	s.updateStats()
}

func (s *ExpensesScreen) getStatsText() string {
	return fmt.Sprintf("Показано:%d з %d витрат", len(s.filteredExpenses), s.totalCount)
}

// updateStats оновлює статистику.
func (s *ExpensesScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
