// presentation/fyne/screens/ownership_shares_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/ownership"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/auth"
)

// OwnershipSharesScreen представляє екран управління частками власності.
type OwnershipSharesScreen struct {
	window           fyne.Window
	apartmentService *service.ApartmentService
	ownerService     *service.OwnerService
	ownershipService *service.OwnershipService
	authManager      *auth.AuthManager

	// UI елементи
	searchEntry     *widget.Entry
	filterSelect    *widget.Select
	apartmentFilter *widget.Entry
	ownerFilter     *widget.Entry
	sharesTable     *widget.Table
	createButton    *widget.Button
	refreshButton   *widget.Button
	statsLabel      *widget.Label

	// Дані
	shares         []*ownership.OwnershipShareDetailsOutput
	filteredShares []*ownership.OwnershipShareDetailsOutput
	currentPage    int
	pageSize       int
	totalCount     int64
}

// NewOwnershipSharesScreen створює новий екран часток власності.
func NewOwnershipSharesScreen(
	window fyne.Window,
	apaertmentService *service.ApartmentService,
	ownerService *service.OwnerService,
	ownershipService *service.OwnershipService,
	authManager *auth.AuthManager,
) *OwnershipSharesScreen {
	screen := &OwnershipSharesScreen{
		window:           window,
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
	s.filterSelect = widget.NewSelect([]string{
		"Всі частки",
		"Тільки активні зараз",
		"Повна власність",
		"Часткова власність",
		"Оренда",
		"Завершені",
	}, func(value string) {
		s.applyFilters()
	})

	// Фільтр за квартирою
	s.apartmentFilter = widget.NewEntry()
	s.apartmentFilter.SetPlaceHolder("Квартира №")
	s.apartmentFilter.OnChanged = func(query string) {
		s.applyFilters()
	}

	// Фільтр за власником
	s.ownerFilter = widget.NewEntry()
	s.ownerFilter.SetPlaceHolder("ПІБ власника")
	s.ownerFilter.OnChanged = func(query string) {
		s.applyFilters()
	}

	// Кнопка створення
	s.createButton = widget.NewButtonWithIcon("Додати частку", theme.ContentAddIcon(), func() {
		s.showCreateDialog()
	})
	s.createButton.Importance = widget.HighImportance

	// Кнопка оновлення
	s.refreshButton = widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadShares()
	})

	// Статистика
	s.statsLabel = widget.NewLabel("")

	// Таблиця
	s.buildTable()

	// 2. Встановлюємо дефолтні значення, які можуть тригерити колбеки
	// Робимо це в самому кінці, коли всі поля (searchEntry, apartmentFilter, ownerFilter) вже існують
	s.filterSelect.SetSelected("Тільки активні зараз")
}

// buildTable створює таблицю з частками власності.
func (s *OwnershipSharesScreen) buildTable() {
	s.sharesTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredShares) + 1, 7 // +1 для заголовків, 7 колонок
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				headers := []string{"№", "Власник", "Квартира", "Частка", "Тип", "Дати", "Дії"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
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
				if share.IsCurrentlyActive {
					label.SetText("✅")
				} else {
					label.SetText("❌")
				}
			}
		},
	)

	// Ширина колонок
	s.sharesTable.SetColumnWidth(0, 50)  // №
	s.sharesTable.SetColumnWidth(1, 250) // Власник
	s.sharesTable.SetColumnWidth(2, 100) // Квартира
	s.sharesTable.SetColumnWidth(3, 120) // Частка
	s.sharesTable.SetColumnWidth(4, 150) // Тип
	s.sharesTable.SetColumnWidth(5, 200) // Дати
	s.sharesTable.SetColumnWidth(6, 80)  // Дії

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
	// Панель фільтрів
	filtersRow1 := container.NewGridWithColumns(2,
		s.searchEntry,
		s.filterSelect,
	)

	filtersRow2 := container.NewGridWithColumns(2,
		s.apartmentFilter,
		s.ownerFilter,
	)

	filters := container.NewVBox(filtersRow1, filtersRow2)

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
		dialog.ShowError(fmt.Errorf("Помилка завантаження часток: %v", err), s.window)
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
	apartmentQuery := s.apartmentFilter.Text
	ownerQuery := s.ownerFilter.Text

	for _, share := range s.shares {
		// Фільтр за типом
		switch filterType {
		case "Тільки активні зараз":
			if !share.IsCurrentlyActive {
				continue
			}
		case "Повна власність":
			if share.OwnershipType != entity.OwnershipTypeFull {
				continue
			}
		case "Часткова власність":
			if share.OwnershipType != entity.OwnershipTypeShared {
				continue
			}
		case "Оренда":
			if share.OwnershipType != entity.OwnershipTypeRent {
				continue
			}
		case "Завершені":
			if share.IsCurrentlyActive {
				continue
			}
		}

		// Загальний пошук
		if searchQuery != "" {
			match := false
			if contains(share.OwnerName, searchQuery) {
				match = true
			}
			if contains(share.ApartmentNumber, searchQuery) {
				match = true
			}
			if !match {
				continue
			}
		}

		// Пошук за квартирою
		if apartmentQuery != "" {
			if !contains(share.ApartmentNumber, apartmentQuery) {
				continue
			}
		}

		// Пошук за власником
		if ownerQuery != "" {
			if !contains(share.OwnerName, ownerQuery) {
				continue
			}
		}

		s.filteredShares = append(s.filteredShares, share)
	}

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
		intPtrToString(share.ApartmentEntrance, "не вказано"),
		share.ShareFraction,
		share.SharePercentage,
		share.OwnershipTypeName,
		share.StartDate.Format("02.01.2006"),
		timeOrDefault(share.EndDate, "теперішній час"),
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

	dialog.ShowInformation("Деталі частки власності", details, s.window)
}

// showActionsMenu показує меню дій з часткою.
func (s *OwnershipSharesScreen) showActionsMenu(share *ownership.OwnershipShareDetailsOutput) {
	editButton := widget.NewButton("✏️ Редагувати", func() {
		s.showEditDialog(share)
	})

	deleteButton := widget.NewButton("🗑️ Видалити", func() {
		s.confirmDelete(share)
	})
	deleteButton.Importance = widget.DangerImportance

	detailsButton := widget.NewButton("ℹ️ Деталі", func() {
		s.showShareDetails(share)
	})

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("%s - Кв. %s (%s)",
			share.OwnerName, share.ApartmentNumber, share.ShareFraction)),
		layout.NewSpacer(),
		editButton,
		deleteButton,
		detailsButton,
	)

	dialog.ShowCustom("Дії", "Закрити", content, s.window)
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
	dialog.ShowConfirm(
		"Підтвердження видалення",
		fmt.Sprintf(
			"Ви впевнені, що хочете видалити частку власності?\n\n"+
				"Власник: %s\n"+
				"Квартира: №%s\n"+
				"Частка: %s (%.1f%%)",
			share.OwnerName,
			share.ApartmentNumber,
			share.ShareFraction,
			share.SharePercentage,
		),
		func(confirmed bool) {
			if confirmed {
				s.deleteShare(share)
			}
		},
		s.window,
	)
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
		dialog.ShowError(fmt.Errorf("Помилка видалення: %v", err), s.window)
		return
	}

	dialog.ShowInformation("Успіх", "Частку власності успішно видалено", s.window)
	s.loadShares()
}

// Helper функції
func timeOrDefault(t *time.Time, defaultVal string) string {
	if t != nil {
		return t.Format("02.01.2006")
	}
	return defaultVal
}

func currentStatus(isActive bool) string {
	if isActive {
		return "✅ Активна зараз"
	}
	return "❌ Завершена"
}

func intPtrToString(ptr *int, defaultVal string) string {
	if ptr != nil {
		return fmt.Sprintf("%d", *ptr)
	}
	return defaultVal
}
