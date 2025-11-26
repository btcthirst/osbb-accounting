package common

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Action defines a button in the actions menu
type Action struct {
	Label      string
	Icon       fyne.Resource
	OnTap      func()
	Importance widget.Importance // Default is widget.MediumImportance
}

// ShowActionsMenu displays a custom dialog with a list of actions
func ShowActionsMenu(window fyne.Window, title string, actions []Action) {
	var buttons []fyne.CanvasObject

	// Title
	buttons = append(buttons, widget.NewLabel(title))
	buttons = append(buttons, layout.NewSpacer())

	for _, action := range actions {
		btn := widget.NewButtonWithIcon(action.Label, action.Icon, action.OnTap)
		if action.Importance != 0 {
			btn.Importance = action.Importance
		}
		buttons = append(buttons, btn)
	}

	content := container.NewVBox(buttons...)
	dialog.ShowCustom("Дії", "Закрити", content, window)
}

// ShowDeleteConfirmation displays a confirmation dialog for deletion
func ShowDeleteConfirmation(window fyne.Window, title, message string, onConfirm func()) {
	dialog.ShowConfirm(
		title,
		message,
		func(confirmed bool) {
			if confirmed {
				onConfirm()
			}
		},
		window,
	)
}

// ShowError displays an error dialog
func ShowError(window fyne.Window, err error) {
	dialog.ShowError(err, window)
}

// ShowSuccess displays a success information dialog
func ShowSuccess(window fyne.Window, message string) {
	dialog.ShowInformation("Успіх", message, window)
}

// ShowInformation displays an information dialog
func ShowInformation(window fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, window)
}
