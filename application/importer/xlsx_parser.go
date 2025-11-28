// application/importer/xlsx_parser.go
package importer

import (
	"fmt"
	"strconv"
	"strings"

	"osbb-accounting/domain/entity"

	"github.com/xuri/excelize/v2"
)

// XLSXParser парсить XLSX файл з нарахуваннями.
type XLSXParser struct{}

// ParsedData представляє розпарсені дані з файлу.
type ParsedData struct {
	Records      []*entity.ImportedMonthlyRecord
	TotalSheets  int
	SuccessCount int
	FailedCount  int
	Errors       []string
}

// NewXLSXParser створює новий парсер.
func NewXLSXParser() *XLSXParser {
	return &XLSXParser{}
}

// ParseFile парсить XLSX файл і повертає записи.
func (p *XLSXParser) ParseFile(filePath string, batchID int64) (*ParsedData, error) {
	// Відкрити файл
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Отримати список аркушів
	sheetList := f.GetSheetList()

	// Мапа назв місяців
	monthMap := map[string]int{
		"січень":   1,
		"лютий":    2,
		"березень": 3,
		"квітень":  4,
		"травень":  5,
		"червень":  6,
		"липень":   7,
		"серпень":  8,
		"вересень": 9,
		"жовтень":  10,
		"листопад": 11,
		"грудень":  12,
	}

	data := &ParsedData{
		Records: make([]*entity.ImportedMonthlyRecord, 0),
		Errors:  make([]string, 0),
	}

	// Парсити кожен аркуш
	for _, sheetName := range sheetList {
		// Пропустити "Лист1" та інші не-місячні аркуші
		month, ok := monthMap[strings.ToLower(sheetName)]
		if !ok {
			continue
		}

		data.TotalSheets++

		// Парсити аркуш
		sheetRecords, sheetErrs := p.parseSheet(f, sheetName, month, batchID)
		data.Records = append(data.Records, sheetRecords...)
		data.SuccessCount += len(sheetRecords)
		data.FailedCount += len(sheetErrs)

		if len(sheetErrs) > 0 {
			data.Errors = append(data.Errors, sheetErrs...)
		}
	}

	return data, nil
}

// parseSheet парсить один аркуш.
func (p *XLSXParser) parseSheet(f *excelize.File, sheetName string, month int, batchID int64) ([]*entity.ImportedMonthlyRecord, []string) {
	// Отримати всі рядки
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, []string{fmt.Sprintf("Sheet %s: failed to read rows: %v", sheetName, err)}
	}

	// Визначити рік з рядка 3 (у січні 2023 року)
	year := p.extractYear(rows)
	if year == 0 {
		year = 2023 // default
	}

	records := make([]*entity.ImportedMonthlyRecord, 0)
	errors := make([]string, 0)

	// Парсити рядки 7-221 (індекси 6-220)
	for i := 6; i < len(rows) && i < 221; i++ {
		row := rows[i]

		// Пропустити порожні рядки або підсумкові рядки
		if len(row) == 0 || p.isSummaryRow(row) {
			continue
		}

		// Парсити рядок
		record, err := p.parseRow(row, month, year, sheetName, batchID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Sheet %s, Row %d: %v", sheetName, i+1, err))
			continue
		}

		if record != nil {
			records = append(records, record)
		}
	}

	return records, errors
}

// parseRow парсить один рядок даних.
func (p *XLSXParser) parseRow(row []string, month, year int, sheetName string, batchID int64) (*entity.ImportedMonthlyRecord, error) {
	// Перевірка мінімальної довжини
	if len(row) < 10 {
		return nil, fmt.Errorf("row too short: %d columns", len(row))
	}

	// Col 0: № кв.
	apartmentNumber := strings.TrimSpace(row[0])
	if apartmentNumber == "" {
		return nil, nil // Пропустити пусті рядки
	}

	// Col 1: ПІБ
	ownerName := strings.TrimSpace(row[1])
	if ownerName == "" {
		return nil, fmt.Errorf("owner name is required")
	}

	// Створити запис
	record, err := entity.NewImportedMonthlyRecord(batchID, month, year, apartmentNumber, ownerName)
	if err != nil {
		return nil, err
	}

	// Встановити назву аркушу
	record.SheetName = &sheetName

	// Col 2: Особ. рах.
	if len(row) > 2 && row[2] != "" {
		accountNumber := strings.TrimSpace(row[2])
		record.AccountNumber = &accountNumber
	}

	// Col 3: Д-Т (opening debit)
	if len(row) > 3 {
		record.OpeningDebit = p.parseFloat(row[3])
	}

	// Col 4: К-Т (opening credit)
	if len(row) > 4 {
		record.OpeningCredit = p.parseFloat(row[4])
	}

	// Col 5: Пільга %
	if len(row) > 5 {
		record.DiscountPercent = p.parseInt(row[5])
	}

	// Col 6: Заг.площа
	if len(row) > 6 {
		area := p.parseFloat(row[6])
		if area > 0 {
			record.TotalArea = &area
		}
	}

	// Col 7: Пільг.Площа
	if len(row) > 7 {
		record.DiscountArea = p.parseFloat(row[7])
	}

	// Col 8: Внески (tariff)
	if len(row) > 8 {
		tariff := p.parseFloat(row[8])
		if tariff > 0 {
			record.Tariff = &tariff
		}
	}

	// Col 9: 100% нарах.
	if len(row) > 9 {
		record.ChargeAmount = p.parseFloat(row[9])
	}

	// Col 10: Пільгова сума
	if len(row) > 10 {
		record.DiscountAmount = p.parseFloat(row[10])
	}

	// Col 12: Коригування
	if len(row) > 12 && row[12] != "" {
		corrections := p.parseFloat(row[12])
		record.Corrections = &corrections
	}

	// Col 13: До сплати
	if len(row) > 13 {
		record.AmountDue = p.parseFloat(row[13])
	}

	// Col 14: Сплачено
	if len(row) > 14 {
		record.AmountPaid = p.parseFloat(row[14])
	}

	// Col 15: Д-Т (closing debit)
	if len(row) > 15 {
		record.ClosingDebit = p.parseFloat(row[15])
	}

	// Col 16: К-Т (closing credit)
	if len(row) > 16 {
		record.ClosingCredit = p.parseFloat(row[16])
	}

	return record, nil
}

// extractYear витягує рік з заголовка (рядок 3).
func (p *XLSXParser) extractYear(rows [][]string) int {
	if len(rows) < 3 {
		return 0
	}

	// Рядок 3: " у      січні    2023 року"
	headerRow := rows[2]
	if len(headerRow) == 0 {
		return 0
	}

	text := headerRow[0]
	words := strings.Fields(text)

	// Шукати 4-значне число
	for _, word := range words {
		if year, err := strconv.Atoi(word); err == nil {
			if year >= 2020 && year <= 2100 {
				return year
			}
		}
	}

	return 0
}

// isSummaryRow перевіряє чи це підсумковий рядок.
func (p *XLSXParser) isSummaryRow(row []string) bool {
	// Підсумкові рядки зазвичай мають тільки суми без № кв. та ПІБ
	if len(row) < 2 {
		return true
	}

	apartmentNumber := strings.TrimSpace(row[0])
	ownerName := strings.TrimSpace(row[1])

	return apartmentNumber == "" && ownerName == ""
}

// parseFloat парсить значення як float64.
func (p *XLSXParser) parseFloat(value string) float64 {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", ".")
	value = strings.ReplaceAll(value, " ", "")

	if value == "" {
		return 0
	}

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}

	return result
}

// parseInt парсить значення як int.
func (p *XLSXParser) parseInt(value string) int {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")

	if value == "" {
		return 0
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return result
}
