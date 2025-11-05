package ui

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// OwnersScreen представляє екран управління власниками.
type OwnersScreen struct {
	window       fyne.Window
	ownerService *service.OwnerService
	currentUser  *domain.User

	// UI елементи
	ownersList  *widget.List
	searchEntry *widget.Entry
	owners      []*domain.Owner
}

// NewOwnersScreen створює новий екран власників.
func NewOwnersScreen(
	window fyne.Window,
	ownerService *service.OwnerService,
	currentUser *domain.User,
) *OwnersScreen {
	screen := &OwnersScreen{
		window:       window,
		ownerService: ownerService,
		currentUser:  currentUser,
		owners:       []*domain.Owner{},
	}

	screen.initUI()
	screen.loadOwners()

	return screen
}

// initUI ініціалізує UI елементи.
func (s *OwnersScreen) initUI() {
	// Поле пошуку
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("Пошук за ПІБ або ІПН...")
	s.searchEntry.OnChanged = func(query string) {
		s.searchOwners(query)
	}

	// Список власників
	s.ownersList = widget.NewList(
		func() int {
			return len(s.owners)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				container.NewVBox(
					widget.NewLabel("Template"),
					widget.NewLabel("Template"),
				),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(s.owners) {
				return
			}

			owner := s.owners[id]
			box := obj.(*fyne.Container)

			icon := box.Objects[0].(*widget.Icon)
			icon.SetResource(theme.AccountIcon())

			infoBox := box.Objects[1].(*fyne.Container)
			nameLabel := infoBox.Objects[0].(*widget.Label)
			detailsLabel := infoBox.Objects[1].(*widget.Label)

			nameLabel.SetText(owner.FullName)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			detailsLabel.SetText(fmt.Sprintf("ІПН: %s • %s", owner.TaxID, owner.Phone))
		},
	)

	s.ownersList.OnSelected = func(id widget.ListItemID) {
		if id >= len(s.owners) {
			return
		}
		s.showOwnerDetails(s.owners[id])
		s.ownersList.UnselectAll()
	}
}

// Render повертає UI компонент екрану.
func (s *OwnersScreen) Render() fyne.CanvasObject {
	// Тулбар
	toolbar := s.createToolbar()

	// Пошук
	searchContainer := container.NewBorder(
		nil, nil,
		widget.NewIcon(theme.SearchIcon()),
		nil,
		s.searchEntry,
	)

	// Статистика
	stats := s.createStatsCard()

	// Контент
	content := container.NewBorder(
		container.NewVBox(
			toolbar,
			searchContainer,
			stats,
		),
		nil,
		nil,
		nil,
		s.ownersList,
	)

	return container.NewPadded(content)
}

// createToolbar створює панель інструментів.
func (s *OwnersScreen) createToolbar() *fyne.Container {
	addBtn := widget.NewButtonWithIcon("Додати власника", theme.ContentAddIcon(), func() {
		s.showAddOwnerDialog()
	})
	addBtn.Importance = widget.HighImportance

	// Перевірка прав
	if !s.currentUser.IsAdmin() && !s.currentUser.CanManageFinances() {
		addBtn.Disable()
	}

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		s.loadOwners()
	})

	debtorsBtn := widget.NewButton("Боржники", func() {
		s.showDebtors()
	})

	return container.NewHBox(
		addBtn,
		debtorsBtn,
		widget.NewSeparator(),
		refreshBtn,
	)
}

// createStatsCard створює картку зі статистикою.
func (s *OwnersScreen) createStatsCard() *widget.Card {
	count, err := s.ownerService.CountOwners(s.currentUser.ID)
	if err != nil {
		return widget.NewCard("", "", widget.NewLabel("Помилка завантаження"))
	}

	statsText := fmt.Sprintf("Всього власників: %d", count)
	return widget.NewCard("", "", widget.NewLabel(statsText))
}

// loadOwners завантажує список власників.
func (s *OwnersScreen) loadOwners() {
	owners, err := s.ownerService.GetAllOwners(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження: %w", err), s.window)
		return
	}

	s.owners = owners
	s.ownersList.Refresh()
}

// searchOwners шукає власників за запитом.
func (s *OwnersScreen) searchOwners(query string) {
	if query == "" {
		s.loadOwners()
		return
	}

	if len(query) < 2 {
		return
	}

	owners, err := s.ownerService.SearchOwners(s.currentUser.ID, query)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.owners = owners
	s.ownersList.Refresh()
}

// showOwnerDetails показує детальну інформацію про власника.
func (s *OwnersScreen) showOwnerDetails(owner *domain.Owner) {
	// Отримуємо зведену інформацію
	summary, err := s.ownerService.GetOwnerSummary(s.currentUser.ID, owner.ID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	summaryLabel := widget.NewLabel(summary)
	summaryLabel.Wrapping = fyne.TextWrapWord

	// Кнопки дій
	editBtn := widget.NewButton("Редагувати", func() {
		s.showEditOwnerDialog(owner)
	})

	deleteBtn := widget.NewButton("Деактивувати", func() {
		s.confirmDeactivate(owner)
	})
	deleteBtn.Importance = widget.DangerImportance

	// Перевірка прав
	if !s.currentUser.IsAdmin() && !s.currentUser.CanManageFinances() {
		editBtn.Disable()
	}
	if !s.currentUser.IsAdmin() {
		deleteBtn.Disable()
	}

	buttons := container.NewHBox(editBtn, deleteBtn)

	content := container.NewVBox(
		summaryLabel,
		widget.NewSeparator(),
		buttons,
	)

	d := dialog.NewCustom("Деталі власника", "Закрити", content, s.window)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// showAddOwnerDialog показує діалог додавання власника.
func (s *OwnersScreen) showAddOwnerDialog() {
	owner := &domain.Owner{}

	form := s.createOwnerForm(owner)

	dialog.ShowCustomConfirm(
		"Додати власника",
		"Додати",
		"Скасувати",
		form,
		func(add bool) {
			if !add {
				return
			}

			err := s.ownerService.CreateOwner(s.currentUser.ID, owner)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Власника додано!", s.window)
			s.loadOwners()
		},
		s.window,
	)
}

// showEditOwnerDialog показує діалог редагування власника.
func (s *OwnersScreen) showEditOwnerDialog(owner *domain.Owner) {
	// Копія для редагування
	editOwner := *owner

	form := s.createOwnerForm(&editOwner)

	dialog.ShowCustomConfirm(
		"Редагувати власника",
		"Зберегти",
		"Скасувати",
		form,
		func(save bool) {
			if !save {
				return
			}

			err := s.ownerService.UpdateOwner(s.currentUser.ID, &editOwner)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Зміни збережено!", s.window)
			s.loadOwners()
		},
		s.window,
	)
}

// createOwnerForm створює форму для введення даних власника.
func (s *OwnersScreen) createOwnerForm(owner *domain.Owner) *container.Scroll {
	// Основні дані
	fullNameEntry := widget.NewEntry()
	fullNameEntry.SetPlaceHolder("Прізвище Ім'я По батькові")
	fullNameEntry.SetText(owner.FullName)
	fullNameEntry.OnChanged = func(val string) { owner.FullName = val }

	taxIDEntry := widget.NewEntry()
	taxIDEntry.SetPlaceHolder("1234567890 (10 цифр)")
	taxIDEntry.SetText(owner.TaxID)
	taxIDEntry.OnChanged = func(val string) { owner.TaxID = val }

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("+380501234567")
	phoneEntry.SetText(owner.Phone)
	phoneEntry.OnChanged = func(val string) { owner.Phone = val }

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("owner@example.com")
	emailEntry.SetText(owner.Email)
	emailEntry.OnChanged = func(val string) { owner.Email = val }

	altPhoneEntry := widget.NewEntry()
	altPhoneEntry.SetPlaceHolder("+380672345678")
	altPhoneEntry.SetText(owner.AlternativePhone)
	altPhoneEntry.OnChanged = func(val string) { owner.AlternativePhone = val }

	// Паспортні дані
	passportSeriesEntry := widget.NewEntry()
	passportSeriesEntry.SetPlaceHolder("АА")
	passportSeriesEntry.SetText(owner.PassportSeries)
	passportSeriesEntry.OnChanged = func(val string) { owner.PassportSeries = val }

	passportNumberEntry := widget.NewEntry()
	passportNumberEntry.SetPlaceHolder("123456")
	passportNumberEntry.SetText(owner.PassportNumber)
	passportNumberEntry.OnChanged = func(val string) { owner.PassportNumber = val }

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Додаткова інформація...")
	notesEntry.SetText(owner.Notes)
	notesEntry.OnChanged = func(val string) { owner.Notes = val }

	form := container.NewVBox(
		widget.NewLabel("Основні дані:"),
		widget.NewFormItem("ПІБ", fullNameEntry).Widget,
		widget.NewFormItem("ІПН", taxIDEntry).Widget,
		widget.NewFormItem("Телефон", phoneEntry).Widget,
		widget.NewFormItem("Email", emailEntry).Widget,
		widget.NewFormItem("Додатковий телефон", altPhoneEntry).Widget,
		widget.NewSeparator(),
		widget.NewLabel("Паспортні дані (опціонально):"),
		widget.NewFormItem("Серія", passportSeriesEntry).Widget,
		widget.NewFormItem("Номер", passportNumberEntry).Widget,
		widget.NewSeparator(),
		widget.NewLabel("Примітки:"),
		notesEntry,
	)

	return container.NewVScroll(form)
}

// confirmDeactivate підтверджує деактивацію власника.
func (s *OwnersScreen) confirmDeactivate(owner *domain.Owner) {
	dialog.ShowConfirm(
		"Деактивація власника",
		fmt.Sprintf("Деактивувати власника %s?\n\nУвага: неможливо деактивувати власника з непогашеним боргом!", owner.FullName),
		func(confirm bool) {
			if !confirm {
				return
			}

			err := s.ownerService.DeactivateOwner(s.currentUser.ID, owner.ID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Власника деактивовано!", s.window)
			s.loadOwners()
		},
		s.window,
	)
}

// showDebtors показує список власників з боргом.
func (s *OwnersScreen) showDebtors() {
	debtors, err := s.ownerService.GetOwnersWithDebt(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	if len(debtors) == 0 {
		dialog.ShowInformation("Боржники", "Власників з боргом немає! ✓", s.window)
		return
	}

	// Формуємо список
	var debtorsList string
	totalDebt := 0.0
	for ownerID, debt := range debtors {
		owner, err := s.ownerService.GetOwnerByID(s.currentUser.ID, ownerID)
		if err != nil {
			continue
		}
		debtorsList += fmt.Sprintf("%s: %.2f грн\n", owner.FullName, debt)
		totalDebt += debt
	}

	debtorsList += fmt.Sprintf("\nЗагальний борг: %.2f грн", totalDebt)

	debtorsLabel := widget.NewLabel(debtorsList)
	debtorsLabel.Wrapping = fyne.TextWrapWord

	d := dialog.NewCustom(
		fmt.Sprintf("Боржники (%d)", len(debtors)),
		"Закрити",
		container.NewVScroll(debtorsLabel),
		s.window,
	)
	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}
