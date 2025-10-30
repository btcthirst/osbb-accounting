package ui

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MainWindow представляє головне вікно додатку після входу.
// Містить навігаційне меню та контент-область для різних модулів.
type MainWindow struct {
	// window - посилання на головне вікно
	window fyne.Window

	// currentUser - дані поточного авторизованого користувача
	currentUser *domain.User

	// authService - сервіс для операцій з користувачами
	authService *service.AuthService

	// onLogout - callback для виходу з системи
	onLogout func()

	// UI елементи
	contentContainer *fyne.Container
	menuList         *widget.List
	userInfoLabel    *widget.RichText
}

// MenuSection представляє розділ меню з групою пунктів.
type MenuSection struct {
	Title string
	Items []MenuItem
}

// MenuItem представляє окремий пункт меню.
type MenuItem struct {
	Icon     fyne.Resource
	Label    string
	Badge    string // Опціональна мітка (наприклад, кількість нових)
	OnTap    func()
	Disabled bool
}

// NewMainWindow створює нове головне вікно додатку.
//
// Параметри:
//   - window: головне вікно Fyne
//   - user: авторизований користувач
//   - authService: сервіс аутентифікації
//   - onLogout: callback для виходу
//
// Повертає:
//   - *MainWindow: новий екземпляр головного вікна
func NewMainWindow(
	window fyne.Window,
	user *domain.User,
	authService *service.AuthService,
	onLogout func(),
) *MainWindow {
	mw := &MainWindow{
		window:      window,
		currentUser: user,
		authService: authService,
		onLogout:    onLogout,
	}

	mw.initUI()
	return mw
}

// initUI ініціалізує UI елементи головного вікна.
func (mw *MainWindow) initUI() {
	// Контейнер для основного контенту
	mw.contentContainer = container.NewMax()

	// Показуємо початковий екран (Dashboard)
	mw.showDashboard()
}

// getMenuSections повертає структуру меню залежно від ролі користувача.
func (mw *MainWindow) getMenuSections() []MenuSection {
	sections := []MenuSection{
		{
			Title: "Основне",
			Items: []MenuItem{
				{
					Icon:  theme.HomeIcon(),
					Label: "Головна",
					OnTap: mw.showDashboard,
				},
			},
		},
	}

	// Додаємо розділ управління для користувачів з відповідними правами
	if mw.currentUser.CanManageFinances() || mw.currentUser.IsAdmin() {
		managementSection := MenuSection{
			Title: "Управління",
			Items: []MenuItem{},
		}

		// Квартири доступні всім ролям
		managementSection.Items = append(managementSection.Items, MenuItem{
			Icon:  theme.FolderIcon(),
			Label: "Квартири",
			OnTap: mw.showApartments,
		})

		// Платежі доступні бухгалтерам та адмінам
		if mw.currentUser.CanManageFinances() {
			managementSection.Items = append(managementSection.Items, MenuItem{
				Icon:  theme.DocumentIcon(),
				Label: "Платежі",
				OnTap: mw.showPayments,
			})
		}

		sections = append(sections, managementSection)
	}

	// Розділ звітності
	if mw.currentUser.CanManageFinances() || mw.currentUser.HasRole(domain.RoleHead) {
		reportsSection := MenuSection{
			Title: "Звіти",
			Items: []MenuItem{
				{
					Icon:  theme.DocumentIcon(),
					Label: "Борги",
					OnTap: mw.showDebtsReport,
				},
				{
					Icon:  theme.InfoIcon(),
					Label: "Загальний звіт",
					OnTap: mw.showGeneralReport,
				},
			},
		}
		sections = append(sections, reportsSection)
	}

	// Адміністрування (тільки для адмінів)
	if mw.currentUser.IsAdmin() {
		adminSection := MenuSection{
			Title: "Адміністрування",
			Items: []MenuItem{
				{
					Icon:  theme.AccountIcon(),
					Label: "Користувачі",
					OnTap: mw.showUsers,
				},
				{
					Icon:  theme.SettingsIcon(),
					Label: "Налаштування",
					OnTap: mw.showSettings,
				},
			},
		}
		sections = append(sections, adminSection)
	}

	// Розділ профілю (доступний всім)
	profileSection := MenuSection{
		Title: "Профіль",
		Items: []MenuItem{
			{
				Icon:  theme.AccountIcon(),
				Label: "Мій профіль",
				OnTap: mw.showProfile,
			},
			{
				Icon:  theme.LogoutIcon(),
				Label: "Вийти",
				OnTap: mw.handleLogout,
			},
		},
	}
	sections = append(sections, profileSection)

	return sections
}

// createMenu створює навігаційне меню.
func (mw *MainWindow) createMenu() fyne.CanvasObject {
	sections := mw.getMenuSections()

	// Розгортаємо всі пункти меню в один список
	var allItems []MenuItem
	var sectionHeaders []int // Індекси заголовків розділів

	for _, section := range sections {
		// Додаємо заголовок розділу як окремий item
		sectionHeaders = append(sectionHeaders, len(allItems))
		allItems = append(allItems, MenuItem{Label: section.Title})

		// Додаємо пункти розділу
		allItems = append(allItems, section.Items...)
	}

	// Створюємо List widget
	list := widget.NewList(
		func() int {
			return len(allItems)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			item := allItems[id]
			box := obj.(*fyne.Container)

			// Перевіряємо, чи це заголовок розділу
			isHeader := false
			for _, headerIdx := range sectionHeaders {
				if id == headerIdx {
					isHeader = true
					break
				}
			}

			if isHeader {
				// Це заголовок розділу
				icon := box.Objects[0].(*widget.Icon)
				icon.Hide()

				label := box.Objects[1].(*widget.Label)
				label.SetText(item.Label)
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				// Це звичайний пункт меню
				icon := box.Objects[0].(*widget.Icon)
				if item.Icon != nil {
					icon.SetResource(item.Icon)
					icon.Show()
				} else {
					icon.Hide()
				}

				label := box.Objects[1].(*widget.Label)
				label.SetText(item.Label)
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		// Перевіряємо, чи не заголовок це
		isHeader := false
		for _, headerIdx := range sectionHeaders {
			if id == headerIdx {
				isHeader = true
				break
			}
		}

		if !isHeader && allItems[id].OnTap != nil {
			allItems[id].OnTap()
		}

		// Знімаємо виділення
		list.UnselectAll()
	}

	// Інформація про користувача вгорі меню
	mw.userInfoLabel = widget.NewRichTextFromMarkdown(fmt.Sprintf(
		"**%s**\n%s",
		mw.currentUser.FullName,
		mw.getRoleDisplayName(),
	))

	userInfo := container.NewVBox(
		container.NewPadded(
			container.NewVBox(
				widget.NewIcon(theme.AccountIcon()),
				mw.userInfoLabel,
			),
		),
		widget.NewSeparator(),
	)

	return container.NewBorder(
		userInfo,
		nil,
		nil,
		nil,
		list,
	)
}

// getRoleDisplayName повертає українську назву ролі для відображення.
func (mw *MainWindow) getRoleDisplayName() string {
	switch mw.currentUser.Role {
	case domain.RoleAdmin:
		return "Адміністратор"
	case domain.RoleHead:
		return "Голова ОСББ"
	case domain.RoleAccountant:
		return "Бухгалтер"
	default:
		return "Користувач"
	}
}

// Render повертає Fyne контейнер головного вікна.
func (mw *MainWindow) Render() fyne.CanvasObject {
	// Створюємо меню
	menu := mw.createMenu()

	// Split контейнер: меню зліва, контент справа
	split := container.NewHSplit(
		container.NewBorder(nil, nil, nil, nil, menu),
		mw.contentContainer,
	)

	// Встановлюємо початковий розмір меню (25% ширини)
	split.Offset = 0.2

	return split
}

// handleLogout обробляє вихід з системи.
func (mw *MainWindow) handleLogout() {
	dialog.ShowConfirm(
		"Вихід",
		"Ви впевнені, що хочете вийти з системи?",
		func(confirm bool) {
			if confirm && mw.onLogout != nil {
				mw.onLogout()
			}
		},
		mw.window,
	)
}

// setContent змінює вміст контент-області.
func (mw *MainWindow) setContent(content fyne.CanvasObject) {
	mw.contentContainer.Objects = []fyne.CanvasObject{content}
	mw.contentContainer.Refresh()
}

// Методи для показу різних екранів (заглушки для MVP)

func (mw *MainWindow) showDashboard() {
	mw.setContent(mw.createDashboardView())
}

func (mw *MainWindow) showApartments() {
	mw.setContent(mw.createPlaceholder("Квартири", "Модуль управління квартирами буде реалізований наступним"))
}

func (mw *MainWindow) showPayments() {
	mw.setContent(mw.createPlaceholder("Платежі", "Модуль обліку платежів буде реалізований наступним"))
}

func (mw *MainWindow) showDebtsReport() {
	mw.setContent(mw.createPlaceholder("Звіт про борги", "Модуль звітності про борги буде реалізований наступним"))
}

func (mw *MainWindow) showGeneralReport() {
	mw.setContent(mw.createPlaceholder("Загальний звіт", "Модуль загальної звітності буде реалізований наступним"))
}

func (mw *MainWindow) showUsers() {
	mw.setContent(mw.createPlaceholder("Користувачі", "Модуль управління користувачами буде реалізований наступним"))
}

func (mw *MainWindow) showSettings() {
	mw.setContent(mw.createPlaceholder("Налаштування", "Модуль налаштувань буде реалізований наступним"))
}

func (mw *MainWindow) showProfile() {
	mw.setContent(mw.createProfileView())
}

// createDashboardView створює головну інформаційну панель.
func (mw *MainWindow) createDashboardView() fyne.CanvasObject {
	welcomeText := widget.NewRichTextFromMarkdown(fmt.Sprintf(
		"# Вітаємо, %s!\n\nВи увійшли як **%s**",
		mw.currentUser.FullName,
		mw.getRoleDisplayName(),
	))

	infoCard := widget.NewCard(
		"Система Обліку ОСББ",
		"",
		container.NewVBox(
			welcomeText,
			layout.NewSpacer(),
			widget.NewLabel("Оберіть розділ у меню ліворуч для початку роботи."),
		),
	)

	// TODO: Додати статистичні картки (загальна сума боргів, кількість квартир, тощо)

	return container.NewPadded(
		container.NewVBox(
			infoCard,
		),
	)
}

// createProfileView створює екран профілю користувача.
func (mw *MainWindow) createProfileView() fyne.CanvasObject {
	profileCard := widget.NewCard(
		"Мій профіль",
		"",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Ім'я користувача: %s", mw.currentUser.Username)),
			widget.NewLabel(fmt.Sprintf("Повне ім'я: %s", mw.currentUser.FullName)),
			widget.NewLabel(fmt.Sprintf("Роль: %s", mw.getRoleDisplayName())),
			widget.NewLabel(fmt.Sprintf("Дата створення: %s", mw.currentUser.CreatedAt.Format("02.01.2006"))),
		),
	)

	changePasswordBtn := widget.NewButton("Змінити пароль", func() {
		mw.showChangePasswordDialog()
	})

	return container.NewPadded(
		container.NewVBox(
			profileCard,
			changePasswordBtn,
		),
	)
}

// showChangePasswordDialog відображає діалог зміни пароля.
func (mw *MainWindow) showChangePasswordDialog() {
	oldPasswordEntry := widget.NewPasswordEntry()
	oldPasswordEntry.SetPlaceHolder("Поточний пароль")

	newPasswordEntry := widget.NewPasswordEntry()
	newPasswordEntry.SetPlaceHolder("Новий пароль (мін. 6 символів)")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Підтвердіть новий пароль")

	form := container.NewVBox(
		widget.NewLabel("Поточний пароль:"),
		oldPasswordEntry,
		widget.NewLabel("Новий пароль:"),
		newPasswordEntry,
		widget.NewLabel("Підтвердження:"),
		confirmPasswordEntry,
	)

	dialog.NewCustomConfirm(
		"Зміна пароля",
		"Змінити",
		"Скасувати",
		form,
		func(change bool) {
			if !change {
				return
			}

			oldPass := oldPasswordEntry.Text
			newPass := newPasswordEntry.Text
			confirmPass := confirmPasswordEntry.Text

			if newPass != confirmPass {
				dialog.ShowError(fmt.Errorf("паролі не співпадають"), mw.window)
				return
			}

			if len(newPass) < 6 {
				dialog.ShowError(fmt.Errorf("пароль має бути не менше 6 символів"), mw.window)
				return
			}

			err := mw.authService.ChangePassword(mw.currentUser.ID, oldPass, newPass)
			if err != nil {
				dialog.ShowError(err, mw.window)
				return
			}

			dialog.ShowInformation("Успіх", "Пароль успішно змінено!", mw.window)
		},
		mw.window,
	).Show()
}

// createPlaceholder створює заглушку для модулів, що ще не реалізовані.
func (mw *MainWindow) createPlaceholder(title, message string) fyne.CanvasObject {
	icon := widget.NewIcon(theme.InfoIcon())

	card := widget.NewCard(
		title,
		"",
		container.NewVBox(
			container.NewCenter(icon),
			layout.NewSpacer(),
			widget.NewLabel(message),
			layout.NewSpacer(),
		),
	)

	return container.NewPadded(
		container.NewCenter(card),
	)
}
