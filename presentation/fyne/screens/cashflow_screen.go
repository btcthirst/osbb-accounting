// presentation/fyne/screens/cashflow_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/cashflow"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/text"
)

// CashFlowScreen - екран руху коштів (read-only звіт)
type CashFlowScreen struct {
	window          fyne.Window
	cashflowService *service.CashFlowService
	authManager     interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	}

	// UI
	refreshButton *widget.Button
	exportButton  *widget.Button
	table         *widget.Table
	monthSelect   *widget.Select
	yearSelect    *widget.Select
	statsLabel    *widget.Label

	// Data
	entries      []*cashflow.CashFlowEntry
	currentMonth int
	currentYear  int
}

// NewCashFlowScreen створює новий екран
func NewCashFlowScreen(
	window fyne.Window,
	cashflowService *service.CashFlowService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *CashFlowScreen {
	now := time.Now()
	s := &CashFlowScreen{
		window:          window,
		cashflowService: cashflowService,
		authManager:     authManager,
		currentMonth:    int(now.Month()),
		currentYear:     now.Year(),
	}
	s.buildUI()
	return s
}

func (s *CashFlowScreen) buildUI() {
	// Buttons
	s.refreshButton = newRefreshButton(func() {
		s.loadCashFlow()
	})

	s.exportButton = widget.NewButton(text.ActionExportXlsx, func() {
		s.exportToXLSX()
	})

	// Month/Year filters "Січень", "Лютий", "Березень", "Квітень", "Травень", "Червень", "Липень", "Серпень", "Вересень", "Жовтень", "Листопад", "Грудень",
	months := text.Months
	s.monthSelect = widget.NewSelect(months, func(selected string) {
		// Знаходимо індекс місяця
		for i, m := range months {
			if m == selected {
				s.currentMonth = i + 1
				break
			}
		}
		s.loadCashFlow()
	})
	// Don't set selected yet - will do after table is created

	// Years from 2020 to current +1
	years := []string{}
	for y := 2020; y <= time.Now().Year()+1; y++ {
		years = append(years, fmt.Sprintf("%d", y))
	}
	s.yearSelect = widget.NewSelect(years, func(selected string) {
		fmt.Sscanf(selected, "%d", &s.currentYear)
		s.loadCashFlow()
	})
	// Don't set selected yet - will do after table is created

	// Table
	s.createTable()

	// Stats
	s.statsLabel = widget.NewLabel("")

	// Now set selected values (after table is created) - this will trigger initial load
	s.monthSelect.SetSelected(months[s.currentMonth-1])
	s.yearSelect.SetSelected(fmt.Sprintf("%d", s.currentYear))
}

func (s *CashFlowScreen) createTable() {
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.entries) + 1, len(text.CashFlowTableHeaders) // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers "№ п/п", "Контрагент", "Дата", "Дт рах. 311", "Оборот по дт", "313", "63", "641", "641.1", "651", "94", "Оборот по кт"
				renderTableHeader(label, text.CashFlowTableHeaders, id.Col)
				return
			}

			// Data
			if id.Row-1 >= len(s.entries) {
				label.SetText("")
				return
			}
			entry := s.entries[id.Row-1]

			// Reset style
			label.TextStyle = fyne.TextStyle{}

			switch id.Col {
			case 0: // № п/п
				label.SetText(fmt.Sprintf("%d", id.Row))
			case 1: // Контрагент
				label.SetText(entry.Counterparty)
			case 2: // Дата
				label.SetText(entry.Date.Format("02.01.2006"))
			case 3: // Дт рах. 311 (надходження)
				if entry.Debit > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Debit))
					label.TextStyle = fyne.TextStyle{Bold: true}
				} else {
					label.SetText("")
				}
			case 4: // Оборот по дт (дублює колонку 3)
				if entry.Debit > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Debit))
					label.TextStyle = fyne.TextStyle{Bold: true}
				} else {
					label.SetText("")
				}
			case 5: // 313 - Кошти на картку
				if entry.Credit313 > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Credit313))
				} else {
					label.SetText("")
				}
			case 6: // 63 - Електроенергія
				if entry.Credit63 > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Credit63))
				} else {
					label.SetText("")
				}
			case 7: // 641 - ПДФО 18%
				if entry.Credit641 > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Credit641))
				} else {
					label.SetText("")
				}
			case 8: // 641.1 - Військовий збір
				if entry.Credit6411 > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Credit6411))
				} else {
					label.SetText("")
				}
			case 9: // 651 - ЄСВ 22%
				if entry.Credit651 > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Credit651))
				} else {
					label.SetText("")
				}
			case 10: // 94 - Комісія банку
				if entry.Credit94 > 0 {
					label.SetText(fmt.Sprintf("%.2f", entry.Credit94))
				} else {
					label.SetText("")
				}
			case 11: // Оборот по кт (сума колонок 6-11)
				totalCredit := entry.Credit313 + entry.Credit63 + entry.Credit641 + entry.Credit6411 + entry.Credit651 + entry.Credit94
				if totalCredit > 0 {
					label.SetText(fmt.Sprintf("%.2f", totalCredit))
					label.TextStyle = fyne.TextStyle{Bold: true}
				} else if entry.Credit > 0 && totalCredit == 0 {
					// Якщо є витрата, але вона не в категоріях - показуємо загальну суму
					label.SetText(fmt.Sprintf("%.2f", entry.Credit))
				} else {
					label.SetText("")
				}
			}
		},
	)

	// Column widths
	setupTableColumnWidths(s.table, map[int]float32{
		0:  60,  // № п/п
		1:  150, // Контрагент
		2:  100, // Дата
		3:  100, // Дт рах. 311
		4:  100, // Оборот по дт
		5:  90,  // 313
		6:  90,  // 63
		7:  90,  // 641
		8:  90,  // 641.1
		9:  90,  // 651
		10: 90,  // 94
		11: 100, // Оборот по кт
	})
}

func (s *CashFlowScreen) Render() fyne.CanvasObject {
	// Toolbar with filters
	filterBar := container.NewBorder(
		nil, nil,
		widget.NewLabel(text.LabelPeriod),
		nil,
		container.NewHBox(
			s.monthSelect,
			s.yearSelect,
		),
	)

	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(s.refreshButton, s.exportButton),
		nil,
		filterBar,
	)

	return buildStandardLayout(toolbar, s.table, s.statsLabel)
}

func (s *CashFlowScreen) loadCashFlow() {
	// Safety check
	if s.cashflowService == nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorServiceNotInitialized, "CashFlow"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	input := cashflow.ListCashFlowInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Month:         &s.currentMonth,
		Year:          &s.currentYear,
	}

	output, err := s.cashflowService.List(ctx, input)
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorLoading, err))
		return
	}

	s.entries = output.Entries
	s.table.Refresh()
	s.updateStats(output)
}

func (s *CashFlowScreen) updateStats(output *cashflow.ListCashFlowOutput) {
	stats := fmt.Sprintf(
		text.MsgCashFlowStats,
		output.Total,
		output.TotalDebit,
		output.TotalCredit,
		output.EndBalance,
	)
	s.statsLabel.SetText(stats)
}

func (s *CashFlowScreen) exportToXLSX() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Отримуємо OSBB назву
	osbbName := "ОСББ" // TODO: отримувати з налаштувань

	input := cashflow.ListCashFlowInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Month:         &s.currentMonth,
		Year:          &s.currentYear,
	}

	// Генеруємо файл
	file, filename, err := s.cashflowService.ExportToXLSX(ctx, input, osbbName)
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorExport, err))
		return
	}

	// Зберігаємо файл
	if err := file.SaveAs(filename); err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorSaveFile, err))
		return
	}

	common.ShowSuccess(s.window, fmt.Sprintf(text.MsgSuccessFileSaved, filename))
}
