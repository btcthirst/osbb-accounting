// presentation/fyne/screens/settings_screen.go
package screens

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/usecase/auth"
	"osbb-accounting/presentation/fyne/text"
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
	profileCard := widget.NewCard(text.TitleUserProfile, "", container.NewVBox(
		widget.NewLabel(fmt.Sprintf(text.LabelName, user.FullName)),
		widget.NewLabel(fmt.Sprintf(text.LabelLogin, user.Username)),
		widget.NewLabel(fmt.Sprintf(text.LabelEmail, user.Email)),
		widget.NewLabel(fmt.Sprintf(text.LabelRoles, s.authManager.(interface{ GetRoles() []string }).GetRoles())),
		widget.NewButtonWithIcon(text.ActionLogout, theme.LogoutIcon(), func() {
			s.authManager.Logout()
		}),
	))

	// 2. App Info Section
	appInfoCard := widget.NewCard(text.TitleAppInfo, "", container.NewVBox(
		widget.NewLabel(text.LabelAppName),
		widget.NewLabel(text.LabelAppVersion),
		widget.NewLabel(text.LabelAppDeveloper),
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
