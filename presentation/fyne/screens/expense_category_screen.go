// presentation/fyne/screens/expense_category_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/dialogs"
	"osbb-accounting/presentation/fyne/text"
)

// ExpenseCategoryScreen - екран списку категорій витрат
type ExpenseCategoryScreen struct {
	window          fyne.Window
	categoryService *service.ExpenseCategoryService
	authManager     interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	}

	// UI
	searchEntry   *widget.Entry
	filterSelect  *widget.Select
	createButton  *widget.Button
	refreshButton *widget.Button
	table         *widget.Table
	statsLabel    *widget.Label

	// Data
	categories         []*expense_category.ExpenseCategoryOutput
	filteredCategories []*expense_category.ExpenseCategoryOutput
	currentPage        int
	pageSize           int
	totalCount         int64
}

type ExpenseCategoryFilterType string

const (
	ExpenseCategoryFilterAll      ExpenseCategoryFilterType = text.FilterCategoryAll
	ExpenseCategoryFilterActive   ExpenseCategoryFilterType = text.FilterCategoryActive
	ExpenseCategoryFilterInactive ExpenseCategoryFilterType = text.FilterCategoryInactive
)

// NewExpenseCategoryScreen створює новий екран
func NewExpenseCategoryScreen(
	window fyne.Window,
	categoryService *service.ExpenseCategoryService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *ExpenseCategoryScreen {
	s := &ExpenseCategoryScreen{
		window:          window,
		categoryService: categoryService,
		authManager:     authManager,
	}
	s.buildUI()
	return s
}

func (s *ExpenseCategoryScreen) buildUI() {
	// Buttons
	s.createButton = newCreateButton(text.ActionAdd, func() {
		s.showCategoryDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadCategories()
	})

	// Filters
	s.searchEntry = newSearchEntry(text.SearchPlaceholder, func(_ string) { s.applyFilters() })
	s.filterSelect = newFilterSelect([]string{string(ExpenseCategoryFilterAll), string(ExpenseCategoryFilterActive), string(ExpenseCategoryFilterInactive)}, func(_ string) { s.applyFilters() })

	s.statsLabel = widget.NewLabel("")
	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.filteredCategories) + 1, 5 // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				renderTableHeader(label, text.ExpenseCategoriesTableHeaders, id.Col)
				return
			}

			// Data
			if id.Row-1 >= len(s.filteredCategories) {
				label.SetText("")
				return
			}
			c := s.filteredCategories[id.Row-1]

			// Reset style
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0: // Name
				label.SetText(c.Name)
			case 1: // Type
				label.SetText(c.TypeDisplayName)
			case 2: // Description
				label.SetText(ptrToString(c.Description, "-"))
			case 3: // Status
				label.SetText(activeStatus(c.IsActive))
			case 4: // Actions
				renderActionColumn(label)
			}
		},
	)

	// Column widths
	setupTableColumnWidths(s.table, map[int]float32{
		0: 200, // Name
		1: 150, // Type
		2: 300, // Description
		3: 100, // Status
		4: 50,  // Actions
	})

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id)
		if id.Row == 0 {
			return
		}
		if id.Col == 4 {
			s.showActionsMenu(id.Row - 1)
		}
	}

	s.filterSelect.SetSelected(string(ExpenseCategoryFilterAll))
	s.loadCategories()
}

func (s *ExpenseCategoryScreen) Render() fyne.CanvasObject {
	toolbar := buildStandardToolbar(s.createButton, s.refreshButton, s.searchEntry, s.filterSelect)
	return buildStandardLayout(toolbar, s.table, s.statsLabel)
}

func (s *ExpenseCategoryScreen) loadCategories() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	input := expense_category.ListExpenseCategoriesInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
	}

	output, err := s.categoryService.List(ctx, input)
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorCategoryLoading, err))
		return
	}

	s.categories = output.Categories
	s.totalCount = output.Total
	s.applyFilters()
}

func (s *ExpenseCategoryScreen) showCategoryDialog(existing *expense_category.ExpenseCategoryOutput) {
	d := dialogs.NewExpenseCategoryFormDialog(
		s.window,
		s.categoryService,
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

	actions := buildStandardActions(
		func() { s.showCategoryDialog(c) },
		func() { s.confirmDelete(c) },
		nil,
		nil,
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf(text.TitleCategoryActions, c.Name), actions)
}

func (s *ExpenseCategoryScreen) confirmDelete(c *expense_category.ExpenseCategoryOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf(text.LabelCategoryName, c.Name), func() {
		s.deleteCategory(c.ID)
	})
}

func (s *ExpenseCategoryScreen) deleteCategory(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.categoryService.Delete(ctx, expense_category.DeleteExpenseCategoryInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		CategoryID:    id,
	})
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorDeleting, err))
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessCategoryDeleted)
	s.loadCategories()
}

func (s *ExpenseCategoryScreen) applyFilters() {

	s.filteredCategories = make([]*expense_category.ExpenseCategoryOutput, 0)
	searchQuery := ""
	if s.searchEntry != nil {
		searchQuery = s.searchEntry.Text
	}

	s.filteredCategories = FilterList(
		s.categories,
		searchQuery,
		s.filterSelect.Selected,
		func(c *expense_category.ExpenseCategoryOutput, query string) bool {
			return contains(c.Name, query)
		},
		func(c *expense_category.ExpenseCategoryOutput, filter string) bool {
			switch ExpenseCategoryFilterType(filter) {
			case ExpenseCategoryFilterAll:
				return true
			case ExpenseCategoryFilterActive:
				return c.IsActive
			case ExpenseCategoryFilterInactive:
				return !c.IsActive
			}
			return true
		},
	)
	s.table.Refresh()
	s.updateStats()
}

func (s *ExpenseCategoryScreen) getStatsText() string {
	return fmt.Sprintf("Показано: %d з %d категорій", len(s.filteredCategories), s.totalCount)
}

// updateStats оновлює статистику.
func (s *ExpenseCategoryScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
