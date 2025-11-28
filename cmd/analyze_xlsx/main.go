package main

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func main() {
	filePath := "workdata/нарахування23.xlsx"

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	// Get all sheets
	sheets := f.GetSheetList()
	fmt.Println("=== XLSX File Analysis ===")
	fmt.Printf("Total sheets: %d\n", len(sheets))
	fmt.Printf("Sheets: %v\n\n", sheets)

	// Analyze the first sheet in detail
	firstSheet := sheets[0]
	fmt.Printf("=== Detailed Analysis of Sheet: %s ===\n", firstSheet)

	// Get all rows
	rows, err := f.GetRows(firstSheet)
	if err != nil {
		log.Fatalf("Error reading sheet %s: %v", firstSheet, err)
	}

	fmt.Printf("Total rows: %d\n\n", len(rows))

	// Print header rows (1-6)
	fmt.Println("--- Header Rows (1-6) ---")
	for i := 0; i < 6 && i < len(rows); i++ {
		fmt.Printf("Row %d: %v\n", i+1, rows[i])
	}

	// Print a few data rows (7-10)
	fmt.Println("\n--- Sample Data Rows (7-10) ---")
	for i := 6; i < 10 && i < len(rows); i++ {
		fmt.Printf("Row %d: ", i+1)
		if len(rows[i]) > 0 {
			for colIdx, cell := range rows[i] {
				fmt.Printf("Col%d=\"%v\" | ", colIdx, cell)
			}
		}
		fmt.Println()
	}

	// Column structure from row 6 (headers)
	fmt.Println("\n--- Column Headers (Row 6) ---")
	if len(rows) >= 6 {
		headers := rows[5] // 0-indexed, so row 6 is index 5
		for i, header := range headers {
			fmt.Printf("Column %d: \"%s\"\n", i, header)
		}
	}

	// Print summary rows (last 2 rows)
	fmt.Println("\n--- Summary Rows (last 2) ---")
	if len(rows) >= 2 {
		for i := len(rows) - 2; i < len(rows); i++ {
			fmt.Printf("Row %d: %v\n", i+1, rows[i])
		}
	}

	fmt.Printf("\n=== Statistics ===\n")
	fmt.Printf("Data rows (7-221): %d rows\n", len(rows)-8) // minus 6 header + 2 summary
	fmt.Printf("Expected apartment records: ~%d\n", len(rows)-8)
}
