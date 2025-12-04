// presentation/fyne/screens/owners_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/owner"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/dialogs"
	"osbb-accounting/presentation/fyne/text"
)

// OwnersScreen представляє екран управління власниками.
type OwnersScreen struct {
	window       fyne.Window
	ownerService service.OwnerServiceInterface
	authManager  *auth.AuthManager

	// UI елементи
	searchEntry   *widget.Entry
	filterSelect  *widget.Select
	ownersTable   *widget.Table
	createButton  *widget.Button
	refreshButton *widget.Button
	statsLabel    *widget.Label

	// Дані
	owners         []*owner.OwnerOutput
	filteredOwners []*owner.OwnerOutput
	currentPage    int
	pageSize       int
	totalCount     int64
}

type OwnerFilterType string

const (
	OwnerFilterAll          OwnerFilterType = text.FilterOwnerAll
	OwnerFilterWithTax      OwnerFilterType = text.FilterOwnerWithTax
	OwnerFilterNoTax        OwnerFilterType = text.FilterOwnerNoTax
	OwnerFilterWithContacts OwnerFilterType = text.FilterOwnerWithContacts
	OwnerFilterNoContacts   OwnerFilterType = text.FilterOwnerNoContacts
	OwnerFilterActive       OwnerFilterType = text.FilterOwnerActive
	OwnerFilterInactive     OwnerFilterType = text.FilterOwnerInactive
)

// NewOwnersScreen створює новий екран власників.
func NewOwnersScreen(
	window fyne.Window,
	ownerService service.OwnerServiceInterface,
	authManager *auth.AuthManager,
) *OwnersScreen {
	screen := &OwnersScreen{
		window:       window,
		ownerService: ownerService,
		authManager:  authManager,
		pageSize:     50,
		currentPage:  0,
	}

	screen.buildUI()
	screen.loadOwners()

	return screen
}

// buildUI створює інтерфейс екрану.
func (s *OwnersScreen) buildUI() {
	// Пошук
	s.searchEntry = newSearchEntry(text.SearchPlaceholder, func(query string) {
		s.applyFilters()
	})

	// Фільтр
	// Фільтр
	s.filterSelect = newFilterSelect([]string{
		string(OwnerFilterAll),
		string(OwnerFilterWithTax),
		string(OwnerFilterNoTax),
		string(OwnerFilterWithContacts),
		string(OwnerFilterNoContacts),
		string(OwnerFilterActive),
		string(OwnerFilterInactive),
	}, func(value string) {
		s.applyFilters()
	})

	// Кнопка створення
	s.createButton = newCreateButton(text.ActionAdd, func() {
		s.showCreateDialog()
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadOwners()
	})

	// Статистика
	s.statsLabel = widget.NewLabel("")

	// Таблиця
	s.buildTable()

	// Перенесено вниз щоб уникнути nil pointer
	s.filterSelect.SetSelected(string(OwnerFilterAll))
	s.loadOwners()
}

// buildTable створює таблицю з власниками.
func (s *OwnersScreen) buildTable() {
	s.ownersTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredOwners) + 1, 6 // +1 для заголовків, 6 колонок
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				renderTableHeader(label, text.OwnersTableHeaders, id.Col)
				return
			}

			// Дані
			if id.Row-1 >= len(s.filteredOwners) {
				label.SetText("")
				return
			}

			owner := s.filteredOwners[id.Row-1]

			switch id.Col {
			case 0: // №
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1: // ПІБ
				label.SetText(owner.FullName)
				if !owner.IsActive {
					label.TextStyle = fyne.TextStyle{Italic: true}
				}
			case 2: // Телефон
				if owner.Phone != nil {
					label.SetText(*owner.Phone)
				} else {
					label.SetText("—")
				}
			case 3: // Email
				if owner.Email != nil {
					label.SetText(*owner.Email)
				} else {
					label.SetText("—")
				}
			case 4: // ІПН
				if owner.TaxNumber != nil {
					label.SetText(*owner.TaxNumber)
				} else {
					label.SetText("—")
				}
			case 5: // Дії
				renderActionColumn(label)
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(s.ownersTable, map[int]float32{
		0: 50,  // №
		1: 250, // ПІБ
		2: 150, // Телефон
		3: 200, // Email
		4: 120, // ІПН
		5: 80,  // Дії
	})

	// Клік на комірку
	s.ownersTable.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return // Заголовок
		}
		if id.Row-1 >= len(s.filteredOwners) {
			return
		}

		owner := s.filteredOwners[id.Row-1]

		if id.Col == 5 {
			// Клік на "Дії"
			s.showActionsMenu(owner)
		} else {
			// Клік на інші колонки - показати деталі
			s.showOwnerDetails(owner)
		}
	}
}

// Render повертає контейнер з UI екрану.
func (s *OwnersScreen) Render() fyne.CanvasObject {
	// Toolbar
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(s.createButton, s.refreshButton),
		nil,
		container.NewVBox(
			s.searchEntry,
			s.filterSelect,
		),
	)

	// Головний контейнер
	content := container.NewBorder(
		toolbar,
		s.statsLabel,
		nil,
		nil,
		s.ownersTable,
	)

	return content
}

// getStatsText повертає текст статистики.
func (s *OwnersScreen) getStatsText() string {
	return fmt.Sprintf("Показано: %d з %d власників", len(s.filteredOwners), s.totalCount)
}

// updateStats оновлює статистику.
func (s *OwnersScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}

// loadOwners завантажує список власників.
func (s *OwnersScreen) loadOwners() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := s.ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         1000, // Завантажуємо всіх для клієнтської фільтрації
		Offset:        0,
		OrderBy:       "name",
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorOwnerLoading, err))
		return
	}

	s.owners = output.Owners
	s.totalCount = output.Total
	s.applyFilters()
}

// applyFilters застосовує фільтри до списку.
func (s *OwnersScreen) applyFilters() {
	s.filteredOwners = make([]*owner.OwnerOutput, 0)

	searchQuery := s.searchEntry.Text
	filterType := s.filterSelect.Selected

	s.filteredOwners = FilterList(
		s.owners,
		searchQuery,
		filterType,
		func(o *owner.OwnerOutput, query string) bool {
			return contains(o.FullName, query) ||
				(o.Phone != nil && contains(*o.Phone, query)) ||
				(o.Email != nil && contains(*o.Email, query)) ||
				(o.TaxNumber != nil && contains(*o.TaxNumber, query))
		},
		func(o *owner.OwnerOutput, filter string) bool {
			switch OwnerFilterType(filter) {
			case OwnerFilterWithTax:
				return o.TaxNumber != nil && *o.TaxNumber != ""
			case OwnerFilterNoTax:
				return o.TaxNumber == nil || *o.TaxNumber == ""
			case OwnerFilterWithContacts:
				return (o.Phone != nil && *o.Phone != "") || (o.Email != nil && *o.Email != "")
			case OwnerFilterNoContacts:
				return (o.Phone == nil || *o.Phone == "") && (o.Email == nil || *o.Email == "")
			case OwnerFilterActive:
				return o.IsActive
			case OwnerFilterInactive:
				return !o.IsActive
			}
			return true
		},
	)

	s.ownersTable.Refresh()
	s.updateStats()
}

// showOwnerDetails показує детальну інформацію про власника.
func (s *OwnersScreen) showOwnerDetails(o *owner.OwnerOutput) {
	details := fmt.Sprintf(
		"ПІБ: %s\n"+
			"Коротко: %s\n\n"+
			"Телефон: %s\n"+
			"Email: %s\n"+
			"ІПН: %s\n\n"+
			"Статус: %s\n",
		o.FullName,
		o.ShortName,
		ptrToString(o.Phone, "не вказано"),
		ptrToString(o.Email, "не вказано"),
		ptrToString(o.TaxNumber, "не вказано"),
		activeStatus(o.IsActive),
	)

	if o.PassportSeries != nil || o.PassportNumber != nil {
		details += fmt.Sprintf("\nПаспорт: %s %s\n",
			ptrToString(o.PassportSeries, ""),
			ptrToString(o.PassportNumber, ""),
		)
	}

	if o.RegisteredAddress != nil {
		details += fmt.Sprintf("\nАдреса реєстрації: %s\n", *o.RegisteredAddress)
	}

	if o.ActualAddress != nil {
		details += fmt.Sprintf("Фактична адреса: %s\n", *o.ActualAddress)
	}

	if o.Notes != nil {
		details += fmt.Sprintf("\nПримітки: %s\n", *o.Notes)
	}

	common.ShowInformation(s.window, text.TitleOwnerDetails, details)
}

// showActionsMenu показує меню дій з власником.
func (s *OwnersScreen) showActionsMenu(o *owner.OwnerOutput) {
	actions := buildStandardActions(
		func() { s.showOwnerDialog(o) },
		func() { s.confirmDelete(o) },
		func() { s.showOwnerDetails(o) },
		nil,
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf(text.TitleOwnerActions, o.FullName), actions)
}

// showCreateDialog показує діалог створення власника.
func (s *OwnersScreen) showCreateDialog() {
	s.showOwnerDialog(nil)
}

// showOwnerDialog показує діалог створення/редагування власника.
func (s *OwnersScreen) showOwnerDialog(existing *owner.OwnerOutput) {
	d := dialogs.NewOwnerFormDialog(
		s.window,
		s.ownerService,
		s.authManager,
		existing,
	)
	d.OnSuccess = func() {
		s.loadOwners()
	}
	d.Show()
}

// confirmDelete підтверджує видалення.
func (s *OwnersScreen) confirmDelete(o *owner.OwnerOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf(text.LabelOwnerName, o.FullName), func() {
		s.deleteOwner(o)
	})
}

// deleteOwner видаляє власника.
func (s *OwnersScreen) deleteOwner(o *owner.OwnerOutput) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.ownerService.Delete(ctx, owner.DeleteOwnerInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		OwnerID:       o.ID,
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка видалення: %v", err))
		return
	}

	common.ShowSuccess(s.window, fmt.Sprintf(text.MsgSuccessOwnerDeleted, o.FullName))
	s.loadOwners()
}
