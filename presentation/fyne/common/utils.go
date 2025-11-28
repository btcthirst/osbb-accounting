package common

// GetMonthName повертає назву місяця українською мовою.
func GetMonthName(month int) string {
	months := []string{
		"", "Січень", "Лютий", "Березень", "Квітень", "Травень", "Червень",
		"Липень", "Серпень", "Вересень", "Жовтень", "Листопад", "Грудень",
	}

	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}
