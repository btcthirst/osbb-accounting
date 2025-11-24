package screens

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/presentation/fyne/auth"
)

// ShowMainScreen будує головний інтерфейс з боковим меню та контентом
func ShowMainScreen(
	window fyne.Window,
	authManager *auth.AuthManager,
	ownerService *service.OwnerService,
	apartmentService *service.ApartmentService,
	ownershipService *service.OwnershipService,
	chargeService *service.ChargeService,
	paymentService *service.PaymentService,
	expenseService *service.ExpenseService,
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
	contentContainer := container.NewStack()

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
			screen := NewDashboardScreen(
				window,
				chargeService,
				paymentService,
				authManager,
			)
			newContent = screen.Render()

		case 1: // Власники
			if authManager.HasPermission("owners.read") {
				// Ініціалізуємо екран власників
				screen := NewOwnersScreen(window, ownerService, authManager)
				newContent = screen.Render()
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		case 2: // Квартири
			if authManager.HasPermission("apartments.read") {
				screen := NewApartmentsScreen(window, apartmentService, authManager)
				newContent = screen.Render()
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		case 3: // Частки власності
			if authManager.HasPermission("ownership.read") {
				// Ініціалізуємо екран часток
				screen := NewOwnershipSharesScreen(
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

		case 4: // Нарахування
			if authManager.HasPermission("charges.read") {
				screen := NewChargesScreen(
					window,
					chargeService,
					ownershipService,
					authManager,
				)
				newContent = screen.Render()
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		case 6: // Витрати
			if authManager.HasPermission("expenses.read") {
				screen := NewExpensesScreen(
					window,
					expenseService,
					authManager,
				)
				newContent = screen.Render()
			} else {
				newContent = createAccessDeniedPlaceholder()
			}

		case 7: // Налаштування
			screen := NewSettingsScreen(window, authManager)
			newContent = screen.Render()

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
