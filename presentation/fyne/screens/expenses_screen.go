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
	"osbb-accounting/presentation/fyne/dialogs"
	"osbb-accounting/presentation/fyne/text"
	"time"
)

// ExpensesScreen - екран списку витрат
type ExpensesScreen struct {
	window                 fyne.Window
	expenseService         *service.ExpenseService
	expenseCategoryService *service.ExpenseCategoryService
	authManager            interface {
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
	ExpenseFilterAll         ExpenseFilterType = text.FilterExpenseAll
	ExpenseFilterApproved    ExpenseFilterType = text.FilterExpenseApproved
	ExpenseFilterNotApproved ExpenseFilterType = text.FilterExpenseNotApproved
)

// NewExpensesScreen створює новий екран
func NewExpensesScreen(
	window fyne.Window,
	expenseService *service.ExpenseService,
	expenseCategoryService *service.ExpenseCategoryService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *ExpensesScreen {
	s := &ExpensesScreen{
		window:                 window,
		expenseService:         expenseService,
		expenseCategoryService: expenseCategoryService,
		currentPage:            0,
		authManager:            authManager,
	}
	s.buildUI()
	s.loadExpenses()
	return s
}

func (s *ExpensesScreen) buildUI() {
	// Buttons
	s.createButton = newCreateButton(text.ActionAdd, func() {
		s.showExpenseDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadExpenses()
	})

	// Filters
	s.searchEntry = newSearchEntry(text.SearchPlaceholder, func(_ string) { s.applyFilters() })
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
				renderTableHeader(label, text.ExpensesTableHeaders, id.Col)
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
					label.SetText(text.LabelYes)
				} else {
					label.SetText(text.LabelNo)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	input := expense.ListExpensesInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         1000, // Load all for client-side filtering
		SearchQuery:   "",   // We filter on client side for now to match other screens
	}

	output, err := s.expenseService.List(ctx, input)
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorLoading, err))
		return
	}

	s.expenses = output.Expenses
	s.totalCount = output.Total
	s.applyFilters()
}

func (s *ExpensesScreen) showExpenseDialog(existing *expense.ExpenseOutput) {
	d := dialogs.NewExpenseFormDialog(
		s.window,
		s.expenseService,
		s.expenseCategoryService,
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
				common.ShowInformation(s.window, "Інформація", text.MsgErrorEditApproved)
				return
			}
			s.showExpenseDialog(e)
		},
		func() {
			if e.IsApproved {
				common.ShowInformation(s.window, "Інформація", text.MsgErrorDeleteApproved)
				return
			}
			s.confirmDelete(e)
		},
		nil,
		nil,
	)

	if e.IsApproved {
		actions = append(actions, common.Action{
			Label: text.ActionUnapprove,
			OnTap: func() { s.unapproveExpense(e.ID) },
		})
	} else {
		actions = append(actions, common.Action{
			Label: text.ActionApprove,
			OnTap: func() { s.approveExpense(e.ID) },
		})
	}

	common.ShowActionsMenu(s.window, fmt.Sprintf(text.TitleExpenseActions, e.ID), actions)
}

func (s *ExpensesScreen) confirmDelete(e *expense.ExpenseOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf(text.LabelExpenseID, e.ID), func() {
		s.deleteExpense(e.ID)
	})
}

func (s *ExpensesScreen) deleteExpense(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.expenseService.Delete(ctx, expense.DeleteExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorDeleting, err))
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessExpenseDeleted)
	s.loadExpenses()
}

func (s *ExpensesScreen) approveExpense(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.expenseService.Approve(ctx, expense.ApproveExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessExpenseApproved)
	s.loadExpenses()
}

func (s *ExpensesScreen) unapproveExpense(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.expenseService.Unapprove(ctx, expense.UnapproveExpenseInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ExpenseID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessExpenseUnapproved)
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
