// presentation/fyne/screens/expense_category_screen.go
package screens

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/expense_category"
	"osbb-accounting/presentation/fyne/common"
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
	s.createButton = newCreateButton("Додати категорію", func() {
		s.showCategoryDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadCategories()
	})

	// Filters
	s.searchEntry = newSearchEntry("Пошук (Назва)...", func(_ string) { s.applyFilters() })
	s.filterSelect = newFilterSelect([]string{"Всі", "Активні", "Неактивні"}, func(_ string) { s.applyFilters() })

	s.statsLabel = widget.NewLabel("")
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
				label.SetText(ptrToString(c.Description, "-"))
			case 3: // Status
				label.SetText(activeStatus(c.IsActive))
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

	s.filterSelect.SetSelectedIndex(0)
	s.loadCategories()
}

func (s *ExpenseCategoryScreen) Render() fyne.CanvasObject {
	// Toolbar container
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(s.createButton, s.refreshButton),
		nil,
		container.NewVBox(s.searchEntry, s.filterSelect),
	)

	// Layout
	content := container.NewBorder(
		toolbar,
		s.statsLabel,
		nil, nil,
		s.table,
	)
	return content
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
		common.ShowError(s.window, err)
		return
	}

	s.categories = output.Categories
	s.totalCount = output.Total
	s.table.Refresh()
	s.updateStats()
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

	actions := []common.Action{
		{
			Label: "✏️ Редагувати",
			OnTap: func() { s.showCategoryDialog(c) },
		},
		{
			Label:      "🗑️ Видалити",
			OnTap:      func() { s.confirmDelete(c) },
			Importance: widget.DangerImportance,
		},
	}

	common.ShowActionsMenu(s.window, "Дії з категорією: "+c.Name, actions)
}

func (s *ExpenseCategoryScreen) confirmDelete(c *expense_category.ExpenseCategoryOutput) {
	common.ShowDeleteConfirmation(
		s.window,
		"Підтвердження видалення",
		"Ви впевнені, що хочете видалити категорію '"+c.Name+"'?",
		func() {
			s.deleteCategory(c.ID)
		},
	)
}

func (s *ExpenseCategoryScreen) deleteCategory(id int64) {
	ctx := context.Background()
	_, err := s.expenseCategoryService.Delete(ctx, expense_category.DeleteExpenseCategoryInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		CategoryID:    id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Категорію успішно видалено")
	s.loadCategories()
}

func (s *ExpenseCategoryScreen) applyFilters() {

	s.filteredCategories = make([]*expense_category.ExpenseCategoryOutput, 0)
	for _, c := range s.categories {

		switch s.filterSelect.Selected {
		case "Всі":
			s.filteredCategories = append(s.filteredCategories, c)
		case "Активні":
			if c.IsActive {
				s.filteredCategories = append(s.filteredCategories, c)
			}
		case "Неактивні":
			if !c.IsActive {
				s.filteredCategories = append(s.filteredCategories, c)
			}
		}
	}
	s.table.Refresh()
	s.updateStats()
}

func (s *ExpenseCategoryScreen) getStatsText() string {
	return fmt.Sprintf("Показано: %d з %d категорій", len(s.categories), s.totalCount)
}

// updateStats оновлює статистику.
func (s *ExpenseCategoryScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
