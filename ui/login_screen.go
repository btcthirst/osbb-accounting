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

// LoginScreen представляє екран входу в систему.
// Цей екран відокремлений від бізнес-логіки і взаємодіє тільки через AuthService.
type LoginScreen struct {
	// authService - сервіс для аутентифікації
	authService *service.AuthService

	// window - посилання на головне вікно додатку
	window fyne.Window

	// onLoginSuccess - callback, який викликається після успішного входу
	onLoginSuccess func(*domain.User)

	// UI елементи
	usernameEntry *widget.Entry
	passwordEntry *widget.Entry
	loginButton   *widget.Button
	statusLabel   *widget.Label
}

// NewLoginScreen створює новий екран входу.
//
// Параметри:
//   - authService: сервіс аутентифікації
//   - window: головне вікно додатку
//   - onLoginSuccess: callback для обробки успішного входу
//
// Повертає:
//   - *LoginScreen: новий екземпляр екрану входу
func NewLoginScreen(authService *service.AuthService, window fyne.Window, onLoginSuccess func(*domain.User)) *LoginScreen {
	screen := &LoginScreen{
		authService:    authService,
		window:         window,
		onLoginSuccess: onLoginSuccess,
	}

	screen.initUI()
	return screen
}

// initUI ініціалізує UI елементи екрану входу.
func (s *LoginScreen) initUI() {
	// Поле для введення імені користувача
	s.usernameEntry = widget.NewEntry()
	s.usernameEntry.SetPlaceHolder("Введіть ім'я користувача")
	s.usernameEntry.OnSubmitted = func(string) {
		// При натисканні Enter переходимо до поля пароля
		s.window.Canvas().Focus(s.passwordEntry)
	}

	// Поле для введення пароля (приховане)
	s.passwordEntry = widget.NewPasswordEntry()
	s.passwordEntry.SetPlaceHolder("Введіть пароль")
	s.passwordEntry.OnSubmitted = func(string) {
		// При натисканні Enter виконуємо вхід
		s.handleLogin()
	}

	// Кнопка входу
	s.loginButton = widget.NewButton("Увійти", s.handleLogin)
	s.loginButton.Importance = widget.HighImportance

	// Мітка статусу (для відображення помилок)
	s.statusLabel = widget.NewLabel("")
	s.statusLabel.Wrapping = fyne.TextWrapWord
	s.statusLabel.Hide()
}

// handleLogin обробляє спробу входу користувача.
func (s *LoginScreen) handleLogin() {
	username := s.usernameEntry.Text
	password := s.passwordEntry.Text

	// Очищаємо попередні повідомлення про помилки
	s.statusLabel.SetText("")
	s.statusLabel.Hide()

	// Валідація введених даних на клієнтській стороні
	if err := s.authService.ValidateCredentials(username, password); err != nil {
		s.showError(err.Error())
		return
	}

	// Виконуємо аутентифікацію (синхронно для простоти в MVP)
	user, err := s.authService.Authenticate(username, password)

	if err != nil {
		// Помилка аутентифікації
		s.showError("Невірне ім'я користувача або пароль")
		return
	}

	// Успішний вхід
	s.clearFields()

	// Викликаємо callback з даними користувача
	if s.onLoginSuccess != nil {
		s.onLoginSuccess(user)
	}
}

// showError відображає повідомлення про помилку.
func (s *LoginScreen) showError(message string) {
	s.statusLabel.SetText(message)
	s.statusLabel.Show()
	s.statusLabel.Importance = widget.DangerImportance
}

// clearFields очищає поля введення.
func (s *LoginScreen) clearFields() {
	s.usernameEntry.SetText("")
	s.passwordEntry.SetText("")
}

// Render повертає Fyne контейнер з UI екрану входу.
func (s *LoginScreen) Render() fyne.CanvasObject {
	// Заголовок
	title := widget.NewLabel("ОСББ Облік")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("Система обліку платежів та боргів")
	subtitle.Alignment = fyne.TextAlignCenter

	// Іконка входу
	loginIcon := widget.NewIcon(theme.LoginIcon())

	// Форма входу
	form := container.NewVBox(
		container.NewCenter(loginIcon),
		layout.NewSpacer(),

		title,
		subtitle,

		layout.NewSpacer(),

		widget.NewLabel("Ім'я користувача:"),
		s.usernameEntry,

		widget.NewLabel("Пароль:"),
		s.passwordEntry,

		s.statusLabel,

		layout.NewSpacer(),
		s.loginButton,

		layout.NewSpacer(),

		// Підказка для першого входу
		container.NewCenter(widget.NewLabel("За замовчуванням: admin / admin123")),
	)

	// Обмежуємо ширину форми
	formContainer := container.NewBorder(
		nil, nil,
		layout.NewSpacer(),
		layout.NewSpacer(),
		container.NewPadded(
			container.NewMax(
				container.NewPadded(form),
			),
		),
	)

	return formContainer
}

// SetFocus встановлює фокус на поле username після відображення екрану.
// Викликається після того, як контент вже доданий до canvas.
func (s *LoginScreen) SetFocus() {
	// Використовуємо невелику затримку для гарантії, що елемент вже в canvas
	s.window.Canvas().Focus(s.usernameEntry)
}

// ShowPasswordChangeDialog відображає діалог зміни пароля.
// Викликається після першого входу з дефолтним паролем.
func (s *LoginScreen) ShowPasswordChangeDialog(user *domain.User) {
	oldPasswordEntry := widget.NewPasswordEntry()
	oldPasswordEntry.SetPlaceHolder("Поточний пароль")

	newPasswordEntry := widget.NewPasswordEntry()
	newPasswordEntry.SetPlaceHolder("Новий пароль (мін. 6 символів)")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Підтвердіть новий пароль")

	form := container.NewVBox(
		widget.NewLabel("Змініть пароль для безпеки вашого облікового запису"),
		layout.NewSpacer(),
		widget.NewLabel("Поточний пароль:"),
		oldPasswordEntry,
		widget.NewLabel("Новий пароль:"),
		newPasswordEntry,
		widget.NewLabel("Підтвердження:"),
		confirmPasswordEntry,
	)

	d := dialog.NewCustomConfirm(
		"Зміна пароля",
		"Змінити",
		"Пізніше",
		form,
		func(change bool) {
			if !change {
				return
			}

			oldPass := oldPasswordEntry.Text
			newPass := newPasswordEntry.Text
			confirmPass := confirmPasswordEntry.Text

			// Валідація
			if newPass != confirmPass {
				dialog.ShowError(fmt.Errorf("паролі не співпадають"), s.window)
				return
			}

			if len(newPass) < 6 {
				dialog.ShowError(fmt.Errorf("пароль має бути не менше 6 символів"), s.window)
				return
			}

			// Змінюємо пароль через сервіс
			err := s.authService.ChangePassword(user.ID, oldPass, newPass)
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}

			dialog.ShowInformation(
				"Успіх",
				"Пароль успішно змінено!",
				s.window,
			)
		},
		s.window,
	)

	d.Resize(fyne.NewSize(400, 350))
	d.Show()
}
