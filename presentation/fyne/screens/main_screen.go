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
// MenuItem defines a single item in the navigation menu
type MenuItem struct {
	Title      string
	Icon       fyne.Resource
	Permission string // Empty if public
	ScreenFunc func() fyne.CanvasObject
}

// ShowMainScreen будує головний інтерфейс з боковим меню та контентом
func ShowMainScreen(
	window fyne.Window,
	authManager *auth.AuthManager,
	services *service.ServiceContainer,
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
	contentContainer := container.NewStack()

	// Початковий екран - Dashboard
	dashboardScreen := NewDashboardScreen(
		window,
		services.ChargeService,
		services.PaymentService,
		authManager,
	)
	contentContainer.Objects = []fyne.CanvasObject{dashboardScreen.Render()}

	// --- Navigation Menu Definition ---
	allMenuItems := []MenuItem{
		{
			Title: "📊 Головна",
			Icon:  theme.HomeIcon(),
			ScreenFunc: func() fyne.CanvasObject {
				return NewDashboardScreen(window, services.ChargeService, services.PaymentService, authManager).Render()
			},
		},
		{
			Title: "🏠 Про ОСББ",
			Icon:  theme.InfoIcon(),
			ScreenFunc: func() fyne.CanvasObject {
				return NewOSBBScreen(window, services.OSBBService, authManager).Render()
			},
		},
		{
			Title:      "👥 Власники",
			Icon:       theme.AccountIcon(),
			Permission: "owners.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewOwnersScreen(window, services.OwnerService, authManager).Render()
			},
		},
		{
			Title:      "🏢 Квартири",
			Icon:       theme.StorageIcon(),
			Permission: "apartments.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewApartmentsScreen(window, services.ApartmentService, authManager).Render()
			},
		},
		{
			Title:      "📝 Частки власності",
			Icon:       theme.DocumentIcon(),
			Permission: "ownership.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewOwnershipSharesScreen(window, services.ApartmentService, services.OwnerService, services.OwnershipService, authManager).Render()
			},
		},
		{
			Title:      "💰 Нарахування",
			Icon:       theme.GridIcon(),
			Permission: "charges.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewChargesScreen(window, services.ChargeService, services.OwnershipService, authManager).Render()
			},
		},
		{
			Title:      "💳 Платежі",
			Icon:       theme.DocumentPrintIcon(),
			Permission: "payments.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewPaymentsScreen(window, services.PaymentService, services.OwnershipService, authManager).Render()
			},
		},
		{
			Title:      "💰 Рух коштів",
			Icon:       theme.DocumentSaveIcon(),
			Permission: "payments.read", // Використовуємо той самий дозвіл як для платежів
			ScreenFunc: func() fyne.CanvasObject {
				return NewCashFlowScreen(window, services.CashFlowService, authManager).Render()
			},
		},
		{
			Title:      "💸 Витрати",
			Icon:       theme.ContentRemoveIcon(),
			Permission: "expenses.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewExpensesScreen(window, services.ExpenseService, authManager).Render()
			},
		},
		{
			Title:      "📂 Категорії витрат",
			Icon:       theme.ListIcon(),
			Permission: "expenses.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewExpenseCategoryScreen(window, services.ExpenseCategoryService, authManager).Render()
			},
		},
		{
			Title:      "👷 Підрядники",
			Icon:       theme.FolderIcon(),
			Permission: "contractors.read",
			ScreenFunc: func() fyne.CanvasObject {
				return NewContractorsScreen(window, services.ContractorService, authManager).Render()
			},
		},
		{
			Title: "📥 Імпорт/Експорт",
			Icon:  theme.DownloadIcon(),
			ScreenFunc: func() fyne.CanvasObject {
				userID := authManager.GetCurrentUser().ID
				return NewImportScreen(window, services.ImportService, userID).Render()
			},
		},
		{
			Title: "⚙️ Налаштування",
			Icon:  theme.SettingsIcon(),
			ScreenFunc: func() fyne.CanvasObject {
				return NewSettingsScreen(window, authManager).Render()
			},
		},
	}

	// Filter menu items based on permissions
	var visibleMenuItems []MenuItem
	for _, item := range allMenuItems {
		if item.Permission == "" || authManager.HasPermission(item.Permission) {
			visibleMenuItems = append(visibleMenuItems, item)
		}
	}

	// --- Navigation Menu UI ---
	menuList := widget.NewList(
		func() int { return len(visibleMenuItems) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.HomeIcon()), widget.NewLabel("Template"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)

			menuItem := visibleMenuItems[id]
			label.SetText(menuItem.Title)
			icon.SetResource(menuItem.Icon)
		},
	)

	menuList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(visibleMenuItems) {
			newContent := visibleMenuItems[id].ScreenFunc()
			contentContainer.Objects = []fyne.CanvasObject{newContent}
			contentContainer.Refresh()
		}
	}

	// --- Layout Assembly ---
	sidebar := container.NewBorder(
		nil,         // top
		userInfoBox, // bottom
		nil,         // left
		nil,         // right
		menuList,    // center
	)

	split := container.NewHSplit(sidebar, contentContainer)
	split.Offset = 0.25 // 25% width for menu

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
