package ui

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/service"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// OSBBSettingsScreen представляє екран налаштувань ОСББ.
type OSBBSettingsScreen struct {
	window      fyne.Window
	osbbService *service.OSBBService
	currentUser *domain.User

	// UI елементи
	osbb            *domain.OSBB
	infoCard        *widget.Card
	statsCard       *widget.Card
	bankDetailsCard *widget.Card
}

// NewOSBBSettingsScreen створює новий екран налаштувань ОСББ.
func NewOSBBSettingsScreen(
	window fyne.Window,
	osbbService *service.OSBBService,
	currentUser *domain.User,
) *OSBBSettingsScreen {
	screen := &OSBBSettingsScreen{
		window:      window,
		osbbService: osbbService,
		currentUser: currentUser,
	}

	screen.loadOSBB()
	return screen
}

// loadOSBB завантажує дані ОСББ з бази.
func (s *OSBBSettingsScreen) loadOSBB() {
	osbb, err := s.osbbService.GetOSBB(s.currentUser.ID)
	if err == domain.ErrOSBBNotFound {
		// ОСББ ще не створено
		s.osbb = nil
		return
	}
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження ОСББ: %w", err), s.window)
		return
	}
	s.osbb = osbb
}

// Render повертає UI компонент екрану.
func (s *OSBBSettingsScreen) Render() fyne.CanvasObject {
	// Якщо ОСББ не створено, показуємо форму створення
	if s.osbb == nil {
		return s.renderSetupWizard()
	}

	// ОСББ існує, показуємо інформацію та налаштування
	return s.renderSettings()
}

// renderSetupWizard показує майстер початкового налаштування.
func (s *OSBBSettingsScreen) renderSetupWizard() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"Налаштування ОСББ",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	description := widget.NewLabel(
		"Вітаємо! Для початку роботи необхідно налаштувати основні дані вашої організації ОСББ.",
	)
	description.Wrapping = fyne.TextWrapWord

	setupBtn := widget.NewButtonWithIcon(
		"Почати налаштування",
		theme.SettingsIcon(),
		func() {
			s.showCreateOSBBDialog()
		},
	)
	setupBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(widget.NewIcon(theme.InfoIcon())),
		layout.NewSpacer(),
		container.NewCenter(title),
		container.NewCenter(description),
		layout.NewSpacer(),
		container.NewCenter(setupBtn),
		layout.NewSpacer(),
	)

	return container.NewPadded(content)
}

// renderSettings показує налаштування існуючого ОСББ.
func (s *OSBBSettingsScreen) renderSettings() fyne.CanvasObject {
	// Тулбар
	toolbar := s.createToolbar()

	// Картки з інформацією
	s.infoCard = s.createInfoCard()
	s.statsCard = s.createStatsCard()
	s.bankDetailsCard = s.createBankDetailsCard()

	// Контент
	leftColumn := container.NewVBox(
		s.infoCard,
		s.statsCard,
	)

	rightColumn := container.NewVBox(
		s.bankDetailsCard,
	)

	content := container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		container.NewHSplit(leftColumn, rightColumn),
	)

	return container.NewPadded(content)
}

// createToolbar створює панель інструментів.
func (s *OSBBSettingsScreen) createToolbar() *fyne.Container {
	// Перевірка прав на редагування
	canModify, _ := s.osbbService.CanModifySettings(s.currentUser.ID)

	editBtn := widget.NewButtonWithIcon(
		"Редагувати",
		theme.DocumentCreateIcon(),
		func() {
			s.showEditOSBBDialog()
		},
	)
	if !canModify {
		editBtn.Disable()
	}

	rateBtn := widget.NewButton("Змінити тариф", func() {
		s.showChangeRateDialog()
	})
	if !s.currentUser.IsAdmin() && !s.currentUser.CanManageFinances() {
		rateBtn.Disable()
	}

	refreshBtn := widget.NewButtonWithIcon(
		"",
		theme.ViewRefreshIcon(),
		func() {
			s.refresh()
		},
	)

	return container.NewHBox(
		editBtn,
		rateBtn,
		layout.NewSpacer(),
		refreshBtn,
	)
}

// createInfoCard створює картку з основною інформацією.
func (s *OSBBSettingsScreen) createInfoCard() *widget.Card {
	if s.osbb == nil {
		return widget.NewCard("", "", widget.NewLabel("Дані відсутні"))
	}

	content := container.NewVBox(
		s.createInfoRow("Назва:", s.osbb.Name),
		s.createInfoRow("Коротка назва:", s.osbb.ShortName),
		s.createInfoRow("Адреса:", s.osbb.Address),
		s.createInfoRow("ЄДРПОУ:", s.osbb.EDRPOU),
		widget.NewSeparator(),
		s.createInfoRow("Голова правління:", s.osbb.ChairmanName),
		s.createInfoRow("Телефон:", s.osbb.ChairmanPhone),
		s.createInfoRow("Email:", s.osbb.ChairmanEmail),
		widget.NewSeparator(),
		s.createInfoRow("Базовий тариф:", fmt.Sprintf("%.2f грн/м²", s.osbb.BaseRate)),
		s.createInfoRow("Засновано:", s.osbb.FoundedAt.Format("02.01.2006")),
	)

	return widget.NewCard("Інформація про ОСББ", "", content)
}

// createStatsCard створює картку зі статистикою.
func (s *OSBBSettingsScreen) createStatsCard() *widget.Card {
	settings, err := s.osbbService.GetSettings(s.currentUser.ID)
	if err != nil {
		return widget.NewCard("Статистика", "", widget.NewLabel("Помилка завантаження"))
	}

	content := container.NewVBox(
		s.createInfoRow("Всього квартир:", fmt.Sprintf("%d", settings.TotalApartments)),
		s.createInfoRow("Загальна площа:", fmt.Sprintf("%.2f м²", settings.TotalArea)),
		s.createInfoRow("Власників:", fmt.Sprintf("%d", settings.TotalOwners)),
		s.createInfoRow("Активних рахунків:", fmt.Sprintf("%d", settings.ActiveAccounts)),
		widget.NewSeparator(),
		s.createInfoRow("Місячне нарахування:", s.calculateMonthlyCharge(settings)),
	)

	return widget.NewCard("Статистика", "", content)
}

// createBankDetailsCard створює картку з банківськими реквізитами.
func (s *OSBBSettingsScreen) createBankDetailsCard() *widget.Card {
	if s.osbb == nil {
		return widget.NewCard("", "", widget.NewLabel("Дані відсутні"))
	}

	content := container.NewVBox(
		s.createInfoRow("Банк:", s.osbb.BankName),
		s.createInfoRow("IBAN:", s.osbb.BankAccount),
		s.createInfoRow("МФО:", s.osbb.MFO),
		widget.NewSeparator(),
		widget.NewButton("Копіювати реквізити", func() {
			s.copyBankDetails()
		}),
	)

	return widget.NewCard("Банківські реквізити", "", content)
}

// createInfoRow створює рядок з інформацією.
func (s *OSBBSettingsScreen) createInfoRow(label, value string) *fyne.Container {
	labelWidget := widget.NewLabel(label)
	labelWidget.TextStyle = fyne.TextStyle{Bold: true}

	valueWidget := widget.NewLabel(value)
	valueWidget.Wrapping = fyne.TextWrapWord

	return container.NewHBox(
		labelWidget,
		layout.NewSpacer(),
		valueWidget,
	)
}

// calculateMonthlyCharge розраховує місячне нарахування.
func (s *OSBBSettingsScreen) calculateMonthlyCharge(settings *domain.OSBBSettings) string {
	charge := settings.OSBB.BaseRate * settings.TotalArea
	return fmt.Sprintf("%.2f грн", charge)
}

// showCreateOSBBDialog показує діалог створення ОСББ.
func (s *OSBBSettingsScreen) showCreateOSBBDialog() {
	osbb := &domain.OSBB{
		FoundedAt: time.Now(),
	}

	form := s.createOSBBForm(osbb, true)

	dialog.ShowCustomConfirm(
		"Створення ОСББ",
		"Створити",
		"Скасувати",
		form,
		func(create bool) {
			if !create {
				return
			}

			err := s.osbbService.CreateOSBB(s.currentUser.ID, osbb)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "ОСББ успішно створено!", s.window)
			s.refresh()
		},
		s.window,
	)
}

// showEditOSBBDialog показує діалог редагування ОСББ.
func (s *OSBBSettingsScreen) showEditOSBBDialog() {
	if s.osbb == nil {
		return
	}

	// Створюємо копію для редагування
	editOSBB := *s.osbb

	form := s.createOSBBForm(&editOSBB, false)

	dialog.ShowCustomConfirm(
		"Редагування ОСББ",
		"Зберегти",
		"Скасувати",
		form,
		func(save bool) {
			if !save {
				return
			}

			err := s.osbbService.UpdateOSBB(s.currentUser.ID, &editOSBB)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Зміни збережено!", s.window)
			s.refresh()
		},
		s.window,
	)
}

// createOSBBForm створює форму для введення даних ОСББ.
func (s *OSBBSettingsScreen) createOSBBForm(osbb *domain.OSBB, isNew bool) *container.Scroll {
	// Основні дані
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Повна назва організації")
	if !isNew {
		nameEntry.SetText(osbb.Name)
	}
	nameEntry.OnChanged = func(val string) { osbb.Name = val }

	shortNameEntry := widget.NewEntry()
	shortNameEntry.SetPlaceHolder("Коротка назва")
	if !isNew {
		shortNameEntry.SetText(osbb.ShortName)
	}
	shortNameEntry.OnChanged = func(val string) { osbb.ShortName = val }

	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Повна адреса будинку")
	if !isNew {
		addressEntry.SetText(osbb.Address)
	}
	addressEntry.OnChanged = func(val string) { osbb.Address = val }

	edrpouEntry := widget.NewEntry()
	edrpouEntry.SetPlaceHolder("12345678 (8 цифр)")
	if !isNew {
		edrpouEntry.SetText(osbb.EDRPOU)
	}
	edrpouEntry.OnChanged = func(val string) { osbb.EDRPOU = val }

	// Голова правління
	chairmanNameEntry := widget.NewEntry()
	chairmanNameEntry.SetPlaceHolder("ПІБ голови")
	if !isNew {
		chairmanNameEntry.SetText(osbb.ChairmanName)
	}
	chairmanNameEntry.OnChanged = func(val string) { osbb.ChairmanName = val }

	chairmanPhoneEntry := widget.NewEntry()
	chairmanPhoneEntry.SetPlaceHolder("+380501234567")
	if !isNew {
		chairmanPhoneEntry.SetText(osbb.ChairmanPhone)
	}
	chairmanPhoneEntry.OnChanged = func(val string) { osbb.ChairmanPhone = val }

	chairmanEmailEntry := widget.NewEntry()
	chairmanEmailEntry.SetPlaceHolder("chairman@osbb.ua")
	if !isNew {
		chairmanEmailEntry.SetText(osbb.ChairmanEmail)
	}
	chairmanEmailEntry.OnChanged = func(val string) { osbb.ChairmanEmail = val }

	// Фінанси
	baseRateEntry := widget.NewEntry()
	baseRateEntry.SetPlaceHolder("15.50")
	if !isNew {
		baseRateEntry.SetText(fmt.Sprintf("%.2f", osbb.BaseRate))
	}
	baseRateEntry.OnChanged = func(val string) {
		if rate, err := strconv.ParseFloat(val, 64); err == nil {
			osbb.BaseRate = rate
		}
	}

	// Банк
	bankNameEntry := widget.NewEntry()
	bankNameEntry.SetPlaceHolder("ПриватБанк")
	if !isNew {
		bankNameEntry.SetText(osbb.BankName)
	}
	bankNameEntry.OnChanged = func(val string) { osbb.BankName = val }

	bankAccountEntry := widget.NewEntry()
	bankAccountEntry.SetPlaceHolder("UA123456789012345678901234567")
	if !isNew {
		bankAccountEntry.SetText(osbb.BankAccount)
	}
	bankAccountEntry.OnChanged = func(val string) { osbb.BankAccount = val }

	mfoEntry := widget.NewEntry()
	mfoEntry.SetPlaceHolder("305299 (6 цифр)")
	if !isNew {
		mfoEntry.SetText(osbb.MFO)
	}
	mfoEntry.OnChanged = func(val string) { osbb.MFO = val }

	// Форма
	form := container.NewVBox(
		widget.NewLabel("Основні дані:"),
		widget.NewFormItem("Назва ОСББ", nameEntry).Widget,
		widget.NewFormItem("Коротка назва", shortNameEntry).Widget,
		widget.NewFormItem("Адреса", addressEntry).Widget,
		widget.NewFormItem("ЄДРПОУ", edrpouEntry).Widget,
		widget.NewSeparator(),
		widget.NewLabel("Голова правління:"),
		widget.NewFormItem("ПІБ", chairmanNameEntry).Widget,
		widget.NewFormItem("Телефон", chairmanPhoneEntry).Widget,
		widget.NewFormItem("Email", chairmanEmailEntry).Widget,
		widget.NewSeparator(),
		widget.NewLabel("Фінанси:"),
		widget.NewFormItem("Базовий тариф (грн/м²)", baseRateEntry).Widget,
		widget.NewSeparator(),
		widget.NewLabel("Банківські реквізити:"),
		widget.NewFormItem("Банк", bankNameEntry).Widget,
		widget.NewFormItem("IBAN", bankAccountEntry).Widget,
		widget.NewFormItem("МФО", mfoEntry).Widget,
	)

	return container.NewVScroll(form)
}

// showChangeRateDialog показує діалог зміни тарифу.
func (s *OSBBSettingsScreen) showChangeRateDialog() {
	if s.osbb == nil {
		return
	}

	currentRate := widget.NewLabel(fmt.Sprintf("Поточний тариф: %.2f грн/м²", s.osbb.BaseRate))

	newRateEntry := widget.NewEntry()
	newRateEntry.SetPlaceHolder(fmt.Sprintf("%.2f", s.osbb.BaseRate))

	form := container.NewVBox(
		currentRate,
		widget.NewLabel("Новий тариф (грн/м²):"),
		newRateEntry,
		widget.NewLabel("Діапазон: 5.00 - 50.00 грн/м²"),
	)

	dialog.ShowCustomConfirm(
		"Зміна базового тарифу",
		"Змінити",
		"Скасувати",
		form,
		func(change bool) {
			if !change {
				return
			}

			newRate, err := strconv.ParseFloat(newRateEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("невалідне значення тарифу"), s.window)
				return
			}

			err = s.osbbService.UpdateBaseRate(s.currentUser.ID, newRate)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх",
				fmt.Sprintf("Тариф змінено: %.2f → %.2f грн/м²", s.osbb.BaseRate, newRate),
				s.window)
			s.refresh()
		},
		s.window,
	)
}

// copyBankDetails копіює банківські реквізити в буфер обміну.
func (s *OSBBSettingsScreen) copyBankDetails() {
	details, err := s.osbbService.GetBankDetails(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.window.Clipboard().SetContent(details)
	dialog.ShowInformation("Скопійовано", "Реквізити скопійовано в буфер обміну", s.window)
}

// refresh оновлює екран.
func (s *OSBBSettingsScreen) refresh() {
	s.loadOSBB()
	// Перемальовуємо екран - в реальному додатку краще викликати callback
	// для оновлення всього контенту
}
