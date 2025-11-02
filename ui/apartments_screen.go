package ui

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/service"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ApartmentsScreen представляє екран управління квартирами.
type ApartmentsScreen struct {
	window           fyne.Window
	apartmentService *service.ApartmentService
	currentUser      *domain.User

	// UI елементи
	apartmentsList *widget.List
	searchEntry    *widget.Entry
	apartments     []*domain.Apartment

	// Callbacks
	onRefresh func()
}

// NewApartmentsScreen створює новий екран управління квартирами.
func NewApartmentsScreen(
	window fyne.Window,
	apartmentService *service.ApartmentService,
	currentUser *domain.User,
) *ApartmentsScreen {
	// КРИТИЧНО: Перевірка вхідних параметрів
	if apartmentService == nil {
		panic("КРИТИЧНА ПОМИЛКА: ApartmentService не може бути nil!")
	}
	if window == nil {
		panic("КРИТИЧНА ПОМИЛКА: Window не може бути nil!")
	}
	if currentUser == nil {
		panic("КРИТИЧНА ПОМИЛКА: CurrentUser не може бути nil!")
	}

	screen := &ApartmentsScreen{
		window:           window,
		apartmentService: apartmentService,
		currentUser:      currentUser,
		apartments:       []*domain.Apartment{},
	}

	screen.initUI()
	screen.loadApartments()

	return screen
}

// initUI ініціалізує UI елементи.
func (s *ApartmentsScreen) initUI() {
	// Поле пошуку
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("Пошук за номером або власником...")
	s.searchEntry.OnChanged = func(query string) {
		s.filterApartments(query)
	}

	// Список квартир
	s.apartmentsList = widget.NewList(
		func() int {
			return len(s.apartments)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Template"),
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(s.apartments) {
				return
			}

			apt := s.apartments[id]
			box := obj.(*fyne.Container)

			// Номер квартири та поверх
			numberLabel := box.Objects[0].(*widget.Label)
			numberLabel.SetText(fmt.Sprintf("Кв. %s (поверх %d)",
				apt.ApartmentNumber, apt.Floor))
			numberLabel.TextStyle = fyne.TextStyle{Bold: true}

			// Власник та площа
			detailsLabel := box.Objects[1].(*widget.Label)
			detailsLabel.SetText(fmt.Sprintf("%s • %.1f м² • %d кімн.",
				apt.OwnerName, apt.Area, apt.Rooms))
		},
	)

	s.apartmentsList.OnSelected = func(id widget.ListItemID) {
		if id >= len(s.apartments) {
			return
		}
		s.showApartmentDetails(s.apartments[id])
		s.apartmentsList.UnselectAll()
	}
}

// Render повертає UI компонент екрану.
func (s *ApartmentsScreen) Render() fyne.CanvasObject {
	// Тулбар з кнопками
	toolbar := s.createToolbar()

	// Поле пошуку
	searchContainer := container.NewBorder(
		nil, nil,
		widget.NewIcon(theme.SearchIcon()),
		nil,
		s.searchEntry,
	)

	// Статистика
	stats := s.createStatsCard()

	// Основний контент
	content := container.NewBorder(
		container.NewVBox(
			toolbar,
			searchContainer,
			stats,
		),
		nil,
		nil,
		nil,
		s.apartmentsList,
	)

	return container.NewPadded(content)
}

// createToolbar створює панель інструментів.
func (s *ApartmentsScreen) createToolbar() *fyne.Container {
	addBtn := widget.NewButtonWithIcon("Додати квартиру", theme.ContentAddIcon(), func() {
		s.showAddDialog()
	})
	addBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadApartments()
	})

	return container.NewHBox(
		addBtn,
		refreshBtn,
	)
}

// createStatsCard створює картку зі статистикою.
func (s *ApartmentsScreen) createStatsCard() *widget.Card {
	stats, err := s.apartmentService.GetStatistics(s.currentUser.ID)
	if err != nil {
		return widget.NewCard("", "", widget.NewLabel("Помилка завантаження статистики"))
	}

	statsText := fmt.Sprintf(
		"Всього квартир: %d • Загальна площа: %.1f м² • Мешканців: %d",
		stats.TotalApartments,
		stats.TotalArea,
		stats.TotalResidents,
	)

	return widget.NewCard("", "", widget.NewLabel(statsText))
}

// loadApartments завантажує список квартир з БД.
func (s *ApartmentsScreen) loadApartments() {
	// Перевірка на nil сервіс
	if s.apartmentService == nil {
		dialog.ShowError(fmt.Errorf("помилка: сервіс квартир не ініціалізовано"), s.window)
		return
	}

	apartments, err := s.apartmentService.GetAllApartments(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження квартир: %w", err), s.window)
		return
	}

	s.apartments = apartments
	s.apartmentsList.Refresh()
}

// filterApartments фільтрує квартири за пошуковим запитом.
func (s *ApartmentsScreen) filterApartments(query string) {
	if query == "" {
		s.loadApartments()
		return
	}

	filter := &domain.ApartmentFilter{
		SearchQuery: query,
		IsActive:    boolPtr(true),
	}

	apartments, err := s.apartmentService.SearchApartments(s.currentUser.ID, filter)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.apartments = apartments
	s.apartmentsList.Refresh()
}

// showApartmentDetails показує детальну інформацію про квартиру.
func (s *ApartmentsScreen) showApartmentDetails(apt *domain.Apartment) {
	details := fmt.Sprintf(
		"Квартира №%s\n\n"+
			"Поверх: %d\n"+
			"Площа: %.1f м²\n"+
			"Кімнат: %d\n"+
			"Мешканців: %d\n\n"+
			"Власник: %s\n"+
			"Телефон: %s\n",
		apt.ApartmentNumber,
		apt.Floor,
		apt.Area,
		apt.Rooms,
		apt.ResidentsCount,
		apt.OwnerName,
		apt.OwnerPhone,
	)

	if apt.OwnerEmail != "" {
		details += fmt.Sprintf("Email: %s\n", apt.OwnerEmail)
	}

	if apt.Notes != "" {
		details += fmt.Sprintf("\nПримітки: %s\n", apt.Notes)
	}

	detailsLabel := widget.NewLabel(details)

	editBtn := widget.NewButton("Редагувати", func() {
		s.showEditDialog(apt)
	})

	deleteBtn := widget.NewButton("Видалити", func() {
		s.confirmDelete(apt)
	})
	deleteBtn.Importance = widget.DangerImportance

	content := container.NewVBox(
		detailsLabel,
		container.NewHBox(editBtn, deleteBtn),
	)

	d := dialog.NewCustom(
		"Деталі квартири",
		"Закрити",
		content,
		s.window,
	)
	d.Resize(fyne.NewSize(400, 400))
	d.Show()
}

// showAddDialog показує діалог додавання нової квартири.
func (s *ApartmentsScreen) showAddDialog() {
	apt := &domain.Apartment{
		ResidentsCount: 1,
	}

	form := s.createApartmentForm(apt, true)

	dialog.NewCustomConfirm(
		"Додати квартиру",
		"Додати",
		"Скасувати",
		form,
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := s.apartmentService.CreateApartment(s.currentUser.ID, apt); err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Квартиру додано!", s.window)
			s.loadApartments()
		},
		s.window,
	).Show()
}

// showEditDialog показує діалог редагування квартири.
func (s *ApartmentsScreen) showEditDialog(apt *domain.Apartment) {
	// Створюємо копію для редагування
	editApt := *apt

	form := s.createApartmentForm(&editApt, false)

	dialog.NewCustomConfirm(
		"Редагувати квартиру",
		"Зберегти",
		"Скасувати",
		form,
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := s.apartmentService.UpdateApartment(s.currentUser.ID, &editApt); err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Зміни збережено!", s.window)
			s.loadApartments()
		},
		s.window,
	).Show()
}

// createApartmentForm створює форму для введення даних квартири.
func (s *ApartmentsScreen) createApartmentForm(apt *domain.Apartment, isNew bool) *fyne.Container {
	numberEntry := widget.NewEntry()
	numberEntry.SetPlaceHolder("Наприклад: 42 або 12А")
	if !isNew {
		numberEntry.SetText(apt.ApartmentNumber)
	}
	numberEntry.OnChanged = func(val string) {
		apt.ApartmentNumber = val
	}

	floorEntry := widget.NewEntry()
	floorEntry.SetPlaceHolder("1")
	if !isNew {
		floorEntry.SetText(strconv.Itoa(apt.Floor))
	}
	floorEntry.OnChanged = func(val string) {
		if floor, err := strconv.Atoi(val); err == nil {
			apt.Floor = floor
		}
	}

	areaEntry := widget.NewEntry()
	areaEntry.SetPlaceHolder("45.5")
	if !isNew {
		areaEntry.SetText(fmt.Sprintf("%.1f", apt.Area))
	}
	areaEntry.OnChanged = func(val string) {
		if area, err := strconv.ParseFloat(val, 64); err == nil {
			apt.Area = area
		}
	}

	roomsEntry := widget.NewEntry()
	roomsEntry.SetPlaceHolder("2")
	if !isNew {
		roomsEntry.SetText(strconv.Itoa(apt.Rooms))
	}
	roomsEntry.OnChanged = func(val string) {
		if rooms, err := strconv.Atoi(val); err == nil {
			apt.Rooms = rooms
		}
	}

	ownerEntry := widget.NewEntry()
	ownerEntry.SetPlaceHolder("Прізвище Ім'я По батькові")
	if !isNew {
		ownerEntry.SetText(apt.OwnerName)
	}
	ownerEntry.OnChanged = func(val string) {
		apt.OwnerName = val
	}

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("+380501234567")
	if !isNew {
		phoneEntry.SetText(apt.OwnerPhone)
	}
	phoneEntry.OnChanged = func(val string) {
		apt.OwnerPhone = val
	}

	residentsEntry := widget.NewEntry()
	residentsEntry.SetPlaceHolder("2")
	if !isNew {
		residentsEntry.SetText(strconv.Itoa(apt.ResidentsCount))
	}
	residentsEntry.OnChanged = func(val string) {
		if count, err := strconv.Atoi(val); err == nil {
			apt.ResidentsCount = count
		}
	}

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Додаткова інформація...")
	if !isNew {
		notesEntry.SetText(apt.Notes)
	}
	notesEntry.OnChanged = func(val string) {
		apt.Notes = val
	}

	form := container.NewVBox(
		widget.NewLabel("Номер квартири:"),
		numberEntry,
		widget.NewLabel("Поверх:"),
		floorEntry,
		widget.NewLabel("Площа (м²):"),
		areaEntry,
		widget.NewLabel("Кількість кімнат:"),
		roomsEntry,
		widget.NewLabel("ПІБ власника:"),
		ownerEntry,
		widget.NewLabel("Телефон:"),
		phoneEntry,
		widget.NewLabel("Кількість мешканців:"),
		residentsEntry,
		widget.NewLabel("Примітки:"),
		notesEntry,
	)

	return form
}

// confirmDelete показує діалог підтвердження видалення.
func (s *ApartmentsScreen) confirmDelete(apt *domain.Apartment) {
	dialog.ShowConfirm(
		"Видалення квартири",
		fmt.Sprintf("Ви впевнені, що хочете видалити квартиру №%s?", apt.ApartmentNumber),
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := s.apartmentService.DeleteApartment(s.currentUser.ID, apt.ID); err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Квартиру видалено!", s.window)
			s.loadApartments()
		},
		s.window,
	)
}

// boolPtr - допоміжна функція для створення вказівника на bool.
func boolPtr(b bool) *bool {
	return &b
}
