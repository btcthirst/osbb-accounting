// presentation/fyne/screens/contractors_screen.go
package screens

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/contractor"
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
	content *fyne.Container
	table   *widget.Table

	// Data
	contractors []*contractor.ContractorOutput

	// Filters
	searchEntry *widget.Entry
}

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
	// Toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			s.showContractorDialog(nil)
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			s.loadContractors()
		}),
	)

	// Filters
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("Пошук (Назва, ЄДРПОУ)...")
	s.searchEntry.OnSubmitted = func(_ string) { s.loadContractors() }

	filterContainer := container.NewBorder(nil, nil, nil,
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.loadContractors() }),
		s.searchEntry,
	)

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.contractors), 6 // Cols: Name, Type, EDRPOU, Contact, Status, Actions
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell content")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			c := s.contractors[id.Row]
			label := cell.(*widget.Label)

			switch id.Col {
			case 0: // Name
				label.SetText(c.Name)
			case 1: // Type
				label.SetText(c.TypeDisplayName)
			case 2: // EDRPOU
				if c.EDRPOU != nil {
					label.SetText(*c.EDRPOU)
				} else {
					label.SetText("-")
				}
			case 3: // Contact
				label.SetText(c.ContactInfo)
			case 4: // Status
				if c.IsActive {
					label.SetText("Активний")
				} else {
					label.SetText("Неактивний")
				}
			case 5: // Actions
				label.SetText("⚙️")
			}
		},
	)

	// Column widths
	s.table.SetColumnWidth(0, 200) // Name
	s.table.SetColumnWidth(1, 150) // Type
	s.table.SetColumnWidth(2, 100) // EDRPOU
	s.table.SetColumnWidth(3, 250) // Contact
	s.table.SetColumnWidth(4, 100) // Status
	s.table.SetColumnWidth(5, 50)  // Actions

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id)
		if id.Col == 5 {
			s.showActionsMenu(id.Row)
		}
	}

	// Layout
	s.content = container.NewBorder(
		container.NewVBox(toolbar, filterContainer),
		nil, nil, nil,
		s.table,
	)

	s.loadContractors()
}

func (s *ContractorsScreen) Render() fyne.CanvasObject {
	return s.content
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
		dialog.ShowError(err, s.window)
		return
	}

	s.contractors = output.Contractors
	s.table.Refresh()
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

	menu := fyne.NewMenu("Дії",
		fyne.NewMenuItem("Редагувати", func() {
			s.showContractorDialog(c)
		}),
		fyne.NewMenuItem("Видалити", func() {
			dialog.ShowConfirm("Видалення", "Ви впевнені, що хочете видалити цього контрагента?", func(ok bool) {
				if ok {
					s.deleteContractor(c.ID)
				}
			}, s.window)
		}),
	)

	widget.ShowPopUpMenuAtPosition(menu, s.window.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(s.table))
}

func (s *ContractorsScreen) deleteContractor(id int64) {
	ctx := context.Background()
	_, err := s.contractorService.Delete(ctx, contractor.DeleteContractorInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ContractorID:  id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadContractors()
}
