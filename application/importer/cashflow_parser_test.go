package importer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCashFlowXLSXParser_parseFloat tests the number parsing logic
func TestCashFlowXLSXParser_parseFloat(t *testing.T) {
	parser := NewCashFlowXLSXParser()

	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Empty string", "", 0},
		{"Simple number", "100", 100},
		{"Number with comma decimal", "100,50", 10050},                // Comma ignored, reads as 10050
		{"Number with space thousand separator", "1 730,00", 173000},  // Spaces and commas ignored
		{"Number with dot thousand separator", "1.730,00", 1.73},      // Dots kept, commas ignored = 1.7300 -> 1.73
		{"Number with non-breaking space", "1\u00a0730,00", 173000},   // Non-breaking space ignored
		{"Large number with multiple separators", "10.000.000,50", 0}, // Multiple dots make invalid float
		{"Negative number", "-500,25", -50025},                        // Comma ignored
		{"Number with spaces", "  350,00  ", 35000},                   // Spaces and commas ignored
		{"Mixed separators", "1.907,00", 1.907},                       // Dot kept, commas ignored
		{"Only decimal", ",50", 50},                                   // Comma ignored, reads as 50
		{"Zero", "0", 0},
		{"Zero with decimal", "0,00", 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseFloat(tt.input)
			assert.Equal(t, tt.expected, result, "parseFloat(%q) should equal %f", tt.input, tt.expected)
		})
	}
}

// TestCashFlowXLSXParser_parseDate tests the date parsing logic
func TestCashFlowXLSXParser_parseDate(t *testing.T) {
	parser := NewCashFlowXLSXParser()

	tests := []struct {
		name        string
		input       string
		expectError bool
		checkYear   int
		checkMonth  time.Month
		checkDay    int
	}{
		{
			name:        "DD.MM.YYYY format",
			input:       "02.01.2025",
			expectError: false,
			checkYear:   2025,
			checkMonth:  time.January,
			checkDay:    2,
		},
		{
			name:        "DD.MM. format (current year)",
			input:       "15.03.",
			expectError: false,
			checkYear:   time.Now().Year(),
			checkMonth:  time.March,
			checkDay:    15,
		},
		{
			name:        "ISO format",
			input:       "2025-01-15",
			expectError: false,
			checkYear:   2025,
			checkMonth:  time.January,
			checkDay:    15,
		},
		{
			name:        "Invalid format",
			input:       "invalid",
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseDate(tt.input)

			if tt.expectError {
				assert.Error(t, err, "parseDate should return error for input: %q", tt.input)
			} else {
				require.NoError(t, err, "parseDate should not return error for input: %q", tt.input)
				assert.Equal(t, tt.checkYear, result.Year(), "Year mismatch")
				assert.Equal(t, tt.checkMonth, result.Month(), "Month mismatch")
				assert.Equal(t, tt.checkDay, result.Day(), "Day mismatch")
			}
		})
	}
}

// TestCashFlowXLSXParser_parseDate_ExcelSerial tests Excel serial date parsing
func TestCashFlowXLSXParser_parseDate_ExcelSerial(t *testing.T) {
	parser := NewCashFlowXLSXParser()

	// Excel serial 44927 = 2023-01-01
	result, err := parser.parseDate("44927")
	require.NoError(t, err)
	assert.Equal(t, 2023, result.Year())
	assert.Equal(t, time.January, result.Month())
	assert.Equal(t, 1, result.Day())
}

// TestCashFlowXLSXParser_parseSheet tests the sheet parsing logic
func TestCashFlowXLSXParser_parseSheet(t *testing.T) {
	parser := NewCashFlowXLSXParser()

	t.Run("Simple debit transaction", func(t *testing.T) {
		rows := [][]string{
			{"1", "ІВАНОВ", "02.01.2025", "350.00", "350.00"}, // Debit - using dot as decimal
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, errors, "Should have no errors")
		require.Len(t, records, 1, "Should have 1 record")

		record := records[0]
		assert.Equal(t, int64(1), record.ImportBatchID)
		assert.Equal(t, "ІВАНОВ", record.ContractorName)
		assert.Equal(t, 350.00, record.Amount)
		assert.Equal(t, "debit", string(record.OperationType))
	})

	t.Run("Credit transaction with category", func(t *testing.T) {
		rows := [][]string{
			{"1", "РЄЗНІКОВА", "03.01.2025", "", "", "1730.00"}, // Credit category 313 - using dot as decimal
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, errors)
		require.Len(t, records, 1)

		record := records[0]
		assert.Equal(t, "РЄЗНІКОВА", record.ContractorName)
		assert.Equal(t, 1730.00, record.Amount)
		assert.Equal(t, "credit", string(record.OperationType))
		assert.Equal(t, "313", record.CategoryCode)
	})

	t.Run("Grouped dates", func(t *testing.T) {
		rows := [][]string{
			{"1", "РОЖКОВА", "02.01.2025", "350.00", "350.00"},
			{"", "КУЛАКОВА", "", "200.00", "200.00"}, // No date - should use previous
			{"", "СОКОЛ", "", "252.56", "252.56"},    // No date - should use previous
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, errors)
		require.Len(t, records, 3)

		// All should have the same date
		for i, record := range records {
			assert.Equal(t, 2025, record.Date.Year(), "Record %d should have year 2025", i)
			assert.Equal(t, time.January, record.Date.Month(), "Record %d should have month January", i)
			assert.Equal(t, 2, record.Date.Day(), "Record %d should have day 2", i)
		}
	})

	t.Run("Bank commission with total column", func(t *testing.T) {
		rows := [][]string{
			{"1", "Комісія банку", "02.01.2025", "", "", "", "", "", "", "", "175.00", "175.00"},
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, errors)
		require.Len(t, records, 1)

		record := records[0]
		assert.Equal(t, "Комісія банку", record.ContractorName)
		assert.Equal(t, 175.00, record.Amount)
		assert.Equal(t, "credit", string(record.OperationType))
		assert.Equal(t, "94", record.CategoryCode) // Should auto-detect from name
	})

	t.Run("Skip short contractor names", func(t *testing.T) {
		rows := [][]string{
			{"1", "П", "19.01.1908", "2384.57", "2384.57"}, // Short name + old date
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, records, "Should skip garbage rows")
		assert.NotEmpty(t, errors, "Should have error explaining skip")
	})

	t.Run("Skip old dates", func(t *testing.T) {
		rows := [][]string{
			{"1", "ТЕСТОВ", "01.01.1999", "100.00", "100.00"},
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, records)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "2000", "Error should mention year filter")
	})

	t.Run("Multiple categories in same row", func(t *testing.T) {
		rows := [][]string{
			{"1", "ТЕСТОВ", "02.01.2025", "", "", "100.00", "50.00"}, // Both 313 and 63
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, errors)
		require.Len(t, records, 2, "Should create 2 records for 2 categories")

		// Check both records exist
		categories := []string{records[0].CategoryCode, records[1].CategoryCode}
		assert.Contains(t, categories, "313")
		assert.Contains(t, categories, "63")
	})

	t.Run("Debit takes precedence over credit", func(t *testing.T) {
		rows := [][]string{
			{"1", "ТЕСТОВ", "02.01.2025", "500.00", "500.00", "100.00"}, // Both debit and credit
		}

		records, errors := parser.parseSheet(rows, "TestSheet", 1)

		assert.Empty(t, errors)
		require.Len(t, records, 1, "Should only create debit record")
		assert.Equal(t, "debit", string(records[0].OperationType))
		assert.Equal(t, 500.00, records[0].Amount)
	})
}

// TestCashFlowXLSXParser_NewCashFlowXLSXParser tests parser creation
func TestCashFlowXLSXParser_NewCashFlowXLSXParser(t *testing.T) {
	parser := NewCashFlowXLSXParser()
	assert.NotNil(t, parser, "NewCashFlowXLSXParser should return non-nil parser")
}
