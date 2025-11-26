package screens

import (
	"fmt"
	"strings"
	"time"
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
