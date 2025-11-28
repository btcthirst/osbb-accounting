// presentation/fyne/screens/import_screen.go
package screens

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/common"
)

// ImportScreen екран імпорту/експорту даних.
type ImportScreen struct {
	window        fyne.Window
	importService service.ImportService
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
func NewImportScreen(window fyne.Window, importService service.ImportService, userID int64) *ImportScreen {
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
	s.importButton = widget.NewButton("Імпортувати XLSX", s.handleImport)
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
				headers := []string{"ID", "Файл", "Статус", "Аркушів", "Записів", "Дата імпорту", "Дії"}
				renderTableHeader(label, headers, id.Col)
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
			// Виконати імпорт
			summary, err := s.importService.ImportXLSXFile(filePath, s.userID)

			if err != nil {
				fyne.Do(func() {
					progress.Hide()
					dialog.ShowError(err, s.window)
				})
				return
			}

			// Показати результат
			message := fmt.Sprintf(
				"Імпорт завершено!\n\n"+
					"Файл: %s\n"+
					"Аркушів: %d\n"+
					"Записів: %d\n"+
					"Успішно: %d\n"+
					"Помилок: %d\n"+
					"Час: %.2f сек",
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
	batches, err := s.importService.GetImportBatches()
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження: %v", err), s.window)
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
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf("Дії з батчем #%d", batch.ID), actions)
}

// showFullBatchDetails показує повні деталі батчу з можливістю прокрутки.
func (s *ImportScreen) showFullBatchDetails(batch *entity.ImportBatch) {
	// Завантажити записи
	records, err := s.importService.GetImportedRecords(batch.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження записів: %v", err), s.window)
		return
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
		tabs.Append(container.NewTabItem("Немає даних", widget.NewLabel("Немає записів для відображення")))
	}

	// Створити контейнер для діалогу
	content := container.NewBorder(
		widget.NewLabel(fmt.Sprintf("Батч #%d: %s (Всього: %d записів)", batch.ID, batch.FileName, len(records))),
		nil, nil, nil,
		tabs,
	)

	// Показати в діалозі
	d := dialog.NewCustom(
		"Деталі імпорту по місяцях",
		"Закрити",
		content,
		s.window,
	)
	d.Resize(fyne.NewSize(900, 700))
	d.Show()
}

// createMonthTable створює таблицю для конкретного місяця
func (s *ImportScreen) createMonthTable(records []*entity.ImportedMonthlyRecord) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(records) + 1, 6 // +1 для заголовка, 6 колонок
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			// Заголовки
			if id.Row == 0 {
				headers := []string{"Квартира", "ПІБ", "Період", "Нараховано", "Сплачено", "Борг"}
				renderTableHeader(label, headers, id.Col)
				return
			}

			// Дані
			if id.Row-1 < len(records) {
				record := records[id.Row-1]

				switch id.Col {
				case 0:
					label.SetText(record.ApartmentNumber)
				case 1:
					label.SetText(record.OwnerName)
				case 2:
					label.SetText(record.GetPeriodDisplay())
				case 3:
					label.SetText(fmt.Sprintf("%.2f", record.ChargeAmount))
				case 4:
					label.SetText(fmt.Sprintf("%.2f", record.AmountPaid))
				case 5:
					label.SetText(fmt.Sprintf("%.2f", record.AmountDue))
				}
			}
		},
	)

	// Ширина колонок
	setupTableColumnWidths(table, map[int]float32{
		0: 80,  // Квартира
		1: 200, // ПІБ
		2: 100, // Період
		3: 100, // Нараховано
		4: 100, // Сплачено
		5: 100, // Борг
	})

	return table
}

// confirmDeleteBatch підтверджує видалення батчу.
func (s *ImportScreen) confirmDeleteBatch(batch *entity.ImportBatch) {
	message := fmt.Sprintf(
		"Ви впевнені, що хочете видалити цей імпорт?\n\n"+
			"Батч #%d\n"+
			"Файл: %s\n"+
			"Записів: %d\n\n"+
			"Це дія незворотна!",
		batch.ID,
		batch.FileName,
		batch.TotalRecords,
	)

	dialog.ShowConfirm(
		"Підтвердження видалення",
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
	err := s.importService.DeleteImportBatch(batch.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка видалення: %v", err), s.window)
		return
	}

	dialog.ShowInformation(
		"Успішно видалено",
		fmt.Sprintf("Батч #%d успішно видалено", batch.ID),
		s.window,
	)

	// Оновити список
	s.handleRefresh()
}
