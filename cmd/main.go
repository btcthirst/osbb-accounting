// cmd/main.go
package main

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/security"
	"osbb-accounting/presentation/fyne/screens"
)

func main() {
	// 1. Ініціалізація Fyne додатку
	myApp := app.NewWithID("com.osbb.management")
	myApp.Settings().SetTheme(&customTheme{})

	mainWindow := myApp.NewWindow("ОСББ Управління")
	mainWindow.Resize(fyne.NewSize(1024, 768))
	mainWindow.CenterOnScreen()

	// 2. Ініціалізація бази даних
	dbConfig := sqlite.DefaultConfig()
	//dbConfig.Path = "./data/osbb.db"

	db, err := sqlite.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlite.Close(db)

	// 3. Створення repositories
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	roleRepo := sqlite.NewRoleRepository(db)
	permissionRepo := sqlite.NewPermissionRepository(db)

	// 4. Створення security services
	passwordHasher := security.DefaultBCryptHasher()

	// 5. Створення AuthService
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		roleRepo,
		permissionRepo,
		passwordHasher,
	)

	// 6. Створення AuthManager
	authManager := screens.NewAuthManager(myApp, authService)

	// 7. Налаштування callbacks
	authManager.OnLoginSuccess(func() {
		// Після успішного входу показуємо головний екран
		showMainScreen(mainWindow, authManager)
	})

	authManager.OnLogout(func() {
		// Після виходу повертаємось на екран входу
		authManager.ShowLoginScreen(mainWindow)
	})

	// 8. Показуємо екран входу при запуску
	authManager.ShowLoginScreen(mainWindow)

	// 9. Cleanup при закритті
	mainWindow.SetOnClosed(func() {
		authManager.Cleanup()
	})

	// 10. Запуск додатку
	mainWindow.ShowAndRun()
}

// showMainScreen показує головний екран після успішного входу.
func showMainScreen(window fyne.Window, authManager *screens.AuthManager) {
	user := authManager.GetCurrentUser()

	// Привітання
	welcomeLabel := widget.NewLabel("Вітаємо, " + user.FullName + "!")
	welcomeLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Інформація про користувача
	userInfo := widget.NewCard(
		"Інформація про користувача",
		"",
		container.NewVBox(
			widget.NewLabel("Логін: "+user.Username),
			widget.NewLabel("Email: "+user.Email),
			widget.NewLabel("Ролі: "+fmt.Sprint(authManager.GetRoles())),
		),
	)

	// Меню навігації
	menuList := widget.NewList(
		func() int {
			items := []string{
				"📊 Головна",
				"👥 Власники",
				"🏢 Квартири",
				"💰 Нарахування",
				"💳 Платежі",
				"📈 Звіти",
				"⚙️ Налаштування",
			}
			return len(items)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			items := []string{
				"📊 Головна",
				"👥 Власники",
				"🏢 Квартири",
				"💰 Нарахування",
				"💳 Платежі",
				"📈 Звіти",
				"⚙️ Налаштування",
			}
			item.(*widget.Label).SetText(items[id])
		},
	)

	menuList.OnSelected = func(id widget.ListItemID) {
		// Тут буде логіка переключення екранів
		switch id {
		case 0:
			// Головна
		case 1:
			// Власники
			if authManager.HasPermission("owners.read") {
				// Показати екран власників
			}
		case 2:
			// Квартири
		case 3:
			// Нарахування
		case 4:
			// Платежі
		case 5:
			// Звіти
		case 6:
			// Налаштування
		}
	}

	// Кнопка виходу
	logoutButton := widget.NewButton("Вийти", func() {
		authManager.Logout()
	})
	logoutButton.Importance = widget.DangerImportance

	// Бокова панель
	sidebar := container.NewBorder(
		container.NewVBox(welcomeLabel, userInfo),
		logoutButton,
		nil,
		nil,
		menuList,
	)

	// Головний контент (placeholder)
	content := container.NewCenter(
		widget.NewLabel("Оберіть розділ з меню"),
	)

	// Layout
	split := container.NewHSplit(sidebar, content)
	split.Offset = 0.25 // 25% для sidebar

	window.SetContent(split)
}

// customTheme - власна тема (опціонально)
type customTheme struct{}

func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (t *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *customTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
