// presentation/fyne/screens/owner_form_dialog.go
package dialogs

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/owner"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/presentation/fyne/auth"
)

// OwnerFormDialog - діалог для створення/редагування власника.
type OwnerFormDialog struct {
	window        fyne.Window
	ownerService  service.OwnerServiceInterface
	authManager   *auth.AuthManager
	existingOwner *owner.OwnerOutput // nil для створення, заповнено для редагування

	// UI елементи
	dialog                 dialog.Dialog
	firstNameEntry         *widget.Entry
	lastNameEntry          *widget.Entry
	middleNameEntry        *widget.Entry
	phoneEntry             *widget.Entry
	emailEntry             *widget.Entry
	taxNumberEntry         *widget.Entry
	passportSeriesEntry    *widget.Entry
	passportNumberEntry    *widget.Entry
	registeredAddressEntry *widget.Entry
	actualAddressEntry     *widget.Entry
	notesEntry             *widget.Entry
	saveButton             *widget.Button
	cancelButton           *widget.Button

	// Callbacks
	OnSuccess func()
}

// NewOwnerFormDialog створює новий діалог.
func NewOwnerFormDialog(
	window fyne.Window,
	ownerService service.OwnerServiceInterface,
	authManager *auth.AuthManager,
	existingOwner *owner.OwnerOutput,
) *OwnerFormDialog {
	d := &OwnerFormDialog{
		window:        window,
		ownerService:  ownerService,
		authManager:   authManager,
		existingOwner: existingOwner,
	}

	d.buildUI()
	return d
}

// buildUI створює інтерфейс діалогу.
func (d *OwnerFormDialog) buildUI() {
	// Валідатори
	taxNumberRegex := regexp.MustCompile(`^\d{10}$`)
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Основні поля
	d.firstNameEntry = widget.NewEntry()
	d.firstNameEntry.SetPlaceHolder("Ім'я")
	d.firstNameEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		return nil
	}

	d.lastNameEntry = widget.NewEntry()
	d.lastNameEntry.SetPlaceHolder("Прізвище")
	d.lastNameEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		return nil
	}

	d.middleNameEntry = widget.NewEntry()
	d.middleNameEntry.SetPlaceHolder("По батькові (необов'язково)")

	// Контактні дані
	d.phoneEntry = widget.NewEntry()
	d.phoneEntry.SetPlaceHolder("+380XXXXXXXXX")
	d.phoneEntry.Validator = func(s string) error {
		if s != "" && !phoneRegex.MatchString(s) {
			return fmt.Errorf("невірний формат телефону")
		}
		return nil
	}

	d.emailEntry = widget.NewEntry()
	d.emailEntry.SetPlaceHolder("email@example.com")
	d.emailEntry.Validator = func(s string) error {
		if s != "" && !emailRegex.MatchString(s) {
			return fmt.Errorf("невірний формат email")
		}
		return nil
	}

	// ІПН
	d.taxNumberEntry = widget.NewEntry()
	d.taxNumberEntry.SetPlaceHolder("10 цифр (необов'язково)")
	d.taxNumberEntry.Validator = func(s string) error {
		if s != "" && !taxNumberRegex.MatchString(s) {
			return fmt.Errorf("ІПН має містити 10 цифр")
		}
		return nil
	}

	// Паспортні дані
	d.passportSeriesEntry = widget.NewEntry()
	d.passportSeriesEntry.SetPlaceHolder("Серія")

	d.passportNumberEntry = widget.NewEntry()
	d.passportNumberEntry.SetPlaceHolder("Номер")

	// Адреси
	d.registeredAddressEntry = widget.NewEntry()
	d.registeredAddressEntry.SetPlaceHolder("Адреса реєстрації")

	d.actualAddressEntry = widget.NewEntry()
	d.actualAddressEntry.SetPlaceHolder("Фактична адреса")

	// Примітки
	d.notesEntry = widget.NewMultiLineEntry()
	d.notesEntry.SetPlaceHolder("Додаткова інформація...")
	d.notesEntry.SetMinRowsVisible(3)

	// Кнопки
	d.saveButton = widget.NewButton("Зберегти", d.handleSave)
	d.saveButton.Importance = widget.HighImportance

	d.cancelButton = widget.NewButton("Скасувати", func() {
		d.dialog.Hide()
	})

	// Заповнення даних при редагуванні
	if d.existingOwner != nil {
		d.firstNameEntry.SetText(d.existingOwner.FirstName)
		d.lastNameEntry.SetText(d.existingOwner.LastName)
		if d.existingOwner.MiddleName != nil {
			d.middleNameEntry.SetText(*d.existingOwner.MiddleName)
		}
		if d.existingOwner.Phone != nil {
			d.phoneEntry.SetText(*d.existingOwner.Phone)
		}
		if d.existingOwner.Email != nil {
			d.emailEntry.SetText(*d.existingOwner.Email)
		}
		if d.existingOwner.TaxNumber != nil {
			d.taxNumberEntry.SetText(*d.existingOwner.TaxNumber)
		}
		if d.existingOwner.PassportSeries != nil {
			d.passportSeriesEntry.SetText(*d.existingOwner.PassportSeries)
		}
		if d.existingOwner.PassportNumber != nil {
			d.passportNumberEntry.SetText(*d.existingOwner.PassportNumber)
		}
		if d.existingOwner.RegisteredAddress != nil {
			d.registeredAddressEntry.SetText(*d.existingOwner.RegisteredAddress)
		}
		if d.existingOwner.ActualAddress != nil {
			d.actualAddressEntry.SetText(*d.existingOwner.ActualAddress)
		}
		if d.existingOwner.Notes != nil {
			d.notesEntry.SetText(*d.existingOwner.Notes)
		}
	}
}

// Show показує діалог.
func (d *OwnerFormDialog) Show() {
	title := "Додати власника"
	if d.existingOwner != nil {
		title = "Редагувати власника"
	}

	// Форма у дві колонки
	leftColumn := container.NewVBox(
		widget.NewLabel("Основна інформація:"),
		widget.NewForm(
			widget.NewFormItem("* Ім'я", d.firstNameEntry),
			widget.NewFormItem("* Прізвище", d.lastNameEntry),
			widget.NewFormItem("По батькові", d.middleNameEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Контактні дані:"),
		widget.NewForm(
			widget.NewFormItem("Телефон", d.phoneEntry),
			widget.NewFormItem("Email", d.emailEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Документи:"),
		widget.NewForm(
			widget.NewFormItem("ІПН", d.taxNumberEntry),
		),
	)

	rightColumn := container.NewVBox(
		widget.NewLabel("Паспортні дані:"),
		widget.NewForm(
			widget.NewFormItem("Серія", d.passportSeriesEntry),
			widget.NewFormItem("Номер", d.passportNumberEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Адреси:"),
		widget.NewForm(
			widget.NewFormItem("Реєстрації", d.registeredAddressEntry),
			widget.NewFormItem("Фактична", d.actualAddressEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Примітки:"),
		d.notesEntry,
	)

	formContainer := container.NewGridWithColumns(2, leftColumn, rightColumn)

	// Кнопки
	buttons := container.NewGridWithColumns(2, d.cancelButton, d.saveButton)

	// Інфо
	infoLabel := widget.NewLabel("* - обов'язкові поля")
	infoLabel.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewBorder(
		nil,
		container.NewVBox(infoLabel, buttons),
		nil,
		nil,
		container.NewScroll(formContainer),
	)

	d.dialog = dialog.NewCustom(title, "", content, d.window)
	d.dialog.Resize(fyne.NewSize(900, 600))
	d.dialog.Show()
}

// handleSave обробляє збереження даних.
func (d *OwnerFormDialog) handleSave() {
	// Валідація обов'язкових полів
	if err := d.firstNameEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Ім'я: %v", err), d.window)
		return
	}
	if err := d.lastNameEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Прізвище: %v", err), d.window)
		return
	}

	// Валідація опціональних полів
	if d.phoneEntry.Text != "" {
		if err := d.phoneEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Телефон: %v", err), d.window)
			return
		}
	}
	if d.emailEntry.Text != "" {
		if err := d.emailEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Email: %v", err), d.window)
			return
		}
	}
	if d.taxNumberEntry.Text != "" {
		if err := d.taxNumberEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("ІПН: %v", err), d.window)
			return
		}
	}

	// Перевірка - хоча б один контакт має бути вказаний
	if d.phoneEntry.Text == "" && d.emailEntry.Text == "" {
		dialog.ShowConfirm(
			"Увага",
			"Не вказано контактних даних (телефон або email). Продовжити?",
			func(confirmed bool) {
				if confirmed {
					d.performSave()
				}
			},
			d.window,
		)
		return
	}

	d.performSave()
}

// performSave виконує збереження.
func (d *OwnerFormDialog) performSave() {
	// Відключаємо кнопку
	d.saveButton.Disable()
	defer d.saveButton.Enable()

	// Progress
	progress := dialog.NewCustomWithoutButtons(
		"Збереження...",
		widget.NewProgressBarInfinite(),
		d.window,
	)
	progress.Show()
	defer progress.Hide()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Підготовка даних
	var middleName, phone, email, taxNumber *string
	var passportSeries, passportNumber *string
	var registeredAddress, actualAddress, notes *string

	if d.middleNameEntry.Text != "" {
		middleName = &d.middleNameEntry.Text
	}
	if d.phoneEntry.Text != "" {
		phone = &d.phoneEntry.Text
	}
	if d.emailEntry.Text != "" {
		email = &d.emailEntry.Text
	}
	if d.taxNumberEntry.Text != "" {
		taxNumber = &d.taxNumberEntry.Text
	}
	if d.passportSeriesEntry.Text != "" {
		passportSeries = &d.passportSeriesEntry.Text
	}
	if d.passportNumberEntry.Text != "" {
		passportNumber = &d.passportNumberEntry.Text
	}
	if d.registeredAddressEntry.Text != "" {
		registeredAddress = &d.registeredAddressEntry.Text
	}
	if d.actualAddressEntry.Text != "" {
		actualAddress = &d.actualAddressEntry.Text
	}
	if d.notesEntry.Text != "" {
		notes = &d.notesEntry.Text
	}

	var err error
	var result *owner.OwnerOutput

	if d.existingOwner == nil {
		// Створення
		result, err = d.ownerService.Create(ctx, owner.CreateOwnerInput{
			CurrentUserID:     d.authManager.GetCurrentUserID(),
			FirstName:         d.firstNameEntry.Text,
			LastName:          d.lastNameEntry.Text,
			MiddleName:        middleName,
			Phone:             phone,
			Email:             email,
			TaxNumber:         taxNumber,
			PassportSeries:    passportSeries,
			PassportNumber:    passportNumber,
			RegisteredAddress: registeredAddress,
			ActualAddress:     actualAddress,
			Notes:             notes,
		})
	} else {
		// Оновлення
		result, err = d.ownerService.Update(ctx, owner.UpdateOwnerInput{
			CurrentUserID:     d.authManager.GetCurrentUserID(),
			OwnerID:           d.existingOwner.ID,
			FirstName:         d.firstNameEntry.Text,
			LastName:          d.lastNameEntry.Text,
			MiddleName:        middleName,
			Phone:             phone,
			Email:             email,
			TaxNumber:         taxNumber,
			PassportSeries:    passportSeries,
			PassportNumber:    passportNumber,
			RegisteredAddress: registeredAddress,
			ActualAddress:     actualAddress,
			Notes:             notes,
		})
	}

	if err != nil {
		d.handleError(err)
		return
	}

	// Успіх
	d.dialog.Hide()

	successMsg := "Власника успішно створено!"
	if d.existingOwner != nil {
		successMsg = "Дані власника успішно оновлено!"
	}

	dialog.ShowInformation("Успіх",
		fmt.Sprintf("%s\n\n%s", successMsg, result.FullName),
		d.window)

	// Викликаємо callback
	if d.OnSuccess != nil {
		d.OnSuccess()
	}
}

// handleError обробляє помилки збереження.
func (d *OwnerFormDialog) handleError(err error) {
	var message string

	// Перевіряємо тип помилки
	var domainErr *domainErrors.DomainError
	if domainErrors.As(err, &domainErr) {
		if domainErr.Code == domainErrors.CodeDuplicateEntry {
			if field, ok := domainErr.Details["field"].(string); ok && field == "tax_number" {
				message = "Власник з таким ІПН вже існує"
				d.taxNumberEntry.FocusGained()
			} else {
				message = "Такий запис вже існує"
			}
		} else if domainErr.Code == domainErrors.CodeValidationFailed {
			message = fmt.Sprintf("Помилка валідації: %s", domainErr.Message)
		} else if domainErr.Code == domainErrors.CodePermissionDenied {
			message = "Недостатньо прав для виконання операції"
		} else {
			message = domainErr.Message
		}
	} else {
		message = fmt.Sprintf("Помилка збереження: %v", err)
	}

	dialog.ShowError(fmt.Errorf("%s", message), d.window)
}
