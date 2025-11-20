// presentation/fyne/screens/login_screen.go
package auth

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/auth"
	domainErrors "osbb-accounting/domain/errors"
)

// LoginScreen представляє екран входу в систему.
type LoginScreen struct {
	window      fyne.Window
	authService *service.AuthService

	// UI елементи
	usernameEntry *widget.Entry
	passwordEntry *widget.Entry
	rememberCheck *widget.Check
	loginButton   *widget.Button
	registerLink  *widget.Hyperlink

	// Callbacks
	onLoginSuccess  func(sessionToken string, userID int64)
	onRegisterClick func()
}

// NewLoginScreen створює новий екран входу.
func NewLoginScreen(window fyne.Window, authService *service.AuthService) *LoginScreen {
	screen := &LoginScreen{
		window:      window,
		authService: authService,
	}

	screen.buildUI()
	return screen
}

// buildUI створює інтерфейс екрану.
func (s *LoginScreen) buildUI() {
	// Username поле
	s.usernameEntry = widget.NewEntry()
	s.usernameEntry.SetPlaceHolder("Логін або Email")
	s.usernameEntry.Validator = func(text string) error {
		if len(text) < 3 {
			return fmt.Errorf("мінімум 3 символи")
		}
		return nil
	}

	// Password поле
	s.passwordEntry = widget.NewPasswordEntry()
	s.passwordEntry.SetPlaceHolder("Пароль")
	s.passwordEntry.Validator = func(text string) error {
		if len(text) < 8 {
			return fmt.Errorf("мінімум 8 символів")
		}
		return nil
	}

	// Remember me checkbox
	s.rememberCheck = widget.NewCheck("Запам'ятати мене", nil)

	// Login button
	s.loginButton = widget.NewButton("Увійти", s.handleLogin)
	s.loginButton.Importance = widget.HighImportance

	// Register link
	s.registerLink = widget.NewHyperlink("Зареєструватися", nil)
	s.registerLink.OnTapped = func() {
		if s.onRegisterClick != nil {
			s.onRegisterClick()
		}
	}

	// Enter key для входу
	s.passwordEntry.OnSubmitted = func(string) {
		s.handleLogin()
	}
}

// Render повертає контейнер з UI екрану.
func (s *LoginScreen) Render() fyne.CanvasObject {
	// Логотип/Заголовок
	title := widget.NewLabelWithStyle(
		"ОСББ Управління",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	// title.TextSize = 24 - title.TextSize undefined

	subtitle := widget.NewLabel("Увійдіть до системи")
	subtitle.Alignment = fyne.TextAlignCenter

	// Форма входу
	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Логін", s.usernameEntry),
			widget.NewFormItem("Пароль", s.passwordEntry),
		),
		s.rememberCheck,
		layout.NewSpacer(),
		s.loginButton,
	)

	// Посилання на реєстрацію
	registerContainer := container.NewHBox(
		layout.NewSpacer(),
		widget.NewLabel("Немає акаунту?"),
		s.registerLink,
		layout.NewSpacer(),
	)

	// Версія програми
	versionLabel := widget.NewLabel("v1.0.0")
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}
	versionLabel.Alignment = fyne.TextAlignCenter

	// Головний контейнер
	content := container.NewVBox(
		layout.NewSpacer(),
		title,
		subtitle,
		layout.NewSpacer(),
		form,
		layout.NewSpacer(),
		registerContainer,
		layout.NewSpacer(),
		versionLabel,
	)

	// Padding навколо форми
	return container.NewPadded(
		container.NewStack(
			container.NewCenter(
				container.New(
					layout.NewStackLayout(),
					widget.NewCard("", "", content),
				),
			),
		),
	)
}

// handleLogin обробляє спробу входу.
func (s *LoginScreen) handleLogin() {
	// Валідація полів
	if err := s.usernameEntry.Validate(); err != nil {
		s.showError("Помилка", fmt.Sprintf("Логін: %s", err.Error()))
		s.usernameEntry.FocusGained()
		return
	}

	if err := s.passwordEntry.Validate(); err != nil {
		s.showError("Помилка", fmt.Sprintf("Пароль: %s", err.Error()))
		s.passwordEntry.FocusGained()
		return
	}

	username := s.usernameEntry.Text
	password := s.passwordEntry.Text

	if username == "" || password == "" {
		s.showError("Помилка", "Заповніть усі поля")
		return
	}

	// Відключаємо кнопку під час входу
	s.loginButton.Disable()
	defer s.loginButton.Enable()

	// Показуємо progress dialog
	progress := dialog.NewCustomWithoutButtons(
		"Вхід...",
		container.NewVBox(
			widget.NewProgressBarInfinite(),
			widget.NewLabel("Перевірка облікових даних"),
		),
		s.window,
	)
	progress.Show()
	defer progress.Hide()

	// Виконуємо вхід
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sessionDuration := 8 * time.Hour
	if s.rememberCheck.Checked {
		sessionDuration = 30 * 24 * time.Hour // 30 днів
	}

	output, err := s.authService.Login(ctx, auth.LoginUserInput{
		UsernameOrEmail: username,
		Password:        password,
		SessionDuration: sessionDuration,
	})

	if err != nil {
		s.handleLoginError(err)
		return
	}

	// Успішний вхід
	s.handleLoginSuccess(output)
}

// handleLoginError обробляє помилки входу.
func (s *LoginScreen) handleLoginError(err error) {
	var message string

	// Перевіряємо тип помилки
	if domainErrors.Is(err, domainErrors.ErrInvalidCredentials) {
		message = "Невірний логін або пароль"
	} else if domainErrors.Is(err, domainErrors.ErrUserInactive) {
		message = "Ваш акаунт деактивований. Зверніться до адміністратора."
	} else if domainErrors.Is(err, domainErrors.ErrUserDeleted) {
		message = "Акаунт видалено"
	} else {
		message = fmt.Sprintf("Помилка входу: %s", err.Error())
	}

	s.showError("Помилка входу", message)

	// Очищуємо пароль
	s.passwordEntry.SetText("")
	s.passwordEntry.FocusGained()
}

// handleLoginSuccess обробляє успішний вхід.
func (s *LoginScreen) handleLoginSuccess(output *auth.LoginUserOutput) {
	// Показуємо повідомлення
	welcomeMsg := fmt.Sprintf(
		"Вітаємо, %s!\n\nРолі: %v",
		output.User.FullName,
		output.Roles,
	)

	dialog.ShowInformation(
		"Успішний вхід",
		welcomeMsg,
		s.window,
	)

	// Викликаємо callback
	if s.onLoginSuccess != nil {
		s.onLoginSuccess(output.SessionToken, output.User.ID)
	}
}

// showError показує діалог помилки.
func (s *LoginScreen) showError(title, message string) {
	dialog.ShowError(fmt.Errorf("title: %s\n message: %s", title, message), s.window)
}

// OnLoginSuccess встановлює callback для успішного входу.
func (s *LoginScreen) OnLoginSuccess(callback func(sessionToken string, userID int64)) {
	s.onLoginSuccess = callback
}

// OnRegisterClick встановлює callback для кліку на реєстрацію.
func (s *LoginScreen) OnRegisterClick(callback func()) {
	s.onRegisterClick = callback
}

// Clear очищує форму.
func (s *LoginScreen) Clear() {
	s.usernameEntry.SetText("")
	s.passwordEntry.SetText("")
	s.rememberCheck.SetChecked(false)
}

// Focus встановлює фокус на поле логіну.
func (s *LoginScreen) Focus() {
	s.window.Canvas().Focus(s.usernameEntry)
	//s.usernameEntry.FocusGained()
}
