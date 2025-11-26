package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/common"
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
	updateStats        func()
	currentPage        int
	pageSize           int
	totalCount         int64
}

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
	// 1. Спочатку ініціалізуємо всі віджети
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("🔍 Пошук за власником, номером...")
	s.searchEntry.OnChanged = func(query string) {
		s.applyFilters()
	}

	s.filterSelect = widget.NewSelect([]string{
		"Всі квартири",
		"Перший підїзд",
		"Другий підїзд",
		"Третій підїзд",
		"Четвертий підїзд",
		"П'ятий підїзд",
		"Шостий підїзд",
		"Тільки активні",
		"Неактивні",
	}, func(value string) {
		s.applyFilters()
	})

	s.createButton = widget.NewButtonWithIcon("Додати квартиру", theme.ContentAddIcon(), func() {
		s.showCreateDialog()
	})
	s.createButton.Importance = widget.HighImportance

	s.refreshButton = widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadApartments()
	})

	s.statsLabel = widget.NewLabel("")

	// 2. Створюємо таблицю
	s.buildTable()

	// 3. Ініціалізуємо функцію оновлення статистики
	s.updateStats = func() {
		s.statsLabel.SetText(s.getStatsText())
	}

	// 4. Встановлюємо дефолтні значення (викличе applyFilters)
	s.filterSelect.SetSelected("Всі квартири")
}

// buildTable створює таблицю з квартирами.
func (s *ApartmentsScreen) buildTable() {
	s.apartmentsTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredApartments) + 1, 6 // +1 для заголовків, 6 колонок
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("Template")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				headers := []string{"№", "ПІБ", "Номер", "Жила площа", "Підїзд", "Дії"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
					label.TextStyle = fyne.TextStyle{Bold: true}
				}
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
				label.SetText("⚙️")
			}
		},
	)

	// Ширина колонок
	s.apartmentsTable.SetColumnWidth(0, 50)  // №
	s.apartmentsTable.SetColumnWidth(1, 250) // ПІБ
	s.apartmentsTable.SetColumnWidth(2, 100) // Номер
	s.apartmentsTable.SetColumnWidth(3, 100) // Жила площа
	s.apartmentsTable.SetColumnWidth(4, 80)  // Підїзд
	s.apartmentsTable.SetColumnWidth(5, 50)  // Дії

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

	for _, a := range s.apartments {
		// Пошук
		if searchQuery != "" {
			match := false
			// Пошук в ПІБ (DisplayName містить номер)
			if contains(a.DisplayName, searchQuery) {
				match = true
			}
			// Пошук в номері
			if contains(a.ApartmentNumber, searchQuery) {
				match = true
			}
			// Пошук в площі
			if a.AreaLiving != nil && contains(fmt.Sprintf("%v", *a.AreaLiving), searchQuery) {
				match = true
			}

			if !match {
				continue
			}
		}

		// Фільтр
		switch filterType {
		case "Перший підїзд":
			if a.Entrance == nil || *a.Entrance != 1 {
				continue
			}
		case "Другий підїзд":
			if a.Entrance == nil || *a.Entrance != 2 {
				continue
			}
		case "Третій підїзд":
			if a.Entrance == nil || *a.Entrance != 3 {
				continue
			}
		case "Четвертий підїзд":
			if a.Entrance == nil || *a.Entrance != 4 {
				continue
			}
		case "П'ятий підїзд":
			if a.Entrance == nil || *a.Entrance != 5 {
				continue
			}
		case "Шостий підїзд":
			if a.Entrance == nil || *a.Entrance != 6 {
				continue
			}
		case "Тільки активні":
			if !a.IsActive {
				continue
			}
		case "Неактивні":
			if a.IsActive {
				continue
			}
		}

		s.filteredApartments = append(s.filteredApartments, a)
	}

	s.apartmentsTable.Refresh()
	if s.updateStats != nil {
		s.updateStats()
	}
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
	actions := []common.Action{
		{
			Label: "✏️ Редагувати",
			OnTap: func() { s.showEditDialog(a) },
		},
		{
			Label:      "🗑️ Видалити",
			OnTap:      func() { s.confirmDelete(a) },
			Importance: widget.DangerImportance,
		},
		{
			Label: "ℹ️ Деталі",
			OnTap: func() { s.showApartmentDetails(a) },
		},
	}

	common.ShowActionsMenu(s.window, fmt.Sprintf("Дії з квартирою № %s", a.ApartmentNumber), actions)
}

// showCreateDialog показує діалог створення.
func (s *ApartmentsScreen) showCreateDialog() {
	dialog := NewApartmentFormDialog(s.window, s.apartmentService, s.authManager, nil)
	dialog.OnSuccess = func() {
		s.loadApartments()
	}
	dialog.Show()
}

// showEditDialog показує діалог редагування.
func (s *ApartmentsScreen) showEditDialog(a *apartment.ApartmentOutput) {
	dialog := NewApartmentFormDialog(s.window, s.apartmentService, s.authManager, a)
	dialog.OnSuccess = func() {
		s.loadApartments()
	}
	dialog.Show()
}

// confirmDelete підтверджує видалення.
func (s *ApartmentsScreen) confirmDelete(a *apartment.ApartmentOutput) {
	common.ShowDeleteConfirmation(
		s.window,
		"Підтвердження видалення",
		fmt.Sprintf("Ви впевнені, що хочете видалити квартиру №%s?", a.ApartmentNumber),
		func() {
			s.deleteApartment(a)
		},
	)
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
