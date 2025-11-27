// presentation/fyne/screens/contractors_screen.go
package screens

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/contractor"
	"osbb-accounting/presentation/fyne/common"
)

// ContractorsScreen - екран списку контрагентів
type ContractorsScreen struct {
	window            fyne.Window
	contractorService *service.ContractorService
	authManager       interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	}

	// UI
	table         *widget.Table
	statsLabel    *widget.Label
	createButton  *widget.Button
	refreshButton *widget.Button
	searchEntry   *widget.Entry
	filterSelect  *widget.Select

	// Data
	contractors         []*contractor.ContractorOutput
	filteredContractors []*contractor.ContractorOutput
	currentPage         int
	pageSize            int
	totalCount          int64
}

type ContractorFilterType string

const (
	ContractorFilterAll      ContractorFilterType = "Всі контрагенти"
	ContractorFilterActive   ContractorFilterType = "Активні"
	ContractorFilterInactive ContractorFilterType = "Неактивні"
)

// NewContractorsScreen створює новий екран
func NewContractorsScreen(
	window fyne.Window,
	contractorService *service.ContractorService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *ContractorsScreen {
	s := &ContractorsScreen{
		window:            window,
		contractorService: contractorService,
		authManager:       authManager,
	}
	s.buildUI()
	return s
}

func (s *ContractorsScreen) buildUI() {
	// Buttons
	s.createButton = newCreateButton("Додати контрагента", func() {
		s.showContractorDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadContractors()
	})

	// Stats
	s.statsLabel = widget.NewLabel("")

	// Filters
	s.searchEntry = newSearchEntry("Пошук (Назва, ЄДРПОУ)...", func(_ string) { s.applyFilters() })
	s.filterSelect = newFilterSelect([]string{string(ContractorFilterAll), string(ContractorFilterActive), string(ContractorFilterInactive)}, func(value string) { s.applyFilters() })

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.filteredContractors) + 1, 6 // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				renderTableHeader(label,
					[]string{"Назва", "Тип", "ЄДРПОУ", "Контакт", "Статус", "Дії"},
					id.Col)
				return
			}

			// Data
			if id.Row-1 >= len(s.filteredContractors) {
				label.SetText("")
				return
			}
			c := s.contractors[id.Row-1]

			// Reset style
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0: // Name
				label.SetText(c.Name)
			case 1: // Type
				label.SetText(c.TypeDisplayName)
			case 2: // EDRPOU
				label.SetText(ptrToString(c.EDRPOU, "-"))
			case 3: // Contact
				label.SetText(c.ContactInfo)
			case 4: // Status
				label.SetText(activeStatus(c.IsActive))
			case 5: // Actions
				renderActionColumn(label)
			}
		},
	)

	// Column widths
	setupTableColumnWidths(s.table, map[int]float32{
		0: 200, // Name
		1: 150, // Type
		2: 100, // EDRPOU
		3: 250, // Contact
		4: 100, // Status
		5: 50,  // Actions
	})

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id)
		if id.Row == 0 {
			return
		}
		if id.Col == 5 {
			s.showActionsMenu(id.Row - 1)
		}
	}
	s.loadContractors()
	s.filterSelect.SetSelected(string(ContractorFilterAll))
}

func (s *ContractorsScreen) Render() fyne.CanvasObject {
	toolbar := buildStandardToolbar(s.createButton, s.refreshButton, s.searchEntry, s.filterSelect)
	return buildStandardLayout(toolbar, s.table, s.statsLabel)
}

func (s *ContractorsScreen) loadContractors() {
	ctx := context.Background()
	input := contractor.ListContractorsInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
	}

	output, err := s.contractorService.List(ctx, input)
	if err != nil {
		common.ShowError(s.window, err)
		return
	}

	s.contractors = output.Contractors
	s.totalCount = output.Total

	s.applyFilters()
}

func (s *ContractorsScreen) applyFilters() {
	s.filteredContractors = make([]*contractor.ContractorOutput, 0)
	// Безпечне отримання значень UI
	searchQuery := ""
	if s.searchEntry != nil {
		searchQuery = s.searchEntry.Text
	}
	filterType := ""
	if s.filterSelect != nil {
		filterType = s.filterSelect.Selected
	}

	// Фільтрація
	s.filteredContractors = FilterList(
		s.contractors,
		searchQuery,
		filterType,
		func(c *contractor.ContractorOutput, query string) bool {
			return strings.Contains(strings.ToLower(c.Name), strings.ToLower(query))
		},
		func(c *contractor.ContractorOutput, filter string) bool {
			switch ContractorFilterType(filter) {
			case ContractorFilterActive:
				return c.IsActive
			case ContractorFilterInactive:
				return !c.IsActive
			case ContractorFilterAll:
				return true
			default:
				// Якщо це не статус, то це тип контрагента
				return filter == "" || filter == c.TypeDisplayName
			}
		},
	)
	s.table.Refresh()
	s.updateStats()
}

func (s *ContractorsScreen) showContractorDialog(existing *contractor.ContractorOutput) {
	d := NewContractorFormDialog(
		s.window,
		s.contractorService,
		s.authManager,
		existing,
	)
	d.SetOnSaved(func() {
		s.loadContractors()
	})
	d.Show()
}

func (s *ContractorsScreen) showActionsMenu(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(s.contractors) {
		return
	}
	c := s.contractors[rowIndex]

	actions := []common.Action{
		{
			Label: "✏️ Редагувати",
			OnTap: func() { s.showContractorDialog(c) },
		},
		{
			Label:      "🗑️ Видалити",
			OnTap:      func() { s.confirmDelete(c) },
			Importance: widget.DangerImportance,
		},
	}

	common.ShowActionsMenu(s.window, "Дії з контрагентом: "+c.Name, actions)
}

func (s *ContractorsScreen) confirmDelete(c *contractor.ContractorOutput) {
	common.ShowDeleteConfirmation(
		s.window,
		"Підтвердження видалення",
		"Ви впевнені, що хочете видалити контрагента '"+c.Name+"'?",
		func() {
			s.deleteContractor(c.ID)
		},
	)
}

func (s *ContractorsScreen) deleteContractor(id int64) {
	ctx := context.Background()
	_, err := s.contractorService.Delete(ctx, contractor.DeleteContractorInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ContractorID:  id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Контрагента успішно видалено")
	s.loadContractors()
}

func (s *ContractorsScreen) getStatsText() string {
	return fmt.Sprintf("Показано: %d з %d контрагентів", len(s.filteredContractors), s.totalCount)
}

// updateStats оновлює статистику.
func (s *ContractorsScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
