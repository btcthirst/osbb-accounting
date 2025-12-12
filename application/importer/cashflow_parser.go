package importer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"osbb-accounting/domain/entity"

	"github.com/xuri/excelize/v2"
)

// CashFlowXLSXParser parses Cash Flow XLSX files.
type CashFlowXLSXParser struct{}

// ParsedCashFlowData represents parsed data from the file.
type ParsedCashFlowData struct {
	Records      []*entity.ImportedCashFlowRecord
	TotalSheets  int
	SuccessCount int
	FailedCount  int
	Errors       []string
}

// NewCashFlowXLSXParser creates a new parser.
func NewCashFlowXLSXParser() *CashFlowXLSXParser {
	return &CashFlowXLSXParser{}
}

// ParseFile parses the XLSX file and returns records.
func (p *CashFlowXLSXParser) ParseFile(filePath string, batchID int64) (*ParsedCashFlowData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	data := &ParsedCashFlowData{
		Records: make([]*entity.ImportedCashFlowRecord, 0),
		Errors:  make([]string, 0),
	}

	for _, sheetName := range sheetList {
		// Skip non-month sheets if any (simple heuristic: check if it has data)
		rows, err := f.GetRows(sheetName)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Sprintf("Sheet %s: failed to read rows: %v", sheetName, err))
			continue
		}

		if len(rows) < 5 {
			continue // Skip empty or header-only sheets
		}

		data.TotalSheets++
		sheetRecords, sheetErrs := p.parseSheet(rows, sheetName, batchID)
		data.Records = append(data.Records, sheetRecords...)
		data.SuccessCount += len(sheetRecords)
		data.FailedCount += len(sheetErrs)

		if len(sheetErrs) > 0 {
			data.Errors = append(data.Errors, sheetErrs...)
		}
	}

	return data, nil
}

func (p *CashFlowXLSXParser) parseSheet(rows [][]string, sheetName string, batchID int64) ([]*entity.ImportedCashFlowRecord, []string) {
	records := make([]*entity.ImportedCashFlowRecord, 0)
	errors := make([]string, 0)

	// Assume header is first few rows. Data starts when we see a number in column A (index 0)
	// or a date in column C (index 2).

	var lastValidDate time.Time

	for i, row := range rows {
		rowIndex := i + 1 // 1-based index for reporting
		if len(row) < 4 {
			continue
		}

		// Check if it's a data row
		// We expect:
		// Column A (0): No (Optional, Integer)
		// Column B (1): Contractor (String)
		// Column C (2): Date (Date)

		// Try to parse Date in Column C (index 2)
		var date time.Time
		var err error

		if len(row) >= 3 && row[2] != "" {
			dateStr := row[2]
			date, err = p.parseDate(dateStr)
			if err == nil {
				lastValidDate = date
			}
		} else if !lastValidDate.IsZero() {
			// If date is missing, use the last valid date (grouping logic)
			date = lastValidDate
		} else {
			// No date and no previous date - skip (likely header)
			continue
		}

		// Check if Contractor is present
		if len(row) < 2 || row[1] == "" { // Ensure column B exists and is not empty
			// Date is present but no contractor? Might be summary.
			// For now, let's skip if contractor is empty
			continue
		}

		contractorName := strings.TrimSpace(row[1])
		if len(contractorName) < 2 {
			// Skip likely garbage or header rows (e.g. "П")
			continue
		}

		// Check date validity (skip old dates like 1908)
		if date.Year() < 2000 {
			errors = append(errors, fmt.Sprintf("Sheet %s, Row %d: Skipped (Date %s < 2000)", sheetName, rowIndex+1, date.Format("02.01.2006")))
			continue
		}

		// 1. Check Debit (Income) - Column D (3)
		debitRaw := ""
		if len(row) > 3 {
			debitRaw = row[3]
		}
		debitAmount := p.parseFloat(debitRaw)

		if debitAmount > 0 {
			record := &entity.ImportedCashFlowRecord{
				ImportBatchID:  batchID,
				Date:           date,
				ContractorName: contractorName,
				OperationType:  entity.CashFlowOperationDebit,
				Amount:         debitAmount,
				Description:    fmt.Sprintf("Надходження від %s", contractorName),
				CreatedAt:      time.Now(),
			}
			records = append(records, record)

			// If it's a Debit row, we assume it cannot be a Credit row as well.
			// This prevents duplicates if columns are shifted or misinterpreted.
			continue
		}

		// 2. Check Credit (Expense) - Columns F-K (5-10)
		// F: 313 (5)
		// G: 63 (6)
		// H: 641 (7)
		// I: 641.1 (8)
		// J: 651 (9)
		// K: 94 (10)

		creditColumns := map[int]string{
			5:  "313",
			6:  "63",
			7:  "641",
			8:  "641.1",
			9:  "651",
			10: "94",
		}

		foundCredit := false
		for colIdx, categoryCode := range creditColumns {
			if len(row) > colIdx {
				rawVal := row[colIdx]
				amount := p.parseFloat(rawVal)
				if amount > 0 {
					record := &entity.ImportedCashFlowRecord{
						ImportBatchID:  batchID,
						Date:           date,
						ContractorName: contractorName,
						OperationType:  entity.CashFlowOperationCredit,
						Amount:         amount,
						CategoryCode:   categoryCode,
						Description:    fmt.Sprintf("Витрата: %s (%s)", contractorName, categoryCode),
						CreatedAt:      time.Now(),
					}
					records = append(records, record)
					foundCredit = true

					// Assume only one credit category per row?
					// Usually yes, but let's allow multiple just in case,
					// unless we want to be very strict.
					// For now, let's keep allowing multiple credits if they exist,
					// as the "duplicate" issue was about Debit vs Credit.
				}
			}
		}

		if !foundCredit {
			// Try checking "Total Credit" column (Index 11)
			// Sometimes data is only in the total column (e.g. Bank Commission)
			if len(row) > 11 {
				amount := p.parseFloat(row[11])
				if amount > 0 {
					// Determine category
					categoryCode := "other"
					if strings.Contains(strings.ToLower(contractorName), "комісія") {
						categoryCode = "94"
					}

					record := &entity.ImportedCashFlowRecord{
						ImportBatchID:  batchID,
						Date:           date,
						ContractorName: contractorName,
						OperationType:  entity.CashFlowOperationCredit,
						Amount:         amount,
						CategoryCode:   categoryCode,
						Description:    fmt.Sprintf("Витрата: %s (Загальна)", contractorName),
						CreatedAt:      time.Now(),
					}
					records = append(records, record)
					foundCredit = true
				}
			}
		}

		if !foundCredit {
			// Log why it was skipped
			errors = append(errors, fmt.Sprintf("Sheet %s, Row %d: Skipped '%s' (DebitRaw='%s'->%.2f, No Credit found in cols 5-11)", sheetName, rowIndex, contractorName, debitRaw, debitAmount))
		}
	}

	return records, errors
}

func (p *CashFlowXLSXParser) parseDate(value string) (time.Time, error) {
	// Excel dates are often stored as days since 1900-01-01
	// But here it seems to be a string date.
	// Let's try standard formats.

	// If it's a float string (Excel serial date)
	if serial, err := strconv.ParseFloat(value, 64); err == nil {
		return excelize.ExcelDateToTime(serial, false)
	}

	// Try DD.MM. format (e.g. 02.01.) - assume current year
	if len(value) == 6 && value[5] == '.' {
		// Append current year
		currentYear := time.Now().Year()
		valueWithYear := fmt.Sprintf("%s%d", value, currentYear)
		if t, err := time.Parse("02.01.2006", valueWithYear); err == nil {
			return t, nil
		}
	}

	formats := []string{"02.01.2006", "2006-01-02", "01-02-06", "02.01.06"}
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

func (p *CashFlowXLSXParser) parseFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	// In this specific file format, it seems:
	// - Comma (.) is the decimal separator.
	// - Dot (,) is the thousand separator (or garbage).
	// - Spaces/Non-breaking spaces are thousand separators.

	// Strategy:
	// 1. Keep only digits, dot, minus. (Ignore commas and spaces)

	var sb strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		} else if r == '-' || r == '.' {
			sb.WriteRune(r)
		}
		// Ignore ',' and everything else
	}
	cleaned := sb.String()

	if cleaned == "" {
		return 0
	}

	result, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return result
}
