package screens

import (
	"osbb-accounting/presentation/fyne/text"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UI Button Helpers

// newCreateButton creates a standard "Create" button with high importance
func newCreateButton(label string, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, theme.ContentAddIcon(), onTap)
	btn.Importance = widget.HighImportance
	return btn
}

// newRefreshButton creates a standard "Refresh" button
func newRefreshButton(onTap func()) *widget.Button {
	return widget.NewButtonWithIcon(text.ActionRefresh, theme.ViewRefreshIcon(), onTap)
}

// UI Input Helpers

// newSearchEntry creates a standard search entry with placeholder
func newSearchEntry(placeholder string, onChanged func(string)) *widget.Entry {
	entry := widget.NewEntry()
	if placeholder == "" {
		placeholder = text.SearchPlaceholder
	}
	entry.SetPlaceHolder(placeholder)
	if onChanged != nil {
		entry.OnChanged = onChanged
	}
	return entry
}

// newFilterSelect creates a standard filter select dropdown
func newFilterSelect(options []string, onChanged func(string)) *widget.Select {
	sel := widget.NewSelect(options, onChanged)
	return sel
}
