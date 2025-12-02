// Package shared містить спільні типи для use cases
package shared

// DeleteOutput - стандартний результат операції видалення.
// Використовується різними use cases для уніфікації відповідей.
type DeleteOutput struct {
	Success bool
	Message string
}

// NewDeleteOutput створює новий DeleteOutput з успішним результатом.
func NewDeleteOutput(message string) *DeleteOutput {
	return &DeleteOutput{
		Success: true,
		Message: message,
	}
}

// NewDeleteFailure створює новий DeleteOutput з невдалим результатом.
func NewDeleteFailure(message string) *DeleteOutput {
	return &DeleteOutput{
		Success: false,
		Message: message,
	}
}
