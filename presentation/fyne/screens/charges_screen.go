package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/charge"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/dialogs"
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
	statsLabel    *widget.Label

	// Дані
	charges         []*charge.ChargeOutput
	filteredCharges []*charge.ChargeOutput
	currentPage     int
	pageSize        int
	totalCount      int64
}

type ChargeFilterType string

const (
	ChargeFilterAll         ChargeFilterType = "Всі нарахування"
	ChargeFilterCurrent     ChargeFilterType = "Поточний місяць"
	ChargeFilterPrevious    ChargeFilterType = "Минулий місяць"
	ChargeFilterMaintenance ChargeFilterType = "Утримання"
	ChargeFilterUtility     ChargeFilterType = "Комунальні"
	ChargeFilterRepair      ChargeFilterType = "Ремонт"
	ChargeFilterPenalty     ChargeFilterType = "Пеня"
	ChargeFilterOther       ChargeFilterType = "Інше"
)

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
	s.searchEntry = newSearchEntry("🔍 Пошук за описом...", func(query string) {
		s.applyFilters()
	})

	// Фільтр
	// Фільтр
	s.filterSelect = newFilterSelect([]string{
		string(ChargeFilterAll),
		string(ChargeFilterCurrent),
		string(ChargeFilterPrevious),
		string(ChargeFilterMaintenance),
		string(ChargeFilterUtility),
		string(ChargeFilterRepair),
		string(ChargeFilterPenalty),
		string(ChargeFilterOther),
	}, func(value string) {
		s.applyFilters()
	})

	// Кнопка створення
	s.createButton = newCreateButton("Додати нарахування", func() {
		s.showChargeDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadCharges()
	})

	// Таблиця
	s.buildTable()

	s.statsLabel = widget.NewLabel("")
	s.filterSelect.SetSelected(string(ChargeFilterAll))
	s.loadCharges()
}

// buildTable створює таблицю з нарахуваннями.
func (s *ChargesScreen) buildTable() {
	s.chargesTable = widget.NewTable(
		func() (int, int) {
			return len(s.filteredCharges) + 1, 7 // +1 для заголовків, 7 колонок
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Заголовки
				renderTableHeader(label,
					[]string{"№", "Дата", "Період", "Тип", "Сума", "Платник", "Дії"}, id.Col)
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
				renderActionColumn(label)
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(s.chargesTable, map[int]float32{
		0: 50,  // №
		1: 100, // Дата
		2: 80,  // Період
		3: 120, // Тип
		4: 100, // Сума
		5: 150, // Платник
		6: 80,  // Дії
	})

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

	// Головний контейнер
	content := container.NewBorder(
		toolbar,
		s.statsLabel,
		nil,
		nil,
		s.chargesTable,
	)

	return content
}

// getStatsText повертає текст статистики.
func (s *ChargesScreen) getStatsText() string {
	totalAmount := 0.0
	for _, c := range s.filteredCharges {
		totalAmount += c.Amount
	}
	return fmt.Sprintf("Показано: %d з %d нарахувань | Загальна сума: %.2f грн", len(s.filteredCharges), s.totalCount, totalAmount)
}

// updateStats оновлює статистику.
func (s *ChargesScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
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

	s.filteredCharges = FilterList(
		s.charges,
		searchQuery,
		filterType,
		func(c *charge.ChargeOutput, query string) bool {
			return c.Description != nil && contains(*c.Description, query)
		},
		func(c *charge.ChargeOutput, filter string) bool {
			switch ChargeFilterType(filter) {
			case ChargeFilterCurrent:
				return c.PeriodMonth == int(now.Month()) && c.PeriodYear == now.Year()
			case ChargeFilterPrevious:
				prevMonth := now.AddDate(0, -1, 0)
				return c.PeriodMonth == int(prevMonth.Month()) && c.PeriodYear == prevMonth.Year()
			case ChargeFilterMaintenance:
				return c.ChargeType == string(entity.ChargeTypeMaintenance)
			case ChargeFilterUtility:
				return c.ChargeType == string(entity.ChargeTypeUtility)
			case ChargeFilterRepair:
				return c.ChargeType == string(entity.ChargeTypeRepair)
			case ChargeFilterPenalty:
				return c.ChargeType == string(entity.ChargeTypePenalty)
			case ChargeFilterOther:
				return c.ChargeType == string(entity.ChargeTypeOther)
			}
			return true
		},
	)

	s.chargesTable.Refresh()
	s.updateStats()
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
	actions := buildStandardActions(
		func() { s.showChargeDialog(c) },
		func() { s.confirmDelete(c) },
		func() { s.showChargeDetails(c) },
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf("Дії з нарахуванням #%d", c.ID), actions)
}

// showChargeDialog показує діалог створення/редагування.
func (s *ChargesScreen) showChargeDialog(existing *charge.ChargeOutput) {
	d := dialogs.NewChargeFormDialog(
		s.window,
		s.chargeService,
		s.ownershipService,
		s.authManager,
		existing,
	)
	d.OnSuccess = func() {
		s.loadCharges()
	}
	d.Show()
}

// confirmDelete підтверджує видалення.
func (s *ChargesScreen) confirmDelete(c *charge.ChargeOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf("нарахування #%d", c.ID), func() {
		s.deleteCharge(c)
	})
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
