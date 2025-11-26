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
	"osbb-accounting/application/usecase/charge"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/common"
)

// ChargesScreen представляє екран управління нарахуваннями.
type ChargesScreen struct {
	window           fyne.Window
	chargeService    service.ChargeServiceInterface
	ownershipService service.OwnershipServiceInterface
	authManager      *auth.AuthManager

	// UI елементи
	searchEntry   *widget.Entry
	filterSelect  *widget.Select
	chargesTable  *widget.Table
	createButton  *widget.Button
	refreshButton *widget.Button

	// Дані
	charges         []*charge.ChargeOutput
	filteredCharges []*charge.ChargeOutput
	updateStats     func()
	totalCount      int64
}

// NewChargesScreen створює новий екран нарахувань.
func NewChargesScreen(
	window fyne.Window,
	chargeService service.ChargeServiceInterface,
	ownershipService service.OwnershipServiceInterface,
	authManager *auth.AuthManager,
) *ChargesScreen {
	screen := &ChargesScreen{
		window:           window,
		chargeService:    chargeService,
		ownershipService: ownershipService,
		authManager:      authManager,
	}

	screen.buildUI()
	screen.loadCharges()

	return screen
}

// buildUI створює інтерфейс екрану.
func (s *ChargesScreen) buildUI() {
	// Пошук
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("🔍 Пошук за описом...")
	s.searchEntry.OnChanged = func(query string) {
		s.applyFilters()
	}

	// Фільтр
	s.filterSelect = widget.NewSelect([]string{
		"Всі нарахування",
		"Поточний місяць",
		"Минулий місяць",
		"Утримання",
		"Комунальні",
		"Ремонт",
		"Пеня",
		"Інше",
	}, func(value string) {
		s.applyFilters()
	})

	// Кнопка створення
	s.createButton = widget.NewButtonWithIcon("Додати нарахування", theme.ContentAddIcon(), func() {
		s.showCreateDialog()
	})
	s.createButton.Importance = widget.HighImportance

	// Кнопка оновлення
	s.refreshButton = widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadCharges()
	})

	// Таблиця
	s.buildTable()

	s.filterSelect.SetSelected("Всі нарахування")
}

// buildTable створює таблицю з нарахуваннями.
func (s *ChargesScreen) buildTable() {
	s.chargesTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredCharges) + 1, 7 // +1 для заголовків, 7 колонок
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				headers := []string{"№", "Дата", "Період", "Тип", "Сума", "Платник", "Дії"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			// Дані
			if id.Row-1 >= len(s.filteredCharges) {
				label.SetText("")
				return
			}

			c := s.filteredCharges[id.Row-1]

			switch id.Col {
			case 0: // №
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1: // Дата
				label.SetText(c.ChargeDate.Format("02.01.2006"))
			case 2: // Період
				label.SetText(fmt.Sprintf("%02d/%d", c.PeriodMonth, c.PeriodYear))
			case 3: // Тип
				label.SetText(c.TypeName)
			case 4: // Сума
				label.SetText(fmt.Sprintf("%.2f грн", c.Amount))
			case 5: // Платник
				// TODO: Тут краще було б мати ім'я власника або номер квартири в ChargeOutput
				// Поки що виводимо ID частки, але це треба виправити в майбутньому
				label.SetText(fmt.Sprintf("ID: %d", c.OwnershipShareID))
			case 6: // Дії
				label.SetText("⚙️")
			}
		},
	)

	// Ширина колонок
	s.chargesTable.SetColumnWidth(0, 50)  // №
	s.chargesTable.SetColumnWidth(1, 100) // Дата
	s.chargesTable.SetColumnWidth(2, 80)  // Період
	s.chargesTable.SetColumnWidth(3, 120) // Тип
	s.chargesTable.SetColumnWidth(4, 100) // Сума
	s.chargesTable.SetColumnWidth(5, 150) // Платник
	s.chargesTable.SetColumnWidth(6, 80)  // Дії

	// Клік на комірку
	s.chargesTable.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return // Заголовок
		}
		if id.Row-1 >= len(s.filteredCharges) {
			return
		}

		c := s.filteredCharges[id.Row-1]

		if id.Col == 6 {
			// Клік на "Дії"
			s.showActionsMenu(c)
		} else {
			// Клік на інші колонки - показати деталі
			s.showChargeDetails(c)
		}
	}
}

// Render повертає контейнер з UI екрану.
func (s *ChargesScreen) Render() fyne.CanvasObject {
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

	// Статистика
	statsLabel := widget.NewLabel(s.getStatsText())
	s.updateStats = func() {
		statsLabel.SetText(s.getStatsText())
	}

	// Головний контейнер
	content := container.NewBorder(
		toolbar,
		statsLabel,
		nil,
		nil,
		s.chargesTable,
	)

	return content
}

// getStatsText повертає текст статистики.
func (s *ChargesScreen) getStatsText() string {
	var totalAmount float64
	for _, c := range s.filteredCharges {
		totalAmount += c.Amount
	}
	return fmt.Sprintf("Показано: %d нарахувань на суму %.2f грн", len(s.filteredCharges), totalAmount)
}

// loadCharges завантажує список нарахувань.
func (s *ChargesScreen) loadCharges() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := s.chargeService.List(ctx, charge.ListChargesInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         1000,
		OrderBy:       "charge_date",
		OrderDesc:     true,
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка завантаження нарахувань: %v", err))
		return
	}

	s.charges = output.Charges
	s.totalCount = output.Total
	s.applyFilters()
}

// applyFilters застосовує фільтри до списку.
func (s *ChargesScreen) applyFilters() {
	s.filteredCharges = make([]*charge.ChargeOutput, 0)

	searchQuery := s.searchEntry.Text
	filterType := s.filterSelect.Selected
	now := time.Now()

	for _, c := range s.charges {
		// Пошук
		if searchQuery != "" {
			match := false
			if c.Description != nil && contains(*c.Description, searchQuery) {
				match = true
			}
			// Можна додати пошук по сумі або типу
			if !match {
				continue
			}
		}

		// Фільтр
		switch filterType {
		case "Поточний місяць":
			if c.PeriodMonth != int(now.Month()) || c.PeriodYear != now.Year() {
				continue
			}
		case "Минулий місяць":
			prevMonth := now.AddDate(0, -1, 0)
			if c.PeriodMonth != int(prevMonth.Month()) || c.PeriodYear != prevMonth.Year() {
				continue
			}
		case "Утримання":
			if c.ChargeType != string(entity.ChargeTypeMaintenance) {
				continue
			}
		case "Комунальні":
			if c.ChargeType != string(entity.ChargeTypeUtility) {
				continue
			}
		case "Ремонт":
			if c.ChargeType != string(entity.ChargeTypeRepair) {
				continue
			}
		case "Пеня":
			if c.ChargeType != string(entity.ChargeTypePenalty) {
				continue
			}
		case "Інше":
			if c.ChargeType != string(entity.ChargeTypeOther) {
				continue
			}
		}

		s.filteredCharges = append(s.filteredCharges, c)
	}

	s.chargesTable.Refresh()
	if s.updateStats != nil {
		s.updateStats()
	}
}

// showChargeDetails показує детальну інформацію про нарахування.
func (s *ChargesScreen) showChargeDetails(c *charge.ChargeOutput) {
	details := fmt.Sprintf(
		"ID: %d\n"+
			"Дата: %s\n"+
			"Період: %02d/%d\n"+
			"Тип: %s\n"+
			"Сума: %.2f грн\n"+
			"Платник (ID частки): %d\n",
		c.ID,
		c.ChargeDate.Format("02.01.2006"),
		c.PeriodMonth, c.PeriodYear,
		c.TypeName,
		c.Amount,
		c.OwnershipShareID,
	)

	if c.Description != nil {
		details += fmt.Sprintf("\nОпис: %s\n", *c.Description)
	}

	common.ShowInformation(s.window, "Деталі нарахування", details)
}

// showActionsMenu показує меню дій.
func (s *ChargesScreen) showActionsMenu(c *charge.ChargeOutput) {
	actions := []common.Action{
		{
			Label: "✏️ Редагувати",
			OnTap: func() { s.showEditDialog(c) },
		},
		{
			Label:      "🗑️ Видалити",
			OnTap:      func() { s.confirmDelete(c) },
			Importance: widget.DangerImportance,
		},
		{
			Label: "ℹ️ Деталі",
			OnTap: func() { s.showChargeDetails(c) },
		},
	}

	common.ShowActionsMenu(s.window, fmt.Sprintf("Нарахування #%d", c.ID), actions)
}

// showCreateDialog показує діалог створення.
func (s *ChargesScreen) showCreateDialog() {
	dialog := NewChargeFormDialog(s.window, s.chargeService, s.ownershipService, s.authManager, nil)
	dialog.OnSuccess = func() {
		s.loadCharges()
	}
	dialog.Show()
}

// showEditDialog показує діалог редагування.
func (s *ChargesScreen) showEditDialog(c *charge.ChargeOutput) {
	dialog := NewChargeFormDialog(s.window, s.chargeService, s.ownershipService, s.authManager, c)
	dialog.OnSuccess = func() {
		s.loadCharges()
	}
	dialog.Show()
}

// confirmDelete підтверджує видалення.
func (s *ChargesScreen) confirmDelete(c *charge.ChargeOutput) {
	common.ShowDeleteConfirmation(
		s.window,
		"Підтвердження видалення",
		fmt.Sprintf("Видалити нарахування #%d на суму %.2f грн?", c.ID, c.Amount),
		func() {
			s.deleteCharge(c)
		},
	)
}

// deleteCharge видаляє нарахування.
func (s *ChargesScreen) deleteCharge(c *charge.ChargeOutput) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.chargeService.Delete(ctx, charge.DeleteChargeInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		ChargeID:      c.ID,
	})

	if err != nil {
		common.ShowError(s.window, fmt.Errorf("Помилка видалення: %v", err))
		return
	}

	common.ShowSuccess(s.window, "Нарахування успішно видалено")
	s.loadCharges()
}
