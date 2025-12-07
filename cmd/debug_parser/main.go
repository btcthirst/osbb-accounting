package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	tests := []string{
		"1 730,00",
		"1.730,00",
		"350,00",
		"175,00",
		"1\u00a0730,00", // Non-breaking space
		"1 907,00",
		"1.907,00",
	}

	for _, t := range tests {
		res := parseFloat(t)
		fmt.Printf("Input: '%s' -> Output: %f\n", t, res)
	}
}

func parseFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	// Check if we have a comma (decimal separator in this locale)
	hasComma := strings.Contains(value, ",")

	var sb strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		} else if r == '-' {
			sb.WriteRune(r)
		} else if r == ',' {
			sb.WriteRune('.') // Convert decimal comma to dot
		} else if r == '.' {
			if !hasComma {
				sb.WriteRune('.') // Keep dot only if no comma exists (assume dot is decimal)
			}
			// If hasComma is true, skip dot (assume it's a thousand separator)
		}
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
