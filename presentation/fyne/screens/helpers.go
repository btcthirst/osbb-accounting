package screens

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/presentation/fyne/common"
)

// ptrToString safely dereferences a string pointer
func ptrToString(ptr *string, defaultVal string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return defaultVal
}

// ptrIntToString safely dereferences an int pointer
func ptrIntToString(ptr *int, defaultVal string) string {
	if ptr != nil {
		return fmt.Sprintf("%v", *ptr)
	}
	return defaultVal
}

// ptrFloatToString safely dereferences a float64 pointer
func ptrFloatToString(ptr *float64, defaultVal string) string {
	if ptr != nil {
		return fmt.Sprintf("%.1f", *ptr)
	}
	return defaultVal
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
			strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}

// activeStatus returns a formatted string for active/inactive status
func activeStatus(isActive bool) string {
	if isActive {
		return "✅ Активний"
	}
	return "❌ Неактивний"
}

// ptrTimeToString safely dereferences a time pointer
func ptrTimeToString(ptr *time.Time, format, defaultVal string) string {
	if ptr != nil {
		return ptr.Format(format)
	}
	return defaultVal
}

// ============================================================================
// Table Helper Functions
// ============================================================================

// newTableTemplate створює базовий template для комірки таблиці
func newTableTemplate() *widget.Label {
	label := widget.NewLabel("Template")
	label.Truncation = fyne.TextTruncateEllipsis
	return label
}

// renderTableHeader рендерить заголовок таблиці
func renderTableHeader(label *widget.Label, headers []string, colIndex int) {
	if colIndex < len(headers) {
		label.SetText(headers[colIndex])
		label.TextStyle = fyne.TextStyle{Bold: true}
	}
}

// setupTableColumnWidths встановлює ширину колонок таблиці
func setupTableColumnWidths(table *widget.Table, widths map[int]float32) {
	for col, width := range widths {
		table.SetColumnWidth(col, width)
	}
}

// renderActionColumn рендерить колонку дій (іконка налаштувань)
func renderActionColumn(label *widget.Label) {
	label.SetText("⚙️")
}

// ============================================================================
// Layout Helper Functions
// ============================================================================

// buildStandardToolbar створює стандартний toolbar з кнопками та фільтрами
func buildStandardToolbar(createButton, refreshButton *widget.Button, filters ...fyne.CanvasObject) *fyne.Container {
	return container.NewBorder(
		nil, nil,
		container.NewHBox(createButton, refreshButton),
		nil,
		container.NewVBox(filters...),
	)
}

// buildStandardLayout створює стандартний layout екрану зі списком
func buildStandardLayout(toolbar *fyne.Container, table *widget.Table, statsLabel *widget.Label) *fyne.Container {
	return container.NewBorder(
		toolbar,    // top
		statsLabel, // bottom
		nil, nil,   // left, right
		table, // center
	)
}

// ============================================================================
// Dialog Helper Functions
// ============================================================================

// showConfirmDeleteDialog показує діалог підтвердження видалення
func showConfirmDeleteDialog(window fyne.Window, entityName string, onDelete func()) {
	common.ShowDeleteConfirmation(
		window,
		"Підтвердження видалення",
		fmt.Sprintf("Ви впевнені, що хочете видалити %s?", entityName),
		onDelete,
	)
}

// buildStandardActions створює стандартний набір дій (Редагувати, Видалити, Деталі)
func buildStandardActions(onEdit func(), onDelete func(), onDetails func()) []common.Action {
	actions := make([]common.Action, 0)

	if onEdit != nil {
		actions = append(actions, common.Action{
			Label: "✏️ Редагувати",
			OnTap: onEdit,
		})
	}

	if onDelete != nil {
		actions = append(actions, common.Action{
			Label:      "🗑️ Видалити",
			OnTap:      onDelete,
			Importance: widget.DangerImportance,
		})
	}

	if onDetails != nil {
		actions = append(actions, common.Action{
			Label: "ℹ️ Деталі",
			OnTap: onDetails,
		})
	}

	return actions
}

// ============================================================================
// Logic Helper Functions
// ============================================================================

// FilterList фільтрує список елементів за пошуковим запитом та критерієм фільтрації.
// T - тип елемента списку (зазвичай вказівник на структуру).
func FilterList[T any](
	items []T,
	searchQuery string,
	filterValue string,
	matchesSearch func(item T, query string) bool,
	matchesFilter func(item T, filter string) bool,
) []T {
	filtered := make([]T, 0)
	for _, item := range items {
		// 1. Пошук
		if searchQuery != "" && matchesSearch != nil {
			if !matchesSearch(item, searchQuery) {
				continue
			}
		}

		// 2. Фільтр
		if filterValue != "" && matchesFilter != nil {
			if !matchesFilter(item, filterValue) {
				continue
			}
		}

		filtered = append(filtered, item)
	}
	return filtered
}
