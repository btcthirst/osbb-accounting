package screens

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/presentation/fyne/common"
)

func (s *CashFlowScreen) importCashFlow() {
	// Діалог вибору файлу
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			common.ShowError(s.window, err)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()

		// Перевірити розширення
		if filepath.Ext(filePath) != ".xlsx" {
			common.ShowError(s.window, fmt.Errorf("підтримуються тільки .xlsx файли"))
			return
		}

		// Показати прогрес
		progress := dialog.NewCustomWithoutButtons("Імпорт", widget.NewProgressBarInfinite(), s.window)
		progress.Show()

		go func() {
			// Виконати імпорт з timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			userID := s.authManager.GetCurrentUserID()
			summary, err := s.importService.ImportCashFlowFile(ctx, filePath, userID)

			if err != nil {
				fyne.Do(func() {
					progress.Hide()
					common.ShowError(s.window, err)
				})
				return
			}

			fyne.Do(func() {
				progress.Hide()

				// Показати результат
				message := fmt.Sprintf(
					"Файл імпортовано успішно!\n\nФайл: %s\nЗаписів: %d\n\nТепер ви можете обробити цей батч на екрані Імпорту.",
					summary.FileName,
					summary.TotalRecords,
				)

				common.ShowSuccess(s.window, message)
				s.loadCashFlow() // Оновити таблицю (хоча вона не зміниться поки не обробимо)
			})
		}()

	}, s.window)

	// Фільтр для .xlsx файлів
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
	fileDialog.Show()
}
