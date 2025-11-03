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

// PaymentsScreen представляє екран управління платежами.
type PaymentsScreen struct {
	window           fyne.Window
	paymentService   *service.PaymentService
	apartmentService *service.ApartmentService
	currentUser      *domain.User

	// UI елементи
	paymentsList   *widget.List
	payments       []*domain.Payment
	filterSelect   *widget.Select
	apartmentsList []*domain.Apartment
}

// NewPaymentsScreen створює новий екран платежів.
func NewPaymentsScreen(
	window fyne.Window,
	paymentService *service.PaymentService,
	apartmentService *service.ApartmentService,
	currentUser *domain.User,
) *PaymentsScreen {
	// КРИТИЧНО: Перевірка вхідних параметрів
	if apartmentService == nil {
		panic("КРИТИЧНА ПОМИЛКА: ApartmentService не може бути nil!")
	}
	if paymentService == nil {
		panic("КРИТИЧНА ПОМИЛКА: PaymentService не може бути nil!")
	}
	if window == nil {
		panic("КРИТИЧНА ПОМИЛКА: Window не може бути nil!")
	}
	if currentUser == nil {
		panic("КРИТИЧНА ПОМИЛКА: CurrentUser не може бути nil!")
	}
	screen := &PaymentsScreen{
		window:           window,
		paymentService:   paymentService,
		apartmentService: apartmentService,
		currentUser:      currentUser,
		payments:         []*domain.Payment{},
		apartmentsList:   []*domain.Apartment{},
	}

	screen.initUI()
	screen.loadApartments()
	screen.loadPayments()

	return screen
}

// initUI ініціалізує UI елементи.
func (s *PaymentsScreen) initUI() {
	// Список платежів
	s.paymentsList = widget.NewList(
		func() int {
			return len(s.payments)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				container.NewVBox(
					widget.NewLabel("Template"),
					widget.NewLabel("Template"),
				),
				layout.NewSpacer(),
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(s.payments) {
				return
			}

			payment := s.payments[id]
			box := obj.(*fyne.Container)

			infoBox := box.Objects[0].(*fyne.Container)
			titleLabel := infoBox.Objects[0].(*widget.Label)
			detailsLabel := infoBox.Objects[1].(*widget.Label)
			amountLabel := box.Objects[2].(*widget.Label)

			// Іконка типу платежу
			typeIcon := "💰"
			if payment.IsCharge() {
				typeIcon = "📊"
			}

			titleLabel.SetText(fmt.Sprintf("%s %s", typeIcon, payment.Description))
			titleLabel.TextStyle = fyne.TextStyle{Bold: true}

			detailsLabel.SetText(fmt.Sprintf(
				"Кв. %s • %s • %s",
				s.getApartmentNumber(payment.ApartmentID),
				payment.GetCategoryDisplay(),
				payment.PaymentDate.Format("02.01.2006"),
			))

			// Сума зі знаком
			amountText := fmt.Sprintf("%.2f грн", payment.GetDisplayAmount())
			if payment.IsIncoming() {
				amountText = "+ " + amountText
				amountLabel.Importance = widget.SuccessImportance
			} else {
				amountText = "- " + amountText
				amountLabel.Importance = widget.DangerImportance
			}
			amountLabel.SetText(amountText)
		},
	)

	s.paymentsList.OnSelected = func(id widget.ListItemID) {
		if id >= len(s.payments) {
			return
		}
		s.showPaymentDetails(s.payments[id])
		s.paymentsList.UnselectAll()
	}

	// Фільтр по квартирах
	s.filterSelect = widget.NewSelect([]string{"Всі квартири"}, func(value string) {
		s.filterPayments(value)
	})
	s.filterSelect.SetSelected("Всі квартири")
}

// Render повертає UI компонент екрану.
func (s *PaymentsScreen) Render() fyne.CanvasObject {
	// Тулбар
	toolbar := s.createToolbar()

	// Фільтр
	filterContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Квартира:"),
		nil,
		s.filterSelect,
	)

	// Основний контент
	content := container.NewBorder(
		container.NewVBox(toolbar, filterContainer),
		nil,
		nil,
		nil,
		s.paymentsList,
	)

	return container.NewPadded(content)
}

// createToolbar створює панель інструментів.
func (s *PaymentsScreen) createToolbar() *fyne.Container {
	addPaymentBtn := widget.NewButtonWithIcon("Додати платіж", theme.ContentAddIcon(), func() {
		s.showAddPaymentDialog()
	})
	addPaymentBtn.Importance = widget.HighImportance

	addChargeBtn := widget.NewButtonWithIcon("Нарахування", theme.DocumentCreateIcon(), func() {
		s.showAddChargeDialog()
	})

	monthlyChargesBtn := widget.NewButton("Місячні нарахування", func() {
		s.showMonthlyChargesDialog()
	})

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		s.loadPayments()
	})

	return container.NewHBox(
		addPaymentBtn,
		addChargeBtn,
		monthlyChargesBtn,
		layout.NewSpacer(),
		refreshBtn,
	)
}

// loadApartments завантажує список квартир.
func (s *PaymentsScreen) loadApartments() {
	apartments, err := s.apartmentService.GetAllApartments(s.currentUser.ID)
	if err != nil {
		return
	}

	s.apartmentsList = apartments

	// Оновлюємо опції фільтра
	options := []string{"Всі квартири"}
	for _, apt := range apartments {
		options = append(options, fmt.Sprintf("Кв. %s", apt.ApartmentNumber))
	}
	s.filterSelect.Options = options
}

// loadPayments завантажує список платежів.
func (s *PaymentsScreen) loadPayments() {
	if s.paymentService == nil {
		dialog.ShowError(fmt.Errorf("помилка: сервіс платежів не ініціалізовано"), s.window)
		return
	}
	payments, err := s.paymentService.GetRecentPayments(s.currentUser.ID, 100)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження: %w", err), s.window)
		return
	}
	s.payments = payments
	s.paymentsList.Refresh()
}

// filterPayments фільтрує платежі по квартирі.
func (s *PaymentsScreen) filterPayments(selection string) {
	if selection == "Всі квартири" {
		s.loadPayments()
		return
	}

	// Знаходимо ID квартири
	var apartmentID int
	for _, apt := range s.apartmentsList {
		if fmt.Sprintf("Кв. %s", apt.ApartmentNumber) == selection {
			apartmentID = apt.ID
			break
		}
	}

	if apartmentID == 0 {
		return
	}

	payments, err := s.paymentService.GetPaymentsByApartment(s.currentUser.ID, apartmentID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.payments = payments
	s.paymentsList.Refresh()
}

// showPaymentDetails показує деталі платежу.
func (s *PaymentsScreen) showPaymentDetails(payment *domain.Payment) {
	typeText := "Платіж"
	if payment.IsCharge() {
		typeText = "Нарахування"
	}

	details := fmt.Sprintf(
		"%s\n\n"+
			"Квартира: %s\n"+
			"Категорія: %s\n"+
			"Сума: %.2f грн\n"+
			"Дата: %s\n"+
			"Період: %s\n"+
			"Опис: %s\n",
		typeText,
		s.getApartmentNumber(payment.ApartmentID),
		payment.GetCategoryDisplay(),
		payment.GetDisplayAmount(),
		payment.PaymentDate.Format("02.01.2006"),
		payment.FormatPeriod(),
		payment.Description,
	)

	if payment.Notes != "" {
		details += fmt.Sprintf("\nПримітки: %s\n", payment.Notes)
	}

	detailsLabel := widget.NewLabel(details)

	// Кнопки дій (тільки для адмінів)
	var buttons *fyne.Container
	if s.currentUser.IsAdmin() {
		deleteBtn := widget.NewButton("Видалити", func() {
			s.confirmDelete(payment)
		})
		deleteBtn.Importance = widget.DangerImportance
		buttons = container.NewHBox(deleteBtn)
	}

	content := container.NewVBox(detailsLabel)
	if buttons != nil {
		content.Add(buttons)
	}

	d := dialog.NewCustom("Деталі платежу", "Закрити", content, s.window)
	d.Resize(fyne.NewSize(400, 350))
	d.Show()
}

// showAddPaymentDialog показує діалог додавання платежу.
func (s *PaymentsScreen) showAddPaymentDialog() {
	// Вибір квартири
	apartmentSelect := widget.NewSelect([]string{}, nil)
	for _, apt := range s.apartmentsList {
		apartmentSelect.Options = append(apartmentSelect.Options,
			fmt.Sprintf("Кв. %s - %s", apt.ApartmentNumber, apt.OwnerName))
	}
	if len(apartmentSelect.Options) > 0 {
		apartmentSelect.SetSelectedIndex(0)
	}

	// Сума
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("500.00")

	// Категорія
	categorySelect := widget.NewSelect([]string{
		"Утримання будинку",
		"Поточний ремонт",
		"Комунальні послуги",
		"Опалення",
		"Водопостачання",
		"Електроенергія",
		"Газопостачання",
		"Інше",
	}, nil)
	categorySelect.SetSelected("Утримання будинку")

	// Опис
	descriptionEntry := widget.NewEntry()
	descriptionEntry.SetPlaceHolder("Опис платежу")

	// Період (поточний місяць)
	currentPeriod := s.paymentService.GetCurrentPeriod()
	periodEntry := widget.NewEntry()
	periodEntry.SetText(currentPeriod)

	// Примітки
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Додаткова інформація...")

	form := container.NewVBox(
		widget.NewLabel("Квартира:"),
		apartmentSelect,
		widget.NewLabel("Сума (грн):"),
		amountEntry,
		widget.NewLabel("Категорія:"),
		categorySelect,
		widget.NewLabel("Опис:"),
		descriptionEntry,
		widget.NewLabel("Період (YYYY-MM):"),
		periodEntry,
		widget.NewLabel("Примітки:"),
		notesEntry,
	)

	dialog.NewCustomConfirm(
		"Додати платіж",
		"Додати",
		"Скасувати",
		form,
		func(confirm bool) {
			if !confirm {
				return
			}

			// Парсимо дані
			apartmentID := s.getApartmentIDFromSelection(apartmentSelect.Selected)
			amount, err := strconv.ParseFloat(amountEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("невірна сума"), s.window)
				return
			}

			category := s.getCategoryFromDisplay(categorySelect.Selected)

			_, err = s.paymentService.CreatePayment(
				s.currentUser.ID,
				apartmentID,
				amount,
				category,
				descriptionEntry.Text,
				time.Now(),
				periodEntry.Text,
				notesEntry.Text,
			)

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Платіж додано!", s.window)
			s.loadPayments()
		},
		s.window,
	).Show()
}

// showAddChargeDialog показує діалог додавання нарахування.
func (s *PaymentsScreen) showAddChargeDialog() {
	// Вибір квартири
	apartmentSelect := widget.NewSelect([]string{}, nil)
	for _, apt := range s.apartmentsList {
		apartmentSelect.Options = append(apartmentSelect.Options,
			fmt.Sprintf("Кв. %s - %s", apt.ApartmentNumber, apt.OwnerName))
	}
	if len(apartmentSelect.Options) > 0 {
		apartmentSelect.SetSelectedIndex(0)
	}

	// Сума
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("500.00")

	// Категорія
	categorySelect := widget.NewSelect([]string{
		"Утримання будинку",
		"Поточний ремонт",
		"Комунальні послуги",
		"Опалення",
		"Водопостачання",
		"Електроенергія",
		"Газопостачання",
		"Інше",
	}, nil)
	categorySelect.SetSelected("Утримання будинку")

	// Опис
	descriptionEntry := widget.NewEntry()
	descriptionEntry.SetPlaceHolder("Опис нарахування")

	// Період
	currentPeriod := s.paymentService.GetCurrentPeriod()
	periodEntry := widget.NewEntry()
	periodEntry.SetText(currentPeriod)

	form := container.NewVBox(
		widget.NewLabel("Квартира:"),
		apartmentSelect,
		widget.NewLabel("Сума (грн):"),
		amountEntry,
		widget.NewLabel("Категорія:"),
		categorySelect,
		widget.NewLabel("Опис:"),
		descriptionEntry,
		widget.NewLabel("Період (YYYY-MM):"),
		periodEntry,
	)

	dialog.NewCustomConfirm(
		"Додати нарахування",
		"Додати",
		"Скасувати",
		form,
		func(confirm bool) {
			if !confirm {
				return
			}

			apartmentID := s.getApartmentIDFromSelection(apartmentSelect.Selected)
			amount, err := strconv.ParseFloat(amountEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("невірна сума"), s.window)
				return
			}

			category := s.getCategoryFromDisplay(categorySelect.Selected)

			_, err = s.paymentService.CreateCharge(
				s.currentUser.ID,
				apartmentID,
				amount,
				category,
				descriptionEntry.Text,
				periodEntry.Text,
				"",
			)

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Нарахування додано!", s.window)
			s.loadPayments()
		},
		s.window,
	).Show()
}

// showMonthlyChargesDialog показує діалог масових нарахувань.
func (s *PaymentsScreen) showMonthlyChargesDialog() {
	// Період
	currentPeriod := s.paymentService.GetCurrentPeriod()
	periodEntry := widget.NewEntry()
	periodEntry.SetText(currentPeriod)

	// Тариф на утримання
	maintenanceEntry := widget.NewEntry()
	maintenanceEntry.SetPlaceHolder("15.00")
	maintenanceEntry.SetText("15.00")

	// Тариф на комунальні
	utilitiesEntry := widget.NewEntry()
	utilitiesEntry.SetPlaceHolder("10.00")
	utilitiesEntry.SetText("10.00")

	form := container.NewVBox(
		widget.NewLabel("Період (YYYY-MM):"),
		periodEntry,
		widget.NewLabel("Тариф на утримання (грн/м²):"),
		maintenanceEntry,
		widget.NewLabel("Тариф на комунальні (грн/м²):"),
		utilitiesEntry,
		widget.NewLabel(""),
		widget.NewLabel("Буде створено нарахування для всіх квартир."),
	)

	dialog.NewCustomConfirm(
		"Місячні нарахування",
		"Створити",
		"Скасувати",
		form,
		func(confirm bool) {
			if !confirm {
				return
			}

			maintenanceRate, err := strconv.ParseFloat(maintenanceEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("невірний тариф утримання"), s.window)
				return
			}

			utilitiesRate, err := strconv.ParseFloat(utilitiesEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("невірний тариф комунальних"), s.window)
				return
			}

			count, err := s.paymentService.CreateMonthlyCharges(
				s.currentUser.ID,
				periodEntry.Text,
				maintenanceRate,
				utilitiesRate,
			)

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation(
				"Успіх",
				fmt.Sprintf("Створено %d нарахувань!", count),
				s.window,
			)
			s.loadPayments()
		},
		s.window,
	).Show()
}

// confirmDelete підтверджує видалення платежу.
func (s *PaymentsScreen) confirmDelete(payment *domain.Payment) {
	dialog.ShowConfirm(
		"Видалення платежу",
		fmt.Sprintf("Видалити платіж на суму %.2f грн?", payment.Amount),
		func(confirm bool) {
			if !confirm {
				return
			}

			err := s.paymentService.DeletePayment(s.currentUser.ID, payment.ID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Платіж видалено!", s.window)
			s.loadPayments()
		},
		s.window,
	)
}

// Допоміжні функції

func (s *PaymentsScreen) getApartmentNumber(apartmentID int) string {
	for _, apt := range s.apartmentsList {
		if apt.ID == apartmentID {
			return apt.ApartmentNumber
		}
	}
	return fmt.Sprintf("#%d", apartmentID)
}

func (s *PaymentsScreen) getApartmentIDFromSelection(selection string) int {
	for _, apt := range s.apartmentsList {
		if fmt.Sprintf("Кв. %s - %s", apt.ApartmentNumber, apt.OwnerName) == selection {
			return apt.ID
		}
	}
	return 0
}

func (s *PaymentsScreen) getCategoryFromDisplay(display string) domain.PaymentCategory {
	categories := map[string]domain.PaymentCategory{
		"Утримання будинку":  domain.CategoryMaintenance,
		"Поточний ремонт":    domain.CategoryRepairs,
		"Комунальні послуги": domain.CategoryUtilities,
		"Опалення":           domain.CategoryHeating,
		"Водопостачання":     domain.CategoryWater,
		"Електроенергія":     domain.CategoryElectricity,
		"Газопостачання":     domain.CategoryGas,
		"Інше":               domain.CategoryOther,
	}

	if cat, ok := categories[display]; ok {
		return cat
	}
	return domain.CategoryOther
}
