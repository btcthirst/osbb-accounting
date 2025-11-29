package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/dialogs"
)

type ApartmentsScreen struct {
	window           fyne.Window
	apartmentService service.ApartmentServiceInterface
	authManager      *auth.AuthManager

	// UI елементи
	searchEntry     *widget.Entry
	filterSelect    *widget.Select
	apartmentsTable *widget.Table
	createButton    *widget.Button
	refreshButton   *widget.Button
	statsLabel      *widget.Label

	//Дані
	apartments         []*apartment.ApartmentOutput
	filteredApartments []*apartment.ApartmentOutput
	currentPage        int
	pageSize           int
	totalCount         int64
}

type ApartmentFilterType string

const (
	ApartmentFilterAll       ApartmentFilterType = "Всі квартири"
	ApartmentFilterEntrance1 ApartmentFilterType = "Перший підїзд"
	ApartmentFilterEntrance2 ApartmentFilterType = "Другий підїзд"
	ApartmentFilterEntrance3 ApartmentFilterType = "Третій підїзд"
	ApartmentFilterEntrance4 ApartmentFilterType = "Четвертий підїзд"
	ApartmentFilterEntrance5 ApartmentFilterType = "П'ятий підїзд"
	ApartmentFilterEntrance6 ApartmentFilterType = "Шостий підїзд"
	ApartmentFilterActive    ApartmentFilterType = "Тільки активні"
	ApartmentFilterInactive  ApartmentFilterType = "Неактивні"
)

func NewApartmentsScreen(
	window fyne.Window,
	apartmentService service.ApartmentServiceInterface,
	authManager *auth.AuthManager,
) *ApartmentsScreen {
	screen := &ApartmentsScreen{
		window:           window,
		apartmentService: apartmentService,
		authManager:      authManager,
		pageSize:         50,
		currentPage:      0,
	}

	screen.buildUI()
	screen.loadApartments()

	return screen
}

// buildUI створює інтерфейс екрану.
func (s *ApartmentsScreen) buildUI() {
	// 1. Спочатку ініціалізуємо всі віджети // Filters
	s.searchEntry = newSearchEntry("Пошук (ПІБ, Номер, Площа)...", func(_ string) { s.applyFilters() })
	s.filterSelect = newFilterSelect([]string{
		string(ApartmentFilterAll),
		string(ApartmentFilterEntrance1),
		string(ApartmentFilterEntrance2),
		string(ApartmentFilterEntrance3),
		string(ApartmentFilterEntrance4),
		string(ApartmentFilterEntrance5),
		string(ApartmentFilterEntrance6),
		string(ApartmentFilterActive),
		string(ApartmentFilterInactive),
	},
		func(_ string) {
			s.applyFilters()
		},
	)

	s.createButton = newCreateButton("Додати квартиру", func() {
		s.showApartmentDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadApartments()
	})

	s.statsLabel = widget.NewLabel("")

	// 2. Створюємо таблицю
	s.buildTable()

	// 3. Встановлюємо дефолтні значення (викличе applyFilters)
	s.filterSelect.SetSelected(string(ApartmentFilterAll))
}

// buildTable створює таблицю з квартирами.
func (s *ApartmentsScreen) buildTable() {
	s.apartmentsTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredApartments) + 1, 6 // +1 для заголовків, 6 колонок
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				renderTableHeader(label,
					[]string{"№", "ПІБ", "Номер", "Жила площа", "Підїзд", "Дії"}, id.Col)
				return
			}

			// Дані
			dataIndex := id.Row - 1
			if dataIndex >= len(s.filteredApartments) {
				label.SetText("")
				return
			}

			apartment := s.filteredApartments[dataIndex]

			switch id.Col {
			case 0: // №
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1: // ПІБ (DisplayName)
				label.SetText(ptrToString(apartment.CadastralNumber, "-"))
				if !apartment.IsActive {
					label.TextStyle = fyne.TextStyle{Italic: true}
				} else {
					label.TextStyle = fyne.TextStyle{}
				}
			case 2: // Номер квартири
				label.SetText(apartment.ApartmentNumber)
			case 3: // Жила площа
				if apartment.AreaLiving != nil {
					label.SetText(fmt.Sprintf("%.1f", *apartment.AreaLiving))
				} else {
					label.SetText("—")
				}
			case 4: // Підїзд
				if apartment.Entrance != nil {
					label.SetText(fmt.Sprintf("%d", *apartment.Entrance))
				} else {
					label.SetText("—")
				}
			case 5: // Дії
				renderActionColumn(label)
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(s.apartmentsTable, map[int]float32{
		0: 50,  // №
		1: 250, // ПІБ
		2: 100, // Номер
		3: 100, // Жила площа
		4: 80,  // Підїзд
		5: 50,  // Дії
	})

	// Клік на комірку
	s.apartmentsTable.OnSelected = func(id widget.TableCellID) {
		defer s.apartmentsTable.Unselect(id)

		if id.Row == 0 {
			return // Заголовок
		}
		dataIndex := id.Row - 1
		if dataIndex >= len(s.filteredApartments) {
			return
		}

		apartment := s.filteredApartments[dataIndex]

		if id.Col == 5 {
			// Клік на "Дії"
			s.showActionsMenu(apartment)
		} else {
			// Клік на інші колонки - показати деталі
			s.showApartmentDetails(apartment)
		}
	}
}

// Render повертає контейнер з UI екрану.
func (s *ApartmentsScreen) Render() fyne.CanvasObject {
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
		s.apartmentsTable,
	)

	return content
}

// getStatsText повертає текст статистики.
func (s *ApartmentsScreen) getStatsText() string {
	return fmt.Sprintf("Показано: %d з %d квартир", len(s.filteredApartments), s.totalCount)
}

// updateStats оновлює статистику.
func (s *ApartmentsScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}

// loadApartments завантажує список квартир.
func (s *ApartmentsScreen) loadApartments() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Перевірка на nil перед викликом (хоча з фіксом конструктора це не має статися, але для безпеки)
	var currentUserID int64
	if s.authManager != nil {
		currentUserID = s.authManager.GetCurrentUserID()
	}

	output, err := s.apartmentService.List(ctx, apartment.ListApartmentsInput{
		CurrentUserID: currentUserID,
		Limit:         1000, // Завантажуємо всіх для клієнтської фільтрації
		Offset:        0,
		OrderBy:       "number",
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка завантаження квартир: %v", err))
		return
	}

	s.apartments = output.Apartments
	s.totalCount = output.Total
	s.applyFilters()
}

// applyFilters застосовує фільтри до списку.
func (s *ApartmentsScreen) applyFilters() {
	s.filteredApartments = make([]*apartment.ApartmentOutput, 0)

	// Безпечне отримання значень UI
	searchQuery := ""
	if s.searchEntry != nil {
		searchQuery = s.searchEntry.Text
	}
	filterType := ""
	if s.filterSelect != nil {
		filterType = s.filterSelect.Selected
	}

	s.filteredApartments = FilterList(
		s.apartments,
		searchQuery,
		filterType,
		func(a *apartment.ApartmentOutput, query string) bool {
			return contains(a.DisplayName, query) ||
				contains(a.ApartmentNumber, query) ||
				(a.AreaLiving != nil && contains(fmt.Sprintf("%v", *a.AreaLiving), query))
		},
		func(a *apartment.ApartmentOutput, filter string) bool {
			switch ApartmentFilterType(filter) {
			case ApartmentFilterEntrance1:
				return a.Entrance != nil && *a.Entrance == 1
			case ApartmentFilterEntrance2:
				return a.Entrance != nil && *a.Entrance == 2
			case ApartmentFilterEntrance3:
				return a.Entrance != nil && *a.Entrance == 3
			case ApartmentFilterEntrance4:
				return a.Entrance != nil && *a.Entrance == 4
			case ApartmentFilterEntrance5:
				return a.Entrance != nil && *a.Entrance == 5
			case ApartmentFilterEntrance6:
				return a.Entrance != nil && *a.Entrance == 6
			case ApartmentFilterActive:
				return a.IsActive
			case ApartmentFilterInactive:
				return !a.IsActive
			}
			return true
		},
	)

	s.apartmentsTable.Refresh()
	s.updateStats()
}

// showApartmentDetails показує детальну інформацію про квартиру.
func (s *ApartmentsScreen) showApartmentDetails(a *apartment.ApartmentOutput) {
	details := fmt.Sprintf(
		"Квартира: №%s\n"+
			"Поверх: %d\n"+
			"Під'їзд: %s\n\n"+
			"Площа: %.2f м²\n"+
			"Жила площа: %s\n"+
			"Кількість кімнат: %s\n"+
			"Кадастровий номер: %s\n"+
			"Статус: %s\n",
		a.ApartmentNumber,
		a.Floor,
		ptrIntToString(a.Entrance, "не вказано"),
		a.AreaTotal,
		ptrFloatToString(a.AreaLiving, "не вказано"),
		ptrIntToString(a.RoomsCount, "не вказано"),
		ptrToString(a.CadastralNumber, "не вказано"),
		activeStatus(a.IsActive),
	)

	if a.Notes != nil {
		details += fmt.Sprintf("\nПримітки: %s\n", *a.Notes)
	}

	common.ShowInformation(s.window, "Інформація про квартиру", details)
}

// showActionsMenu показує меню дій з власником.
func (s *ApartmentsScreen) showActionsMenu(a *apartment.ApartmentOutput) {
	actions := buildStandardActions(
		func() { s.showApartmentDialog(a) },
		func() { s.confirmDelete(a) },
		func() { s.showApartmentDetails(a) },
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf("Дії з квартирою № %s", a.ApartmentNumber), actions)
}

// showApartmentDialog показує діалог створення/редагування квартири.
func (s *ApartmentsScreen) showApartmentDialog(existing *apartment.ApartmentOutput) {
	d := dialogs.NewApartmentFormDialog(
		s.window,
		s.apartmentService,
		s.authManager,
		existing,
	)
	d.OnSuccess = func() {
		s.loadApartments()
	}
	d.Show()
}

// confirmDelete підтверджує видалення.
func (s *ApartmentsScreen) confirmDelete(a *apartment.ApartmentOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf("квартиру №%s", a.ApartmentNumber), func() {
		s.deleteApartment(a)
	})
}

// deleteApartment видаляє квартиру.
func (s *ApartmentsScreen) deleteApartment(a *apartment.ApartmentOutput) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.apartmentService.Delete(ctx, apartment.DeleteApartmentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ApartmentID:   a.ID,
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка видалення: %v", err))
		return
	}

	common.ShowSuccess(s.window, fmt.Sprintf("Квартиру №%s успішно видалено", a.ApartmentNumber))
	s.loadApartments()
}
