// application/usecase/cashflow/export_cashflow.go
package cashflow

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportToXLSXInput - параметри для експорту
type ExportToXLSXInput struct {
	Entries      []*CashFlowEntry
	Month        int
	Year         int
	OSBBName     string
	TotalDebit   float64
	TotalCredit  float64
	StartBalance float64
	EndBalance   float64
}

// ExportToXLSX генерує XLSX файл з даними руху коштів
func ExportToXLSX(input ExportToXLSXInput) (*excelize.File, error) {
	f := excelize.NewFile()
	sheetName := "Рух коштів"

	// Створюємо або отримуємо аркуш
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Видаляємо дефолтний аркуш

	// Заголовок
	monthNames := []string{
		"", "Січень", "Лютий", "Березень", "Квітень", "Травень", "Червень",
		"Липень", "Серпень", "Вересень", "Жовтень", "Листопад", "Грудень",
	}

	row := 1

	// Рядок 1: Назва ОСББ та дата
	osbbName := input.OSBBName
	if osbbName == "" {
		osbbName = "ОСББ"
	}
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%s - %s %d", osbbName, monthNames[input.Month], input.Year))
	row++

	// Рядок 2: Порожній
	row++

	// Рядок 3: Назва звіту
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Відомість обліку грошових коштів")
	row++

	// Рядок 4: Сальдо на початок
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Сальдо на початок періоду: %.2f грн.", input.StartBalance))
	row++

	// Рядок 5: Заголовки таблиці
	headers := []string{"№", "Дата", "Тип операції", "Контрагент", "Опис", "Дебет (надходження)", "Кредит (витрати)", "Баланс"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), row)
		f.SetCellValue(sheetName, cell, header)
	}

	// Стиль для заголовків
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF00"}, // Жовтий фон
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("H%d", row), headerStyle)
	row++

	// Стиль для даних
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Стиль для чисел
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 2, // 0.00
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Дані
	for i, entry := range input.Entries {
		// Номер
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)

		// Дата
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), entry.Date.Format("02.01.2006"))

		// Тип
		entryType := "Надходження"
		if entry.Type == "expense" {
			entryType = "Витрата"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), entryType)

		// Контрагент
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), entry.Counterparty)

		// Опис
		desc := entry.Description
		if entry.Category != "" {
			desc += fmt.Sprintf(" [%s]", entry.Category)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), desc)

		// Дебет
		if entry.Debit > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), entry.Debit)
			f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), numberStyle)
		}

		// Кредит
		if entry.Credit > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), entry.Credit)
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), numberStyle)
		}

		// Баланс
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), entry.Balance)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), numberStyle)

		// Стиль для тексту
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("E%d", row), dataStyle)

		row++
	}

	// Підсумкові рядки
	row++
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
	})

	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "ВСЬОГО:")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), input.TotalDebit)
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), input.TotalCredit)
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), input.EndBalance)
	f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("H%d", row), summaryStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("H%d", row), numberStyle)

	// Ширина колонок
	f.SetColWidth(sheetName, "A", "A", 5)
	f.SetColWidth(sheetName, "B", "B", 12)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 35)
	f.SetColWidth(sheetName, "F", "F", 18)
	f.SetColWidth(sheetName, "G", "G", 18)
	f.SetColWidth(sheetName, "H", "H", 15)

	return f, nil
}

// GenerateFilename створює ім'я файлу для експорту
func GenerateFilename(month, year int) string {
	monthNames := []string{
		"", "Січень", "Лютий", "Березень", "Квітень", "Травень", "Червень",
		"Липень", "Серпень", "Вересень", "Жовтень", "Листопад", "Грудень",
	}
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("Рух_коштів_%s_%d_%s.xlsx", monthNames[month], year, timestamp)
}
