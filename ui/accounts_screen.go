package ui

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AccountsScreen представляє екран управління особистими рахунками.
type AccountsScreen struct {
	window           fyne.Window
	accountService   *service.AccountService
	apartmentService *service.ApartmentService
	ownerService     *service.OwnerService
	currentUser      *domain.User

	// UI елементи
	accountsList *widget.List
	filterSelect *widget.Select
	accounts     []*domain.PersonalAccount
}

// NewAccountsScreen створює новий екран рахунків.
func NewAccountsScreen(
	window fyne.Window,
	accountService *service.AccountService,
	apartmentService *service.ApartmentService,
	ownerService *service.OwnerService,
	currentUser *domain.User,
) *AccountsScreen {
	screen := &AccountsScreen{
		window:           window,
		accountService:   accountService,
		apartmentService: apartmentService,
		ownerService:     ownerService,
		currentUser:      currentUser,
		accounts:         []*domain.PersonalAccount{},
	}

	screen.initUI()
	screen.loadAccounts()

	return screen
}

// initUI ініціалізує UI елементи.
func (s *AccountsScreen) initUI() {

	// Список рахунків
	s.accountsList = widget.NewList(
		func() int {
			return len(s.accounts)
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
			if id >= len(s.accounts) {
				return
			}

			account := s.accounts[id]
			box := obj.(*fyne.Container)

			infoBox := box.Objects[0].(*fyne.Container)
			titleLabel := infoBox.Objects[0].(*widget.Label)
			detailsLabel := infoBox.Objects[1].(*widget.Label)
			balanceLabel := box.Objects[2].(*widget.Label)

			// Отримуємо дані квартири та власника для відображення
			apartment, _ := s.apartmentService.GetApartmentByID(s.currentUser.ID, account.ApartmentID)
			owner, _ := s.ownerService.GetOwnerByID(s.currentUser.ID, account.OwnerID)

			apartmentNum := fmt.Sprintf("#%d", account.ApartmentID)
			if apartment != nil {
				apartmentNum = apartment.ApartmentNumber
			}

			ownerName := fmt.Sprintf("ID:%d", account.OwnerID)
			if owner != nil {
				ownerName = owner.FullName
			}

			// Номер рахунку та статус
			status := "🟢"
			if account.IsClosed() {
				status = "🔴"
			} else if account.HasDebt() {
				status = "⚠️"
			}

			titleLabel.SetText(fmt.Sprintf("%s Рахунок %s", status, account.AccountNumber))
			titleLabel.TextStyle = fyne.TextStyle{Bold: true}

			detailsLabel.SetText(fmt.Sprintf("Кв. %s • %s", apartmentNum, ownerName))

			// Баланс
			balanceText := fmt.Sprintf("%.2f грн", account.CurrentBalance)
			if account.HasDebt() {
				balanceLabel.SetText(balanceText)
				balanceLabel.Importance = widget.DangerImportance
			} else if account.GetOverpaymentAmount() > 0 {
				balanceLabel.SetText("+" + balanceText)
				balanceLabel.Importance = widget.SuccessImportance
			} else {
				balanceLabel.SetText(balanceText)
				balanceLabel.Importance = widget.MediumImportance
			}
		},
	)

	s.accountsList.OnSelected = func(id widget.ListItemID) {
		if id >= len(s.accounts) {
			return
		}
		s.showAccountDetails(s.accounts[id])
		s.accountsList.UnselectAll()
	}

	// Фільтр
	s.filterSelect = widget.NewSelect(
		[]string{"Всі рахунки", "Тільки з боргом", "Тільки активні"},
		func(selected string) {
			s.filterAccounts(selected)
		},
	)
	s.filterSelect.SetSelected("Всі рахунки")
}

// Render повертає UI компонент екрану.
func (s *AccountsScreen) Render() fyne.CanvasObject {
	// Тулбар
	toolbar := s.createToolbar()

	// Фільтр
	filterContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Фільтр:"),
		nil,
		s.filterSelect,
	)

	// Статистика
	stats := s.createStatsCard()

	// Контент
	content := container.NewBorder(
		container.NewVBox(
			toolbar,
			filterContainer,
			stats,
		),
		nil,
		nil,
		nil,
		s.accountsList,
	)

	return container.NewPadded(content)
}

// createToolbar створює панель інструментів.
func (s *AccountsScreen) createToolbar() *fyne.Container {
	addBtn := widget.NewButtonWithIcon("Створити рахунок", theme.ContentAddIcon(), func() {
		s.showCreateAccountDialog()
	})
	addBtn.Importance = widget.HighImportance

	// Перевірка прав
	if !s.currentUser.IsAdmin() && !s.currentUser.CanManageFinances() {
		addBtn.Disable()
	}

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		s.loadAccounts()
	})

	recalculateBtn := widget.NewButton("Перерахувати баланси", func() {
		s.recalculateAllBalances()
	})
	if !s.currentUser.IsAdmin() {
		recalculateBtn.Disable()
	}

	return container.NewHBox(
		addBtn,
		recalculateBtn,
		layout.NewSpacer(),
		refreshBtn,
	)
}

// createStatsCard створює картку зі статистикою.
func (s *AccountsScreen) createStatsCard() *widget.Card {
	totalDebt, _ := s.accountService.GetTotalDebt(s.currentUser.ID)
	totalOverpayment, _ := s.accountService.GetTotalOverpayment(s.currentUser.ID)
	count, _ := s.accountService.CountOwners(s.currentUser.ID)

	statsText := fmt.Sprintf(
		"Рахунків: %d • Борг: %.2f грн • Переплата: %.2f грн • Баланс: %.2f грн",
		count,
		totalDebt,
		totalOverpayment,
		totalOverpayment-totalDebt,
	)

	return widget.NewCard("", "", widget.NewLabel(statsText))
}

// loadAccounts завантажує список рахунків.
func (s *AccountsScreen) loadAccounts() {
	accounts, err := s.accountService.GetAllAccounts(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження: %w", err), s.window)
		return
	}

	s.accounts = accounts
	s.accountsList.Refresh()
}

// filterAccounts фільтрує рахунки.
func (s *AccountsScreen) filterAccounts(filter string) {
	switch filter {
	case "Тільки з боргом":
		accounts, err := s.accountService.GetDebtorAccounts(s.currentUser.ID)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		s.accounts = accounts

	case "Тільки активні":
		s.loadAccounts()
		// Фільтруємо тільки активні
		var active []*domain.PersonalAccount
		for _, acc := range s.accounts {
			if acc.IsActive && !acc.IsClosed() {
				active = append(active, acc)
			}
		}
		s.accounts = active

	default:
		s.loadAccounts()
	}

	s.accountsList.Refresh()
}

// showAccountDetails показує детальну інформацію про рахунок.
func (s *AccountsScreen) showAccountDetails(account *domain.PersonalAccount) {
	// Отримуємо виписку
	statement, err := s.accountService.GetAccountStatement(s.currentUser.ID, account.ID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	statementLabel := widget.NewLabel(statement)
	statementLabel.Wrapping = fyne.TextWrapWord

	// Кнопки дій
	recalcBtn := widget.NewButton("Перерахувати баланс", func() {
		s.recalculateBalance(account.ID)
	})

	closeBtn := widget.NewButton("Закрити рахунок", func() {
		s.confirmCloseAccount(account)
	})
	closeBtn.Importance = widget.WarningImportance

	if account.IsClosed() {
		closeBtn.SetText("Відкрити рахунок")
		closeBtn.OnTapped = func() {
			s.reopenAccount(account.ID)
		}
	}

	// Перевірка прав
	if !s.currentUser.IsAdmin() && !s.currentUser.CanManageFinances() {
		recalcBtn.Disable()
	}
	if !s.currentUser.IsAdmin() {
		closeBtn.Disable()
	}

	buttons := container.NewHBox(recalcBtn, closeBtn)

	content := container.NewVBox(
		statementLabel,
		widget.NewSeparator(),
		buttons,
	)

	d := dialog.NewCustom("Особистий рахунок", "Закрити", content, s.window)
	d.Resize(fyne.NewSize(500, 500))
	d.Show()
}

// showCreateAccountDialog показує діалог створення рахунку.
func (s *AccountsScreen) showCreateAccountDialog() {
	// Отримуємо списки для вибору
	apartments, err := s.apartmentService.GetAllApartments(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	owners, err := s.ownerService.GetAllOwners(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	if len(apartments) == 0 {
		dialog.ShowInformation("Увага", "Спочатку додайте квартири!", s.window)
		return
	}
	if len(owners) == 0 {
		dialog.ShowInformation("Увага", "Спочатку додайте власників!", s.window)
		return
	}

	// Вибір квартири
	apartmentOptions := make([]string, len(apartments))
	for i, apt := range apartments {
		apartmentOptions[i] = fmt.Sprintf("Кв. %s (%s)", apt.ApartmentNumber, apt.OwnerName)
	}
	apartmentSelect := widget.NewSelect(apartmentOptions, nil)
	apartmentSelect.SetSelected(apartmentOptions[0])

	// Вибір власника
	ownerOptions := make([]string, len(owners))
	for i, owner := range owners {
		ownerOptions[i] = fmt.Sprintf("%s (ІПН: %s)", owner.FullName, owner.TaxID)
	}
	ownerSelect := widget.NewSelect(ownerOptions, nil)
	ownerSelect.SetSelected(ownerOptions[0])

	form := container.NewVBox(
		widget.NewLabel("Квартира:"),
		apartmentSelect,
		widget.NewLabel("Власник:"),
		ownerSelect,
		widget.NewLabel(""),
		widget.NewLabel("Номер рахунку буде згенеровано автоматично."),
	)

	dialog.ShowCustomConfirm(
		"Створити особистий рахунок",
		"Створити",
		"Скасувати",
		form,
		func(create bool) {
			if !create {
				return
			}

			// Визначаємо вибрані ID
			apartmentID := apartments[apartmentSelect.SelectedIndex()].ID
			ownerID := owners[ownerSelect.SelectedIndex()].ID

			account, err := s.accountService.CreateAccount(s.currentUser.ID, apartmentID, ownerID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх",
				fmt.Sprintf("Рахунок створено!\nНомер: %s", account.AccountNumber),
				s.window)
			s.loadAccounts()
		},
		s.window,
	)
}

// recalculateBalance перераховує баланс рахунку.
func (s *AccountsScreen) recalculateBalance(accountID int) {
	newBalance, err := s.accountService.RecalculateBalance(s.currentUser.ID, accountID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	dialog.ShowInformation("Успіх",
		fmt.Sprintf("Баланс перераховано: %.2f грн", newBalance),
		s.window)
	s.loadAccounts()
}

// recalculateAllBalances перераховує всі баланси.
func (s *AccountsScreen) recalculateAllBalances() {
	dialog.ShowConfirm(
		"Перерахунок балансів",
		"Перерахувати баланси всіх рахунків?\n\nЦе може зайняти деякий час.",
		func(confirm bool) {
			if !confirm {
				return
			}

			count, err := s.accountService.RecalculateAllBalances(s.currentUser.ID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх",
				fmt.Sprintf("Оновлено %d рахунків!", count),
				s.window)
			s.loadAccounts()
		},
		s.window,
	)
}

// confirmCloseAccount підтверджує закриття рахунку.
func (s *AccountsScreen) confirmCloseAccount(account *domain.PersonalAccount) {
	if account.HasDebt() {
		dialog.ShowError(
			fmt.Errorf("не можна закрити рахунок з боргом %.2f грн", account.GetDebtAmount()),
			s.window)
		return
	}

	dialog.ShowConfirm(
		"Закриття рахунку",
		fmt.Sprintf("Закрити рахунок %s?\n\nЦю операцію можна буде скасувати пізніше.", account.AccountNumber),
		func(confirm bool) {
			if !confirm {
				return
			}

			err := s.accountService.CloseAccount(s.currentUser.ID, account.ID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Рахунок закрито!", s.window)
			s.loadAccounts()
		},
		s.window,
	)
}

// reopenAccount відкриває закритий рахунок.
func (s *AccountsScreen) reopenAccount(accountID int) {
	err := s.accountService.ReopenAccount(s.currentUser.ID, accountID)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	dialog.ShowInformation("Успіх", "Рахунок відкрито!", s.window)
	s.loadAccounts()
}
