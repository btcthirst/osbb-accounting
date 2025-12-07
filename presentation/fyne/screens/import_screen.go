// presentation/fyne/screens/import_screen.go
package screens

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/text"
)

// ImportScreen екран імпорту/експорту даних.
type ImportScreen struct {
	window        fyne.Window
	importService service.ImportServiceInterface
	userID        int64

	// UI елементи
	importButton  *widget.Button
	refreshButton *widget.Button
	statsLabel    *widget.Label
	batchesTable  *widget.Table

	// Дані
	batches []*entity.ImportBatch
}

// NewImportScreen створює новий екран імпорту.
func NewImportScreen(window fyne.Window, importService service.ImportServiceInterface, userID int64) *ImportScreen {
	screen := &ImportScreen{
		window:        window,
		importService: importService,
		userID:        userID,
		batches:       make([]*entity.ImportBatch, 0),
	}

	screen.buildUI()
	screen.loadBatches()

	return screen
}

// buildUI створює інтерфейс екрану.
func (s *ImportScreen) buildUI() {
	// Кнопки
	s.importButton = widget.NewButton(text.ActionImportXlsx, s.handleImport)
	s.refreshButton = newRefreshButton(s.handleRefresh)
	s.statsLabel = widget.NewLabel("")

	// Таблиця
	s.batchesTable = s.createBatchesTable()
}

// buildTable створює таблицю з батчами імпорту.
func (s *ImportScreen) createBatchesTable() *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(s.batches) + 1, 7 // +1 для заголовка, 7 колонок (додано Дії)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			// Заголовки
			if id.Row == 0 {
				renderTableHeader(label, text.ImportBatchesTableHeaders, id.Col)
				return
			}

			// Дані
			if id.Row-1 < len(s.batches) {
				batch := s.batches[id.Row-1]

				switch id.Col {
				case 0:
					label.SetText(fmt.Sprintf("%d", batch.ID))
				case 1:
					label.SetText(batch.FileName)
				case 2:
					label.SetText(batch.Status.GetDisplayName())
				case 3:
					label.SetText(fmt.Sprintf("%d", batch.TotalSheets))
				case 4:
					label.SetText(fmt.Sprintf("%d/%d", batch.SuccessfulRecords, batch.TotalRecords))
				case 5:
					label.SetText(batch.ImportedAt.Format("02.01.2006 15:04"))
				case 6:
					renderActionColumn(label)
				}
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(table, map[int]float32{
		0: 50,  // ID
		1: 250, // Файл
		2: 120, // Статус
		3: 80,  // Аркушів
		4: 100, // Записів
		5: 150, // Дата
		6: 50,  // Дії
	})

	// Клік на комірку
	table.OnSelected = func(id widget.TableCellID) {
		defer table.Unselect(id)

		if id.Row == 0 {
			return // Заголовок
		}

		dataIndex := id.Row - 1
		if dataIndex >= len(s.batches) {
			return
		}

		batch := s.batches[dataIndex]

		if id.Col == 6 {
			// Клік на "Дії"
			s.showActionsMenu(batch)
		} else {
			// Клік на інші колонки - показати деталі
			s.showFullBatchDetails(batch)
		}
	}

	return table
}

// Render повертає контейнер з UI екрану.
func (s *ImportScreen) Render() fyne.CanvasObject {
	// Toolbar
	toolbar := container.NewHBox(
		s.importButton,
		s.refreshButton,
	)

	// Головний контейнер
	content := container.NewBorder(
		toolbar,
		s.statsLabel,
		nil,
		nil,
		s.batchesTable,
	)

	return content
}

// getStatsText повертає текст статистики.
func (s *ImportScreen) getStatsText() string {
	return fmt.Sprintf("Показано: %d імпортів", len(s.batches))
}

// updateStats оновлює статистику.
func (s *ImportScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}

// handleImport обробляє імпорт файлу.
func (s *ImportScreen) handleImport() {
	// Діалог вибору файлу
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()

		// Перевірити розширення
		if filepath.Ext(filePath) != ".xlsx" {
			dialog.ShowError(fmt.Errorf("підтримуються тільки .xlsx файли"), s.window)
			return
		}

		// Показати прогрес
		progress := dialog.NewCustomWithoutButtons("Імпорт", widget.NewProgressBarInfinite(), s.window)
		progress.Show()

		go func() {
			// Виконати імпорт з timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			summary, err := s.importService.ImportXLSXFile(ctx, filePath, s.userID)

			if err != nil {
				fyne.Do(func() {
					progress.Hide()
					dialog.ShowError(err, s.window)
				})
				return
			}

			// Показати результат
			message := fmt.Sprintf(
				text.MsgImportCompleted,
				summary.FileName,
				summary.TotalSheets,
				summary.TotalRecords,
				summary.SuccessfulRecords,
				summary.FailedRecords,
				summary.DurationSeconds,
			)

			fyne.Do(func() {
				progress.Hide()
				dialog.ShowInformation("Імпорт завершено", message, s.window)
				s.handleRefresh()
			})
		}()

	}, s.window)

	// Фільтр для .xlsx файлів
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
	fileDialog.Show()
}

// handleRefresh оновлює список батчів.
func (s *ImportScreen) handleRefresh() {
	s.loadBatches()
}

// loadBatches завантажує батчі з БД.
func (s *ImportScreen) loadBatches() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	batches, err := s.importService.GetImportBatches(ctx)
	if err != nil {
		dialog.ShowError(fmt.Errorf(text.MsgErrorBatchLoading, err), s.window)
		return
	}

	s.batches = batches
	s.batchesTable.Refresh()
	s.updateStats()
}

// showActionsMenu показує меню дій з батчем.
func (s *ImportScreen) showActionsMenu(batch *entity.ImportBatch) {
	actions := buildStandardActions(
		nil, // немає редагування
		func() { s.confirmDeleteBatch(batch) },
		func() { s.showFullBatchDetails(batch) },
		func() { s.processBatch(batch) },
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf(text.TitleBatchActions, batch.ID), actions)
}

// processBatch запускає обробку батчу.
func (s *ImportScreen) processBatch(batch *entity.ImportBatch) {
	progress := dialog.NewCustomWithoutButtons("Обробка даних", widget.NewProgressBarInfinite(), s.window)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		result, err := s.importService.ProcessBatch(ctx, batch.ID)

		fyne.Do(func() {
			progress.Hide()

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			// Формуємо повідомлення про результат
			msg := fmt.Sprintf("Обробку завершено!\n\nУспішно: %d\nПомилок: %d", result.SuccessCount, result.FailCount)
			if len(result.Errors) > 0 {
				msg += "\n\nПомилки:\n"
				// Показуємо перші 10 помилок
				limit := 10
				if len(result.Errors) < limit {
					limit = len(result.Errors)
				}
				for i := 0; i < limit; i++ {
					msg += fmt.Sprintf("- %s\n", result.Errors[i])
				}
				if len(result.Errors) > limit {
					msg += fmt.Sprintf("... та ще %d помилок", len(result.Errors)-limit)
				}
			}

			dialog.ShowInformation("Результат обробки", msg, s.window)
			s.handleRefresh()
		})
	}()
}

// showFullBatchDetails показує повні деталі батчу з можливістю прокрутки.
func (s *ImportScreen) showFullBatchDetails(batch *entity.ImportBatch) {
	// Завантажити записи
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Спробуємо спочатку стандартні записи
	records, err := s.importService.GetImportedRecords(ctx, batch.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження записів: %v", err), s.window)
		return
	}

	// Якщо стандартних записів немає, спробуємо Cash Flow записи
	if len(records) == 0 {
		cfRecords, err := s.importService.GetImportedCashFlowRecords(ctx, batch.ID)
		if err != nil {
			dialog.ShowError(fmt.Errorf("помилка завантаження записів cashflow: %v", err), s.window)
			return
		}

		if len(cfRecords) > 0 {
			s.showCashFlowBatchDetails(batch, cfRecords)
			return
		}
	}

	// Групування записів по місяцях
	recordsByMonth := make(map[int][]*entity.ImportedMonthlyRecord)
	for _, record := range records {
		recordsByMonth[record.PeriodMonth] = append(recordsByMonth[record.PeriodMonth], record)
	}

	// Створення табів
	tabs := container.NewAppTabs()

	// Ітерація від 1 до 12 (від січня до грудня)
	for month := 1; month <= 12; month++ {
		monthRecords, exists := recordsByMonth[month]
		if !exists || len(monthRecords) == 0 {
			continue
		}

		// Сортування записів за номером квартири (натуральне сортування)
		sort.Slice(monthRecords, func(i, j int) bool {
			n1 := monthRecords[i].ApartmentNumber
			n2 := monthRecords[j].ApartmentNumber

			// Спробуємо перетворити в числа
			i1, err1 := strconv.Atoi(n1)
			i2, err2 := strconv.Atoi(n2)

			if err1 == nil && err2 == nil {
				return i1 < i2
			}

			// Якщо не числа, або змішані - порівнюємо як рядки
			// Але спочатку за довжиною, щоб "2" було перед "10"
			if len(n1) != len(n2) {
				// Це працює тільки для числових рядків, для "1а" і "10" може бути не точно
				// Але для квартир зазвичай ок
				if err1 == nil || err2 == nil {
					// Якщо хоча б одне число - коротке число менше довгого
					return len(n1) < len(n2)
				}
			}

			return n1 < n2
		})

		// Створення таблиці для місяця
		table := s.createMonthTable(monthRecords)

		// Назва місяця
		monthName := common.GetMonthName(month)
		tabItem := container.NewTabItem(fmt.Sprintf("%s (%d)", monthName, len(monthRecords)), table)
		tabs.Append(tabItem)
	}

	// Якщо немає записів (що дивно), показати пустий контейнер
	if len(tabs.Items) == 0 {
		tabs.Append(container.NewTabItem(text.LabelNoData, widget.NewLabel(text.LabelNoRecords)))
	}

	// Створити контейнер для діалогу
	content := container.NewBorder(
		widget.NewLabel(fmt.Sprintf("Батч #%d: %s (Всього: %d записів)", batch.ID, batch.FileName, len(records))),
		nil, nil, nil,
		tabs,
	)

	// Показати в діалозі
	d := dialog.NewCustom(
		text.TitleImportDetails,
		text.ActionClose,
		content,
		s.window,
	)
	d.Resize(fyne.NewSize(900, 700))
	d.Show()
}

func (s *ImportScreen) showCashFlowBatchDetails(batch *entity.ImportBatch, records []*entity.ImportedCashFlowRecord) {
	// Group records by month
	recordsByMonth := make(map[int][]*entity.ImportedCashFlowRecord)
	for _, record := range records {
		month := int(record.Date.Month())
		recordsByMonth[month] = append(recordsByMonth[month], record)
	}

	// Create tabs
	tabs := container.NewAppTabs()

	// Iterate from 1 to 12
	for month := 1; month <= 12; month++ {
		monthRecords, exists := recordsByMonth[month]
		if !exists || len(monthRecords) == 0 {
			continue
		}

		// Sort by date
		sort.Slice(monthRecords, func(i, j int) bool {
			return monthRecords[i].Date.Before(monthRecords[j].Date)
		})

		// Create table for month
		table := s.createCashFlowMonthTable(monthRecords)

		// Month name
		monthName := common.GetMonthName(month)
		tabItem := container.NewTabItem(fmt.Sprintf("%s (%d)", monthName, len(monthRecords)), table)
		tabs.Append(tabItem)
	}

	if len(tabs.Items) == 0 {
		tabs.Append(container.NewTabItem(text.LabelNoData, widget.NewLabel(text.LabelNoRecords)))
	}

	content := container.NewBorder(
		widget.NewLabel(fmt.Sprintf("Cash Flow Батч #%d: %s (Всього: %d записів)", batch.ID, batch.FileName, len(records))),
		nil, nil, nil,
		tabs,
	)

	d := dialog.NewCustom(text.TitleImportDetails, text.ActionClose, content, s.window)
	d.Resize(fyne.NewSize(1000, 700))
	d.Show()
}

func (s *ImportScreen) createCashFlowMonthTable(records []*entity.ImportedCashFlowRecord) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(records) + 1, 12 // +1 header, 12 columns (matching CashFlowScreen)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			label.TextStyle = fyne.TextStyle{} // Reset style

			if id.Row == 0 {
				// Headers from CashFlowScreen: "№ п/п", "Контрагент", "Дата", "Дт рах. 311", "Оборот по дт", "313", "63", "641", "641.1", "651", "94", "Оборот по кт"
				renderTableHeader(label, text.CashFlowTableHeaders, id.Col)
				return
			}

			if id.Row-1 < len(records) {
				r := records[id.Row-1]

				// Helper to check category
				isCategory := func(code string) bool {
					return r.CategoryCode == code
				}

				switch id.Col {
				case 0: // № п/п
					label.SetText(fmt.Sprintf("%d", id.Row))
				case 1: // Контрагент
					label.SetText(r.ContractorName)
				case 2: // Дата
					label.SetText(r.Date.Format("02.01.2006"))
				case 3: // Дт рах. 311
					if r.OperationType == entity.CashFlowOperationDebit {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
						label.TextStyle = fyne.TextStyle{Bold: true}
					} else {
						label.SetText("")
					}
				case 4: // Оборот по дт
					if r.OperationType == entity.CashFlowOperationDebit {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
						label.TextStyle = fyne.TextStyle{Bold: true}
					} else {
						label.SetText("")
					}
				case 5: // 313
					if r.OperationType == entity.CashFlowOperationCredit && isCategory("313") {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
					} else {
						label.SetText("")
					}
				case 6: // 63
					if r.OperationType == entity.CashFlowOperationCredit && isCategory("63") {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
					} else {
						label.SetText("")
					}
				case 7: // 641
					if r.OperationType == entity.CashFlowOperationCredit && isCategory("641") {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
					} else {
						label.SetText("")
					}
				case 8: // 641.1
					if r.OperationType == entity.CashFlowOperationCredit && isCategory("641.1") {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
					} else {
						label.SetText("")
					}
				case 9: // 651
					if r.OperationType == entity.CashFlowOperationCredit && isCategory("651") {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
					} else {
						label.SetText("")
					}
				case 10: // 94
					if r.OperationType == entity.CashFlowOperationCredit && isCategory("94") {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
					} else {
						label.SetText("")
					}
				case 11: // Оборот по кт
					if r.OperationType == entity.CashFlowOperationCredit {
						label.SetText(fmt.Sprintf("%.2f", r.Amount))
						label.TextStyle = fyne.TextStyle{Bold: true}
					} else {
						label.SetText("")
					}
				}
			}
		},
	)

	// Column widths matching CashFlowScreen
	setupTableColumnWidths(table, map[int]float32{
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

	return table
}

// createMonthTable створює таблицю для конкретного місяця
func (s *ImportScreen) createMonthTable(records []*entity.ImportedMonthlyRecord) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(records) + 1, 16 // +1 для заголовка, 16 колонок
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			// Заголовки
			if id.Row == 0 {
				renderTableHeader(label, text.ImportRecordsTableHeaders, id.Col)
				return
			}

			// Дані
			if id.Row-1 < len(records) {
				record := records[id.Row-1]

				switch id.Col {
				case 0: // № кв.
					label.SetText(record.ApartmentNumber)
				case 1: // ПІБ
					label.SetText(record.OwnerName)
				case 2: // Особ. рах
					if record.AccountNumber != nil {
						label.SetText(*record.AccountNumber)
					} else {
						label.SetText("-")
					}
				case 3: // Д-Т (Поч)
					label.SetText(fmt.Sprintf("%.2f", record.OpeningDebit))
				case 4: // К-Т (Поч)
					label.SetText(fmt.Sprintf("%.2f", record.OpeningCredit))
				case 5: // Пільга %
					label.SetText(fmt.Sprintf("%d%%", record.DiscountPercent))
				case 6: // Заг. пл
					if record.TotalArea != nil {
						label.SetText(fmt.Sprintf("%.2f", *record.TotalArea))
					} else {
						label.SetText("-")
					}
				case 7: // Пільг. Площа
					label.SetText(fmt.Sprintf("%.2f", record.DiscountArea))
				case 8: // Тариф
					if record.Tariff != nil {
						label.SetText(fmt.Sprintf("%.2f", *record.Tariff))
					} else {
						label.SetText("-")
					}
				case 9: // 100% нарах.
					label.SetText(fmt.Sprintf("%.2f", record.ChargeAmount))
				case 10: // Пільгова сума
					label.SetText(fmt.Sprintf("%.2f", record.DiscountAmount))
				case 11: // Коригування
					if record.Corrections != nil {
						label.SetText(fmt.Sprintf("%.2f", *record.Corrections))
					} else {
						label.SetText("0.00")
					}
				case 12: // До сплати
					label.SetText(fmt.Sprintf("%.2f", record.AmountDue))
				case 13: // Сплачено
					label.SetText(fmt.Sprintf("%.2f", record.AmountPaid))
				case 14: // Д-Т (Кін)
					label.SetText(fmt.Sprintf("%.2f", record.ClosingDebit))
				case 15: // К-Т (Кін)
					label.SetText(fmt.Sprintf("%.2f", record.ClosingCredit))
				}
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(table, map[int]float32{
		0:  60,  // № кв.
		1:  200, // ПІБ
		2:  100, // Особ. рах
		3:  80,  // Д-Т (Поч)
		4:  80,  // К-Т (Поч)
		5:  60,  // Пільга %
		6:  70,  // Заг. пл
		7:  90,  // Пільг. Площа
		8:  60,  // Тариф
		9:  90,  // 100% нарах.
		10: 90,  // Пільгова сума
		11: 90,  // Коригування
		12: 90,  // До сплати
		13: 90,  // Сплачено
		14: 80,  // Д-Т (Кін)
		15: 80,  // К-Т (Кін)
	})

	return table
}

// confirmDeleteBatch підтверджує видалення батчу.
func (s *ImportScreen) confirmDeleteBatch(batch *entity.ImportBatch) {
	message := fmt.Sprintf(
		text.MsgConfirmDeleteBatch,
		batch.ID,
		batch.FileName,
		batch.TotalRecords,
	)

	dialog.ShowConfirm(
		text.MsgConfirmDeleteTitle,
		message,
		func(confirmed bool) {
			if confirmed {
				s.deleteBatch(batch)
			}
		},
		s.window,
	)
}

// deleteBatch видаляє батч.
func (s *ImportScreen) deleteBatch(batch *entity.ImportBatch) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.importService.DeleteImportBatch(ctx, batch.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка видалення: %v", err), s.window)
		return
	}

	dialog.ShowInformation(
		text.MsgSuccessTitle,
		fmt.Sprintf(text.MsgSuccessBatchDeleted, batch.ID),
		s.window,
	)

	// Оновити список
	s.handleRefresh()
}
