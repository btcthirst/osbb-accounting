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
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/screens"
)

func main() {
	// 1. Ініціалізація Fyne додатку
	myApp := app.NewWithID("com.osbb.management")
	myApp.Settings().SetTheme(&customTheme{})

	mainWindow := myApp.NewWindow("ОСББ Управління")
	mainWindow.Resize(fyne.NewSize(1200, 800))
	mainWindow.CenterOnScreen()

	// 2. Ініціалізація бази даних
	dbConfig := sqlite.DefaultConfig()
	// dbConfig.Path = "./osbb.db" // Розкоментуйте для використання файлу

	db, err := sqlite.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlite.Close(db)

	// 3. Створення repositories (Infrastructure Layer)
	// Auth
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	roleRepo := sqlite.NewRoleRepository(db)
	permissionRepo := sqlite.NewPermissionRepository(db)

	// Domain
	ownerRepo := sqlite.NewOwnerRepository(db)
	apartmentRepo := sqlite.NewApartmentRepository(db)
	ownershipRepo := sqlite.NewOwnershipShareRepository(db)
	// expenseRepo := sqlite.NewExpenseRepository(db) // Якщо буде реалізовано

	// 4. Створення services (Application Layer)
	passwordHasher := security.DefaultBCryptHasher()

	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		roleRepo,
		permissionRepo,
		passwordHasher,
	)

	ownerService := service.NewOwnerService(
		ownerRepo,
		permissionRepo,
	)

	apartmentService := service.NewApartmentService(
		apartmentRepo,
		permissionRepo,
	)

	ownershipService := service.NewOwnershipService(
		ownershipRepo,
		ownerRepo,
		apartmentRepo,
		permissionRepo,
	)

	// 5. Створення AuthManager (Presentation Layer)
	authManager := auth.NewAuthManager(myApp, authService)

	// 6. Налаштування callbacks навігації
	authManager.OnLoginSuccess(func() {
		// Передаємо всі необхідні сервіси у головний екран
		showMainScreen(mainWindow, authManager, ownerService, apartmentService, ownershipService)
	})

	authManager.OnLogout(func() {
		authManager.ShowLoginScreen(mainWindow)
	})

	// 7. Старт
	authManager.ShowLoginScreen(mainWindow)

	mainWindow.SetOnClosed(func() {
		authManager.Cleanup()
	})

	mainWindow.ShowAndRun()
}

// showMainScreen будує головний інтерфейс з боковим меню та контентом
func showMainScreen(
	window fyne.Window,
	authManager *auth.AuthManager,
	ownerService *service.OwnerService,
	apartmentService *service.ApartmentService,
	ownershipService *service.OwnershipService,
) {
	user := authManager.GetCurrentUser()

	// --- Header / Info Panel ---
	welcomeLabel := widget.NewLabel("👤 " + user.FullName)
	welcomeLabel.TextStyle = fyne.TextStyle{Bold: true}

	roleLabel := widget.NewLabel(fmt.Sprintf("Ролі: %v", authManager.GetRoles()))
	roleLabel.TextStyle = fyne.TextStyle{Italic: true}

	logoutButton := widget.NewButtonWithIcon("Вийти", theme.LogoutIcon(), func() {
		authManager.Logout()
	})
	logoutButton.Importance = widget.DangerImportance

	userInfoBox := container.NewVBox(
		welcomeLabel,
		roleLabel,
		widget.NewSeparator(),
		logoutButton,
	)

	// --- Content Container ---
	// Це контейнер, в якому буде змінюватись вміст при кліку на меню
	contentContainer := container.NewMax()

	// Початковий екран - Dashboard (Заглушка)
	dashboardLabel := widget.NewLabel("Ласкаво просимо до системи управління ОСББ!")
	dashboardLabel.Alignment = fyne.TextAlignCenter
	contentContainer.Objects = []fyne.CanvasObject{container.NewCenter(dashboardLabel)}

	// --- Navigation Menu ---
	menuItems := []string{
		"📊 Головна",
		"👥 Власники",
		"🏢 Квартири",
		"📝 Частки власності", // Ownership Shares
		"💰 Нарахування",
		"💳 Платежі",
		"⚙️ Налаштування",
	}

	menuList := widget.NewList(
		func() int { return len(menuItems) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.HomeIcon()), widget.NewLabel("Template"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			icon := box.Objects[0].(*widget.Icon)

			label.SetText(menuItems[id])

			// Іконки для меню
			switch id {
			case 0:
				icon.SetResource(theme.HomeIcon())
			case 1:
				icon.SetResource(theme.AccountIcon())
			case 2:
				icon.SetResource(theme.StorageIcon())
			case 3:
				icon.SetResource(theme.DocumentIcon())
			case 4:
				icon.SetResource(theme.GridIcon())
			case 5:
				icon.SetResource(theme.DocumentPrintIcon())
			case 6:
				icon.SetResource(theme.SettingsIcon())
			}
		},
	)

	// Логіка перемикання екранів
	menuList.OnSelected = func(id widget.ListItemID) {
		var newContent fyne.CanvasObject

		switch id {
		case 0: // Головна
			newContent = container.NewCenter(widget.NewLabel("📊 Дашборд (в розробці)"))

		case 1: // Власники
			if authManager.HasPermission("owners.read") {
				// Ініціалізуємо екран власників
				screen := screens.NewOwnersScreen(window, ownerService, authManager)
				newContent = screen.Render()
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		case 2: // Квартири
			if authManager.HasPermission("apartments.read") {
				// Примітка: Файл apartments_screen.go не був наданий, тому тут заглушка.
				// Якщо він у вас є, розкоментуйте:
				screen := screens.NewApartmentsScreen(window, apartmentService, authManager)
				newContent = screen.Render()
				//newContent = container.NewCenter(widget.NewLabel("🏢 Екран квартир (в розробці/файл відсутній)"))
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		case 3: // Частки власності
			if authManager.HasPermission("ownership.read") {
				// Ініціалізуємо екран часток
				screen := screens.NewOwnershipSharesScreen(
					window,
					apartmentService,
					ownerService,
					ownershipService,
					authManager,
				)
				newContent = screen.Render()
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		default:
			newContent = container.NewCenter(widget.NewLabel(fmt.Sprintf("Розділ '%s' в розробці", menuItems[id])))
		}

		// Оновлюємо центральний контейнер
		contentContainer.Objects = []fyne.CanvasObject{newContent}
		contentContainer.Refresh()
	}

	// --- Layout Assembly ---

	// Ліва панель (Меню)
	sidebar := container.NewBorder(
		nil,         // top
		userInfoBox, // bottom
		nil,         // left
		nil,         // right
		menuList,    // center
	)

	// Розділювач
	split := container.NewHSplit(sidebar, contentContainer)
	split.Offset = 0.25 // 25% ширини для меню

	window.SetContent(split)
}

func createAccessDeniedPlaceholder() fyne.CanvasObject {
	icon := widget.NewIcon(theme.WarningIcon())
	label := widget.NewLabel("Доступ заборонено")
	label.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewCenter(container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(label),
	))
}

// --- Custom Theme Implementation ---

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
