// presentation/fyne/screens/register_screen.go
package auth

import (
	"context"
	"fmt"
	"regexp"
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

// RegisterScreen представляє екран реєстрації.
type RegisterScreen struct {
	window      fyne.Window
	authService service.AuthServiceInterface

	// UI елементи
	usernameEntry    *widget.Entry
	emailEntry       *widget.Entry
	passwordEntry    *widget.Entry
	confirmPassEntry *widget.Entry
	firstNameEntry   *widget.Entry
	lastNameEntry    *widget.Entry
	middleNameEntry  *widget.Entry
	phoneEntry       *widget.Entry
	registerButton   *widget.Button
	backButton       *widget.Button

	// Callbacks
	onRegisterSuccess func()
	onBackClick       func()
}

// NewRegisterScreen створює новий екран реєстрації.
func NewRegisterScreen(window fyne.Window, authService service.AuthServiceInterface) *RegisterScreen {
	screen := &RegisterScreen{
		window:      window,
		authService: authService,
	}

	screen.buildUI()
	return screen
}

// buildUI створює інтерфейс екрану.
func (s *RegisterScreen) buildUI() {
	// Валідатори
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Username
	s.usernameEntry = widget.NewEntry()
	s.usernameEntry.SetPlaceHolder("Логін (мін. 3 символи)")
	s.usernameEntry.Validator = func(text string) error {
		if len(text) < 3 {
			return fmt.Errorf("мінімум 3 символи")
		}
		if len(text) > 50 {
			return fmt.Errorf("максимум 50 символів")
		}
		return nil
	}

	// Email
	s.emailEntry = widget.NewEntry()
	s.emailEntry.SetPlaceHolder("Email")
	s.emailEntry.Validator = func(text string) error {
		if !emailRegex.MatchString(text) {
			return fmt.Errorf("невірний формат email")
		}
		return nil
	}

	// Password
	s.passwordEntry = widget.NewPasswordEntry()
	s.passwordEntry.SetPlaceHolder("Пароль (мін. 8 символів)")
	s.passwordEntry.Validator = func(text string) error {
		if len(text) < 8 {
			return fmt.Errorf("мінімум 8 символів")
		}
		return nil
	}

	// Confirm Password
	s.confirmPassEntry = widget.NewPasswordEntry()
	s.confirmPassEntry.SetPlaceHolder("Підтвердження пароля")
	s.confirmPassEntry.Validator = func(text string) error {
		if text != s.passwordEntry.Text {
			return fmt.Errorf("паролі не співпадають")
		}
		return nil
	}

	// FirstName
	s.firstNameEntry = widget.NewEntry()
	s.firstNameEntry.SetPlaceHolder("Ім'я")
	s.firstNameEntry.Validator = func(text string) error {
		if text == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		return nil
	}

	// LastName
	s.lastNameEntry = widget.NewEntry()
	s.lastNameEntry.SetPlaceHolder("Прізвище")
	s.lastNameEntry.Validator = func(text string) error {
		if text == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		return nil
	}

	// MiddleName (optional)
	s.middleNameEntry = widget.NewEntry()
	s.middleNameEntry.SetPlaceHolder("По батькові (необов'язково)")

	// Phone (optional)
	s.phoneEntry = widget.NewEntry()
	s.phoneEntry.SetPlaceHolder("Телефон (необов'язково)")

	// Register button
	s.registerButton = widget.NewButton("Зареєструватися", s.handleRegister)
	s.registerButton.Importance = widget.HighImportance

	// Back button
	s.backButton = widget.NewButton("Назад", func() {
		if s.onBackClick != nil {
			s.onBackClick()
		}
	})
}

// Render повертає контейнер з UI екрану.
func (s *RegisterScreen) Render() fyne.CanvasObject {
	// Заголовок
	title := widget.NewLabelWithStyle(
		"Реєстрація",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	// title.TextSize = 24 ////////////////////////////////////////////////////////////////////////////////////////////////

	subtitle := widget.NewLabel("Створіть новий акаунт")
	subtitle.Alignment = fyne.TextAlignCenter

	// Форма реєстрації (дві колонки)
	leftColumn := widget.NewForm(
		widget.NewFormItem("* Логін", s.usernameEntry),
		widget.NewFormItem("* Email", s.emailEntry),
		widget.NewFormItem("* Пароль", s.passwordEntry),
		widget.NewFormItem("* Підтвердження", s.confirmPassEntry),
	)

	rightColumn := widget.NewForm(
		widget.NewFormItem("* Ім'я", s.firstNameEntry),
		widget.NewFormItem("* Прізвище", s.lastNameEntry),
		widget.NewFormItem("По батькові", s.middleNameEntry),
		widget.NewFormItem("Телефон", s.phoneEntry),
	)

	formContainer := container.NewGridWithColumns(2,
		leftColumn,
		rightColumn,
	)

	// Кнопки
	buttonsContainer := container.NewGridWithColumns(2,
		s.backButton,
		s.registerButton,
	)

	// Інформаційний текст
	infoLabel := widget.NewLabel("* - обов'язкові поля")
	infoLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Головний контейнер
	content := container.NewVBox(
		title,
		subtitle,
		layout.NewSpacer(),
		formContainer,
		layout.NewSpacer(),
		infoLabel,
		buttonsContainer,
	)

	// Scrollable контейнер для довгої форми
	return container.NewPadded(
		container.NewVScroll(content),
	)
}

// handleRegister обробляє реєстрацію.
func (s *RegisterScreen) handleRegister() {
	// Валідація всіх полів
	fields := []struct {
		entry *widget.Entry
		name  string
	}{
		{s.usernameEntry, "Логін"},
		{s.emailEntry, "Email"},
		{s.passwordEntry, "Пароль"},
		{s.confirmPassEntry, "Підтвердження пароля"},
		{s.firstNameEntry, "Ім'я"},
		{s.lastNameEntry, "Прізвище"},
	}

	for _, field := range fields {
		if err := field.entry.Validate(); err != nil {
			s.showError("Помилка валідації", fmt.Sprintf("%s: %s", field.name, err.Error()))
			field.entry.FocusGained()
			return
		}
		if field.entry.Text == "" {
			s.showError("Помилка", fmt.Sprintf("%s є обов'язковим полем", field.name))
			field.entry.FocusGained()
			return
		}
	}

	// Додаткова перевірка паролів
	if s.passwordEntry.Text != s.confirmPassEntry.Text {
		s.showError("Помилка", "Паролі не співпадають")
		s.confirmPassEntry.FocusGained()
		return
	}

	// Відключаємо кнопку
	s.registerButton.Disable()
	defer s.registerButton.Enable()

	// Progress dialog
	progress := dialog.NewCustomWithoutButtons(
		"Реєстрація...",
		container.NewVBox(
			widget.NewProgressBarInfinite(),
			widget.NewLabel("Створення облікового запису"),
		),
		s.window,
	)
	progress.Show()
	defer progress.Hide()

	// Підготовка даних
	var middleName, phone *string
	if s.middleNameEntry.Text != "" {
		middleName = &s.middleNameEntry.Text
	}
	if s.phoneEntry.Text != "" {
		phone = &s.phoneEntry.Text
	}

	// Виконуємо реєстрацію
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := s.authService.Register(ctx, auth.RegisterUserInput{
		Username:   s.usernameEntry.Text,
		Email:      s.emailEntry.Text,
		Password:   s.passwordEntry.Text,
		FirstName:  s.firstNameEntry.Text,
		LastName:   s.lastNameEntry.Text,
		MiddleName: middleName,
		Phone:      phone,
		RoleName:   "viewer", // Роль за замовчуванням
	})

	if err != nil {
		s.handleRegisterError(err)
		return
	}

	// Успішна реєстрація
	s.handleRegisterSuccess(output)
}

// handleRegisterError обробляє помилки реєстрації.
func (s *RegisterScreen) handleRegisterError(err error) {
	var message string

	// Перевіряємо тип помилки
	var domainErr *domainErrors.DomainError
	if domainErrors.As(err, &domainErr) {
		if domainErr.Code == domainErrors.CodeDuplicateEntry {
			if field, ok := domainErr.Details["field"].(string); ok {
				switch field {
				case "username":
					message = "Користувач з таким логіном вже існує"
					s.usernameEntry.FocusGained()
				case "email":
					message = "Користувач з таким email вже існує"
					s.emailEntry.FocusGained()
				default:
					message = "Такий запис вже існує"
				}
			} else {
				message = "Такий запис вже існує"
			}
		} else if domainErr.Code == domainErrors.CodeValidationFailed {
			message = fmt.Sprintf("Помилка валідації: %s", domainErr.Message)
		} else {
			message = domainErr.Message
		}
	} else {
		message = fmt.Sprintf("Помилка реєстрації: %s", err.Error())
	}

	s.showError("Помилка реєстрації", message)
}

// handleRegisterSuccess обробляє успішну реєстрацію.
func (s *RegisterScreen) handleRegisterSuccess(output *auth.RegisterUserOutput) {
	successMsg := fmt.Sprintf(
		"Вітаємо, %s %s!\n\n"+
			"Ваш акаунт успішно створено.\n"+
			"Логін: %s\n"+
			"Email: %s\n\n"+
			"Тепер ви можете увійти в систему.",
		output.FirstName,
		output.LastName,
		output.Username,
		output.Email,
	)

	dialog.ShowInformation(
		"Реєстрація завершена",
		successMsg,
		s.window,
	)

	// Очищаємо форму
	s.Clear()

	// Викликаємо callback
	if s.onRegisterSuccess != nil {
		s.onRegisterSuccess()
	}
}

// showError показує діалог помилки.
func (s *RegisterScreen) showError(title, message string) {
	dialog.ShowError(fmt.Errorf("title: %s\n message: %s", title, message), s.window)
}

// OnRegisterSuccess встановлює callback для успішної реєстрації.
func (s *RegisterScreen) OnRegisterSuccess(callback func()) {
	s.onRegisterSuccess = callback
}

// OnBackClick встановлює callback для кнопки "Назад".
func (s *RegisterScreen) OnBackClick(callback func()) {
	s.onBackClick = callback
}

// Clear очищує всі поля форми.
func (s *RegisterScreen) Clear() {
	s.usernameEntry.SetText("")
	s.emailEntry.SetText("")
	s.passwordEntry.SetText("")
	s.confirmPassEntry.SetText("")
	s.firstNameEntry.SetText("")
	s.lastNameEntry.SetText("")
	s.middleNameEntry.SetText("")
	s.phoneEntry.SetText("")
}

// Focus встановлює фокус на перше поле.
func (s *RegisterScreen) Focus() {
	s.usernameEntry.FocusGained()
}
