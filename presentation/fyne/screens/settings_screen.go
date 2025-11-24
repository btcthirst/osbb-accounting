// presentation/fyne/screens/settings_screen.go
package screens

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/usecase/auth"
)

// SettingsScreen - екран налаштувань
type SettingsScreen struct {
	window      fyne.Window
	authManager interface {
		GetCurrentUser() *auth.UserInfo
		Logout()
	}
}

// NewSettingsScreen створює новий екран налаштувань
func NewSettingsScreen(
	window fyne.Window,
	authManager interface {
		GetCurrentUser() *auth.UserInfo
		Logout()
	},
) *SettingsScreen {
	return &SettingsScreen{
		window:      window,
		authManager: authManager,
	}
}

func (s *SettingsScreen) Render() fyne.CanvasObject {
	user := s.authManager.GetCurrentUser()

	// 1. User Profile Section
	profileCard := widget.NewCard("Профіль користувача", "", container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Ім'я: %s", user.FullName)),
		widget.NewLabel(fmt.Sprintf("Логін: %s", user.Username)),
		widget.NewLabel(fmt.Sprintf("Email: %s", user.Email)),
		widget.NewLabel(fmt.Sprintf("Ролі: %v", s.authManager.(interface{ GetRoles() []string }).GetRoles())),
		widget.NewButtonWithIcon("Вийти з системи", theme.LogoutIcon(), func() {
			s.authManager.Logout()
		}),
	))

	// 2. App Info Section
	appInfoCard := widget.NewCard("Про програму", "", container.NewVBox(
		widget.NewLabel("OSBB Accounting System"),
		widget.NewLabel("Версія: 1.0.0"),
		widget.NewLabel("Розробник: Google Deepmind Team"),
	))

	// Layout
	return container.NewVBox(
		profileCard,
		widget.NewSeparator(),
		appInfoCard,
	)
}

// Helper to safely dereference string pointer
func safePtrToString(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}
