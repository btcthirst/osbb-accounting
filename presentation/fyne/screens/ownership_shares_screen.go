// presentation/fyne/screens/ownership_shares_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/ownership"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/common"
)

// OwnershipSharesScreen представляє екран управління частками власності.
type OwnershipSharesScreen struct {
	window           fyne.Window
	apartmentService service.ApartmentServiceInterface
	ownerService     service.OwnerServiceInterface
	ownershipService service.OwnershipServiceInterface
	authManager      *auth.AuthManager

	// UI елементи
	searchEntry   *widget.Entry
	filterSelect  *widget.Select
	sharesTable   *widget.Table
	createButton  *widget.Button
	refreshButton *widget.Button
	statsLabel    *widget.Label

	// Дані
	shares         []*ownership.OwnershipShareDetailsOutput
	filteredShares []*ownership.OwnershipShareDetailsOutput
	currentPage    int
	pageSize       int
	totalCount     int64
}

type OwnershipShareFilterType string

const (
	OwnershipShareFilterAll      OwnershipShareFilterType = "Всі частки"
	OwnershipShareFilterActive   OwnershipShareFilterType = "Тільки активні зараз"
	OwnershipShareFilterFull     OwnershipShareFilterType = "Повна власність"
	OwnershipShareFilterShared   OwnershipShareFilterType = "Часткова власність"
	OwnershipShareFilterRent     OwnershipShareFilterType = "Оренда"
	OwnershipShareFilterFinished OwnershipShareFilterType = "Завершені"
)

// NewOwnershipSharesScreen створює новий екран часток власності.
func NewOwnershipSharesScreen(
	window fyne.Window,
	apartmentService service.ApartmentServiceInterface,
	ownerService service.OwnerServiceInterface,
	ownershipService service.OwnershipServiceInterface,
	authManager *auth.AuthManager,
) *OwnershipSharesScreen {
	screen := &OwnershipSharesScreen{
		window:           window,
		apartmentService: apartmentService,
		ownerService:     ownerService,
		ownershipService: ownershipService,
		authManager:      authManager,
		pageSize:         50,
		currentPage:      0,
	}

	screen.buildUI()
	screen.loadShares()

	return screen
}

// buildUI створює інтерфейс екрану.
func (s *OwnershipSharesScreen) buildUI() {
	// 1. Спочатку ініціалізуємо всі віджети, щоб уникнути nil pointer у колбеках
	// Пошук
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("🔍 Пошук за власником або квартирою...")
	s.searchEntry.OnChanged = func(query string) {
		s.applyFilters()
	}

	// Фільтр за типом
	s.filterSelect = newFilterSelect([]string{
		string(OwnershipShareFilterAll),
		string(OwnershipShareFilterActive),
		string(OwnershipShareFilterFull),
		string(OwnershipShareFilterShared),
		string(OwnershipShareFilterRent),
		string(OwnershipShareFilterFinished),
	}, func(value string) {
		s.applyFilters()
	})

	// Кнопка створення
	s.createButton = newCreateButton("Додати частку", func() {
		s.showCreateDialog()
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadShares()
	})

	// Статистика
	s.statsLabel = widget.NewLabel("")

	// Таблиця
	s.buildTable()

	// 2. Встановлюємо дефолтні значення, які можуть тригерити колбеки
	// Робимо це в самому кінці, коли всі поля (searchEntry, apartmentFilter, ownerFilter) вже існують
	s.filterSelect.SetSelected(string(OwnershipShareFilterAll))
	s.loadShares()
}

// buildTable створює таблицю з частками власності.
func (s *OwnershipSharesScreen) buildTable() {
	s.sharesTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredShares) + 1, 7 // +1 для заголовків, 7 колонок
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				renderTableHeader(label,
					[]string{"№", "Власник", "Квартира", "Частка", "Тип", "Дати", "Дії"}, id.Col)
				return
			}

			// Дані
			if id.Row-1 >= len(s.filteredShares) {
				label.SetText("")
				return
			}

			share := s.filteredShares[id.Row-1]

			switch id.Col {
			case 0: // №
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1: // Власник
				label.SetText(share.OwnerName)
				if !share.IsCurrentlyActive {
					label.TextStyle = fyne.TextStyle{Italic: true}
				}
			case 2: // Квартира
				label.SetText(fmt.Sprintf("Кв. %s", share.ApartmentNumber))
			case 3: // Частка
				label.SetText(fmt.Sprintf("%s (%.1f%%)",
					share.ShareFraction, share.SharePercentage))
			case 4: // Тип
				label.SetText(share.OwnershipTypeName)
			case 5: // Дати
				dateStr := share.StartDate.Format("02.01.2006")
				if share.EndDate != nil {
					dateStr += " - " + share.EndDate.Format("02.01.2006")
				} else {
					dateStr += " - теперішній час"
				}
				label.SetText(dateStr)
			case 6: // Дії
				renderActionColumn(label)
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(s.sharesTable, map[int]float32{
		0: 50,  // №
		1: 250, // Власник
		2: 100, // Квартира
		3: 120, // Частка
		4: 150, // Тип
		5: 200, // Дати
		6: 80,  // Дії
	})

	// Клік на комірку
	s.sharesTable.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return // Заголовок
		}
		if id.Row-1 >= len(s.filteredShares) {
			return
		}

		share := s.filteredShares[id.Row-1]

		if id.Col == 6 {
			// Клік на "Дії"
			s.showActionsMenu(share)
		} else {
			// Клік на інші колонки - показати деталі
			s.showShareDetails(share)
		}
	}
}

// Render повертає контейнер з UI екрану.
func (s *OwnershipSharesScreen) Render() fyne.CanvasObject {

	filters := container.NewVBox(s.searchEntry,
		s.filterSelect)

	// Toolbar
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(s.createButton, s.refreshButton),
		nil,
		filters,
	)

	// Головний контейнер
	content := container.NewBorder(
		toolbar,
		s.statsLabel,
		nil,
		nil,
		s.sharesTable,
	)

	return content
}

// loadShares завантажує список часток власності.
func (s *OwnershipSharesScreen) loadShares() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := s.ownershipService.List(ctx, ownership.ListOwnershipSharesInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         1000, // Завантажуємо всіх для клієнтської фільтрації
		Offset:        0,
		OrderBy:       "start_date",
		OrderDesc:     true,
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка завантаження часток: %v", err))
		return
	}

	s.shares = output.Shares
	s.totalCount = output.Total
	s.applyFilters()
}

// applyFilters застосовує фільтри до списку.
func (s *OwnershipSharesScreen) applyFilters() {
	s.filteredShares = make([]*ownership.OwnershipShareDetailsOutput, 0)

	searchQuery := s.searchEntry.Text
	filterType := s.filterSelect.Selected

	s.filteredShares = FilterList(
		s.shares,
		searchQuery,
		filterType,
		func(share *ownership.OwnershipShareDetailsOutput, query string) bool {
			return contains(share.OwnerName, query) || contains(share.ApartmentNumber, query)
		},
		func(share *ownership.OwnershipShareDetailsOutput, filter string) bool {
			switch OwnershipShareFilterType(filter) {
			case OwnershipShareFilterActive:
				return share.IsCurrentlyActive
			case OwnershipShareFilterFull:
				return share.OwnershipType == entity.OwnershipTypeFull
			case OwnershipShareFilterShared:
				return share.OwnershipType == entity.OwnershipTypeShared
			case OwnershipShareFilterRent:
				return share.OwnershipType == entity.OwnershipTypeRent
			case OwnershipShareFilterFinished:
				return !share.IsCurrentlyActive
			}
			return true
		},
	)

	s.sharesTable.Refresh()
	s.updateStats()
}

// updateStats оновлює статистику.
func (s *OwnershipSharesScreen) updateStats() {
	activeCount := 0
	for _, share := range s.filteredShares {
		if share.IsCurrentlyActive {
			activeCount++
		}
	}

	s.statsLabel.SetText(fmt.Sprintf(
		"Показано: %d з %d часток | Активних зараз: %d",
		len(s.filteredShares), s.totalCount, activeCount,
	))
}

// showShareDetails показує детальну інформацію про частку.
func (s *OwnershipSharesScreen) showShareDetails(share *ownership.OwnershipShareDetailsOutput) {
	details := fmt.Sprintf(
		"Власник: %s\n"+
			"Телефон: %s\n"+
			"Email: %s\n\n"+
			"Квартира: №%s, %d поверх\n"+
			"Під'їзд: %s\n\n"+
			"Частка: %s (%.2f%%)\n"+
			"Тип власності: %s\n\n"+
			"Дата початку: %s\n"+
			"Дата завершення: %s\n\n"+
			"Статус: %s\n",
		share.OwnerName,
		ptrToString(share.OwnerPhone, "не вказано"),
		ptrToString(share.OwnerEmail, "не вказано"),
		share.ApartmentNumber,
		share.ApartmentFloor,
		ptrIntToString(share.ApartmentEntrance, "не вказано"),
		share.ShareFraction,
		share.SharePercentage,
		share.OwnershipTypeName,
		share.StartDate.Format("02.01.2006"),
		ptrTimeToString(share.EndDate, "02.01.2006", "теперішній час"),
		currentStatus(share.IsCurrentlyActive),
	)

	if share.DocumentType != nil {
		details += fmt.Sprintf("\nДокумент: %s\n", *share.DocumentType)
		if share.DocumentNumber != nil {
			details += fmt.Sprintf("Номер: %s\n", *share.DocumentNumber)
		}
		if share.DocumentDate != nil {
			details += fmt.Sprintf("Дата: %s\n", share.DocumentDate.Format("02.01.2006"))
		}
	}

	if share.Notes != nil {
		details += fmt.Sprintf("\nПримітки: %s\n", *share.Notes)
	}

	common.ShowInformation(s.window, "Деталі частки власності", details)
}

// showActionsMenu показує меню дій з часткою.
func (s *OwnershipSharesScreen) showActionsMenu(share *ownership.OwnershipShareDetailsOutput) {
	actions := buildStandardActions(
		func() { s.showEditDialog(share) },
		func() { s.confirmDelete(share) },
		func() { s.showShareDetails(share) },
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf("Дії з часткою #%d", share.ID), actions)
}

// showCreateDialog показує діалог створення частки.
func (s *OwnershipSharesScreen) showCreateDialog() {
	formDialog := NewOwnershipFormDialog(s.window, s.ownershipService, s.ownerService, s.apartmentService, s.authManager, nil)
	formDialog.OnSuccess = func() {
		s.loadShares()
	}
	formDialog.Show()
}

// showEditDialog показує діалог редагування частки.
func (s *OwnershipSharesScreen) showEditDialog(share *ownership.OwnershipShareDetailsOutput) {
	formDialog := NewOwnershipFormDialog(s.window, s.ownershipService, s.ownerService, s.apartmentService, s.authManager, share)
	formDialog.OnSuccess = func() {
		s.loadShares()
	}
	formDialog.Show()
}

// confirmDelete підтверджує видалення частки.
func (s *OwnershipSharesScreen) confirmDelete(share *ownership.OwnershipShareDetailsOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf("частку власності #%d", share.ID), func() {
		s.deleteShare(share)
	})
}

// deleteShare видаляє частку.
func (s *OwnershipSharesScreen) deleteShare(share *ownership.OwnershipShareDetailsOutput) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.ownershipService.Delete(ctx, ownership.DeleteOwnershipShareInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ShareID:       share.ID,
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка видалення: %v", err))
		return
	}

	common.ShowSuccess(s.window, "Частку власності успішно видалено")
	s.loadShares()
}

func currentStatus(isActive bool) string {
	if isActive {
		return "✅ Активна зараз"
	}
	return "❌ Завершена"
}
