package ui

import (
	"fmt"
	"osbb-accounting/domain"
	"osbb-accounting/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UsersScreen представляє екран управління користувачами (тільки для адмінів).
type UsersScreen struct {
	window      fyne.Window
	userService *service.UserService
	currentUser *domain.User

	// UI елементи
	usersList *widget.List
	users     []*domain.User
}

// NewUsersScreen створює новий екран управління користувачами.
func NewUsersScreen(
	window fyne.Window,
	userService *service.UserService,
	currentUser *domain.User,
) *UsersScreen {
	screen := &UsersScreen{
		window:      window,
		userService: userService,
		currentUser: currentUser,
		users:       []*domain.User{},
	}

	screen.initUI()
	screen.loadUsers()

	return screen
}

// initUI ініціалізує UI елементи.
func (s *UsersScreen) initUI() {
	// Список користувачів
	s.usersList = widget.NewList(
		func() int {
			return len(s.users)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				container.NewVBox(
					widget.NewLabel("Template"),
					widget.NewLabel("Template"),
				),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(s.users) {
				return
			}

			user := s.users[id]
			box := obj.(*fyne.Container)

			icon := box.Objects[0].(*widget.Icon)
			if user.IsAdmin() {
				icon.SetResource(theme.AccountIcon())
			} else {
				icon.SetResource(theme.AccountIcon())
			}

			infoBox := box.Objects[1].(*fyne.Container)
			nameLabel := infoBox.Objects[0].(*widget.Label)
			detailsLabel := infoBox.Objects[1].(*widget.Label)

			nameLabel.SetText(user.FullName)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			roleDisplay := getRoleDisplayName(user.Role)
			statusIcon := "✓"
			if !user.IsActive {
				statusIcon = "✗"
			}

			detailsLabel.SetText(fmt.Sprintf("%s %s • %s",
				statusIcon, user.Username, roleDisplay))
		},
	)

	s.usersList.OnSelected = func(id widget.ListItemID) {
		if id >= len(s.users) {
			return
		}
		s.showUserDetails(s.users[id])
		s.usersList.UnselectAll()
	}
}

// Render повертає UI компонент екрану.
func (s *UsersScreen) Render() fyne.CanvasObject {
	// Тулбар
	toolbar := s.createToolbar()

	// Статистика
	stats := s.createStatsCard()

	// Основний контент
	content := container.NewBorder(
		container.NewVBox(toolbar, stats),
		nil,
		nil,
		nil,
		s.usersList,
	)

	return container.NewPadded(content)
}

// createToolbar створює панель інструментів.
func (s *UsersScreen) createToolbar() *fyne.Container {
	addBtn := widget.NewButtonWithIcon("Додати користувача", theme.ContentAddIcon(), func() {
		s.showAddUserDialog()
	})
	addBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButtonWithIcon("Оновити", theme.ViewRefreshIcon(), func() {
		s.loadUsers()
	})

	return container.NewHBox(addBtn, refreshBtn)
}

// createStatsCard створює картку зі статистикою.
func (s *UsersScreen) createStatsCard() *widget.Card {
	adminCount := 0
	headCount := 0
	accountantCount := 0
	activeCount := 0

	for _, user := range s.users {
		if user.IsActive {
			activeCount++
		}
		switch user.Role {
		case domain.RoleAdmin:
			adminCount++
		case domain.RoleHead:
			headCount++
		case domain.RoleAccountant:
			accountantCount++
		}
	}

	statsText := fmt.Sprintf(
		"Всього: %d • Активних: %d • Адмінів: %d • Голова: %d • Бухгалтерів: %d",
		len(s.users), activeCount, adminCount, headCount, accountantCount,
	)

	return widget.NewCard("", "", widget.NewLabel(statsText))
}

// loadUsers завантажує список користувачів.
func (s *UsersScreen) loadUsers() {
	users, err := s.userService.GetAllUsers(s.currentUser.ID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("помилка завантаження користувачів: %w", err), s.window)
		return
	}

	s.users = users
	s.usersList.Refresh()
}

// showUserDetails показує детальну інформацію про користувача.
func (s *UsersScreen) showUserDetails(user *domain.User) {
	details := fmt.Sprintf(
		"Користувач: %s\n\n"+
			"Повне ім'я: %s\n"+
			"Роль: %s\n"+
			"Статус: %s\n"+
			"Створено: %s\n",
		user.Username,
		user.FullName,
		getRoleDisplayName(user.Role),
		getStatusText(user.IsActive),
		user.CreatedAt.Format("02.01.2006 15:04"),
	)

	detailsLabel := widget.NewLabel(details)

	// Кнопки дій
	editBtn := widget.NewButton("Редагувати", func() {
		s.showEditUserDialog(user)
	})

	changeRoleBtn := widget.NewButton("Змінити роль", func() {
		s.showChangeRoleDialog(user)
	})

	resetPasswordBtn := widget.NewButton("Скинути пароль", func() {
		s.showResetPasswordDialog(user)
	})

	var toggleBtn *widget.Button
	if user.IsActive {
		toggleBtn = widget.NewButton("Деактивувати", func() {
			s.confirmDeactivate(user)
		})
		toggleBtn.Importance = widget.WarningImportance
	} else {
		toggleBtn = widget.NewButton("Активувати", func() {
			s.confirmActivate(user)
		})
		toggleBtn.Importance = widget.SuccessImportance
	}

	// Перевірка, чи не намагається адмін деактивувати себе
	if user.ID == s.currentUser.ID {
		toggleBtn.Disable()
	}

	buttons := container.NewVBox(
		editBtn,
		changeRoleBtn,
		resetPasswordBtn,
		toggleBtn,
	)

	content := container.NewBorder(
		detailsLabel,
		nil,
		nil,
		nil,
		buttons,
	)

	d := dialog.NewCustom("Деталі користувача", "Закрити", content, s.window)
	d.Resize(fyne.NewSize(400, 450))
	d.Show()
}

// showAddUserDialog показує діалог додавання користувача.
func (s *UsersScreen) showAddUserDialog() {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Мінімум 6 символів")

	fullNameEntry := widget.NewEntry()
	fullNameEntry.SetPlaceHolder("Прізвище Ім'я По батькові")

	roleSelect := widget.NewSelect(
		[]string{"Адміністратор", "Голова ОСББ", "Бухгалтер"},
		nil,
	)
	roleSelect.SetSelected("Бухгалтер")

	form := container.NewVBox(
		widget.NewLabel("Ім'я користувача:"),
		usernameEntry,
		widget.NewLabel("Пароль:"),
		passwordEntry,
		widget.NewLabel("Повне ім'я:"),
		fullNameEntry,
		widget.NewLabel("Роль:"),
		roleSelect,
	)

	dialog.NewCustomConfirm(
		"Додати користувача",
		"Створити",
		"Скасувати",
		form,
		func(create bool) {
			if !create {
				return
			}

			role := getRoleFromDisplay(roleSelect.Selected)

			_, err := s.userService.CreateUser(
				s.currentUser.ID,
				usernameEntry.Text,
				passwordEntry.Text,
				role,
				fullNameEntry.Text,
			)

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Користувача створено!", s.window)
			s.loadUsers()
		},
		s.window,
	).Show()
}

// showEditUserDialog показує діалог редагування користувача.
func (s *UsersScreen) showEditUserDialog(user *domain.User) {
	fullNameEntry := widget.NewEntry()
	fullNameEntry.SetText(user.FullName)

	form := container.NewVBox(
		widget.NewLabel("Повне ім'я:"),
		fullNameEntry,
		widget.NewLabel(fmt.Sprintf("Ім'я користувача: %s (незмінне)", user.Username)),
	)

	dialog.NewCustomConfirm(
		"Редагувати користувача",
		"Зберегти",
		"Скасувати",
		form,
		func(save bool) {
			if !save {
				return
			}

			_, err := s.userService.UpdateUser(
				s.currentUser.ID,
				user.ID,
				fullNameEntry.Text,
				"", // роль не змінюємо тут
			)

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Зміни збережено!", s.window)
			s.loadUsers()
		},
		s.window,
	).Show()
}

// showChangeRoleDialog показує діалог зміни ролі.
func (s *UsersScreen) showChangeRoleDialog(user *domain.User) {
	currentRole := getRoleDisplayName(user.Role)

	roleSelect := widget.NewSelect(
		[]string{"Адміністратор", "Голова ОСББ", "Бухгалтер"},
		nil,
	)
	roleSelect.SetSelected(currentRole)

	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Користувач: %s", user.FullName)),
		widget.NewLabel(fmt.Sprintf("Поточна роль: %s", currentRole)),
		widget.NewLabel("Нова роль:"),
		roleSelect,
	)

	dialog.NewCustomConfirm(
		"Зміна ролі",
		"Змінити",
		"Скасувати",
		form,
		func(change bool) {
			if !change {
				return
			}

			newRole := getRoleFromDisplay(roleSelect.Selected)

			err := s.userService.ChangeUserRole(s.currentUser.ID, user.ID, newRole)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Роль змінено!", s.window)
			s.loadUsers()
		},
		s.window,
	).Show()
}

// showResetPasswordDialog показує діалог скидання пароля.
func (s *UsersScreen) showResetPasswordDialog(user *domain.User) {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Новий пароль (мінімум 6 символів)")

	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Підтвердіть пароль")

	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Скидання пароля для: %s", user.FullName)),
		widget.NewLabel("Новий пароль:"),
		passwordEntry,
		widget.NewLabel("Підтвердження:"),
		confirmEntry,
	)

	dialog.NewCustomConfirm(
		"Скидання пароля",
		"Скинути",
		"Скасувати",
		form,
		func(reset bool) {
			if !reset {
				return
			}

			if passwordEntry.Text != confirmEntry.Text {
				dialog.ShowError(fmt.Errorf("паролі не співпадають"), s.window)
				return
			}

			err := s.userService.ResetUserPassword(
				s.currentUser.ID,
				user.ID,
				passwordEntry.Text,
			)

			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Пароль скинуто!", s.window)
		},
		s.window,
	).Show()
}

// confirmDeactivate підтверджує деактивацію користувача.
func (s *UsersScreen) confirmDeactivate(user *domain.User) {
	dialog.ShowConfirm(
		"Деактивація користувача",
		fmt.Sprintf("Деактивувати користувача %s?\n\nКористувач не зможе увійти в систему.", user.FullName),
		func(confirm bool) {
			if !confirm {
				return
			}

			err := s.userService.DeactivateUser(s.currentUser.ID, user.ID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Користувача деактивовано!", s.window)
			s.loadUsers()
		},
		s.window,
	)
}

// confirmActivate підтверджує активацію користувача.
func (s *UsersScreen) confirmActivate(user *domain.User) {
	dialog.ShowConfirm(
		"Активація користувача",
		fmt.Sprintf("Активувати користувача %s?", user.FullName),
		func(confirm bool) {
			if !confirm {
				return
			}

			err := s.userService.ActivateUser(s.currentUser.ID, user.ID)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation("Успіх", "Користувача активовано!", s.window)
			s.loadUsers()
		},
		s.window,
	)
}

// Допоміжні функції

func getRoleDisplayName(role string) string {
	switch role {
	case domain.RoleAdmin:
		return "Адміністратор"
	case domain.RoleHead:
		return "Голова ОСББ"
	case domain.RoleAccountant:
		return "Бухгалтер"
	default:
		return role
	}
}

func getRoleFromDisplay(display string) string {
	switch display {
	case "Адміністратор":
		return domain.RoleAdmin
	case "Голова ОСББ":
		return domain.RoleHead
	case "Бухгалтер":
		return domain.RoleAccountant
	default:
		return domain.RoleAccountant
	}
}

func getStatusText(isActive bool) string {
	if isActive {
		return "Активний"
	}
	return "Деактивований"
}
