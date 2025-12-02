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

// ExportCashFlowUseCase - use case для експорту руху коштів
type ExportCashFlowUseCase struct{}

// NewExportCashFlowUseCase створює новий use case
func NewExportCashFlowUseCase() *ExportCashFlowUseCase {
	return &ExportCashFlowUseCase{}
}

// Execute генерує XLSX файл з даними руху коштів
func (uc *ExportCashFlowUseCase) Execute(input ExportToXLSXInput) (*excelize.File, error) {
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
	headers := []string{
		"№", "Контрагент", "Дата", "Дт рах. 311", "Оборот по дт",
		"313", "63", "641", "641.1", "651", "94", "Оборот по кт",
	}
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
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("L%d", row), headerStyle)
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
		// A: Номер
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)

		// B: Контрагент
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), entry.Counterparty)

		// C: Дата
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), entry.Date.Format("02.01.2006"))

		// D: Дт рах. 311 (надходження)
		if entry.Debit > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), entry.Debit)
			f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), numberStyle)
		} else if entry.Type == "expense" && entry.Credit > 0 {
			// Для витрат показуємо суму з мінусом або назву контрагента (як в UI)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), entry.Counterparty)
		}

		// E: Оборот по дт
		if entry.Debit > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), entry.Debit)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), numberStyle)
		}

		// F: 313
		if entry.Credit313 > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), entry.Credit313)
			f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), numberStyle)
		}

		// G: 63
		if entry.Credit63 > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), entry.Credit63)
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), numberStyle)
		}

		// H: 641
		if entry.Credit641 > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), entry.Credit641)
			f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), numberStyle)
		}

		// I: 641.1
		if entry.Credit6411 > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), entry.Credit6411)
			f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), numberStyle)
		}

		// J: 651
		if entry.Credit651 > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), entry.Credit651)
			f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), numberStyle)
		}

		// K: 94
		if entry.Credit94 > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), entry.Credit94)
			f.SetCellStyle(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), numberStyle)
		}

		// L: Оборот по кт
		totalCredit := entry.Credit313 + entry.Credit63 + entry.Credit641 + entry.Credit6411 + entry.Credit651 + entry.Credit94
		if totalCredit > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), totalCredit)
			f.SetCellStyle(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), numberStyle)
		} else if entry.Credit > 0 && totalCredit == 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), entry.Credit)
			f.SetCellStyle(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), numberStyle)
		}

		// Стиль для тексту (A-C)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		// Стиль для рамок інших клітинок
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("L%d", row), dataStyle)

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

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "ВСЬОГО:")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), input.TotalDebit)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), input.TotalDebit)
	f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), input.TotalCredit)

	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("L%d", row), summaryStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("L%d", row), numberStyle)

	// Ширина колонок
	// Ширина колонок
	f.SetColWidth(sheetName, "A", "A", 5)  // №
	f.SetColWidth(sheetName, "B", "B", 25) // Контрагент
	f.SetColWidth(sheetName, "C", "C", 12) // Дата
	f.SetColWidth(sheetName, "D", "D", 15) // Дт 311
	f.SetColWidth(sheetName, "E", "E", 15) // Оборот дт
	f.SetColWidth(sheetName, "F", "F", 12) // 313
	f.SetColWidth(sheetName, "G", "G", 12) // 63
	f.SetColWidth(sheetName, "H", "H", 12) // 641
	f.SetColWidth(sheetName, "I", "I", 12) // 641.1
	f.SetColWidth(sheetName, "J", "J", 12) // 651
	f.SetColWidth(sheetName, "K", "K", 12) // 94
	f.SetColWidth(sheetName, "L", "L", 15) // Оборот кт

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
