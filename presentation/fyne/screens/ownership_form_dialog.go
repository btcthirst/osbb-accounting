// presentation/fyne/screens/ownership_form_dialog.go
package screens

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/apartment"
	"osbb-accounting/application/usecase/owner"
	"osbb-accounting/application/usecase/ownership"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/presentation/fyne/auth"
)

// OwnershipFormDialog - діалог для створення/редагування частки власності.
type OwnershipFormDialog struct {
	window           fyne.Window
	ownershipService *service.OwnershipService
	ownerService     *service.OwnerService
	apartmentService *service.ApartmentService
	authManager      *auth.AuthManager
	existingShare    *ownership.OwnershipShareDetailsOutput

	// UI елементи
	dialog              dialog.Dialog
	ownerSelect         *widget.Select
	apartmentSelect     *widget.Select
	numeratorEntry      *widget.Entry
	denominatorEntry    *widget.Entry
	sharePercentLabel   *widget.Label
	remainingLabel      *widget.Label
	ownershipTypeSelect *widget.Select
	startDateEntry      *widget.Entry
	endDateEntry        *widget.Entry
	docTypeSelect       *widget.Select
	docNumberEntry      *widget.Entry
	docDateEntry        *widget.Entry
	notesEntry          *widget.Entry
	saveButton          *widget.Button
	cancelButton        *widget.Button

	// Дані
	owners              []*owner.OwnerOutput
	apartments          []*apartment.ApartmentOutput
	selectedOwnerID     int64
	selectedApartmentID int64

	// Callbacks
	OnSuccess func()
}

// NewOwnershipFormDialog створює новий діалог.
func NewOwnershipFormDialog(
	window fyne.Window,
	ownershipService *service.OwnershipService,
	ownerService *service.OwnerService,
	apartmentService *service.ApartmentService,
	authManager *auth.AuthManager,
	existingShare *ownership.OwnershipShareDetailsOutput,
) *OwnershipFormDialog {
	d := &OwnershipFormDialog{
		window:           window,
		ownershipService: ownershipService,
		ownerService:     ownerService,
		apartmentService: apartmentService,
		authManager:      authManager,
		existingShare:    existingShare,
	}

	d.loadData()
	d.buildUI()
	return d
}

// loadData завантажує списки власників та квартир.
func (d *OwnershipFormDialog) loadData() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завантажуємо власників
	ownersOutput, err := d.ownerService.List(ctx, owner.ListOwnersInput{
		CurrentUserID: d.authManager.GetCurrentUserID(),
		Limit:         1000,
		OrderBy:       "name",
	})
	if err == nil {
		d.owners = ownersOutput.Owners
	}

	// Завантажуємо квартири
	apartmentsOutput, err := d.apartmentService.List(ctx, apartment.ListApartmentsInput{
		CurrentUserID: d.authManager.GetCurrentUserID(),
		Limit:         1000,
		OrderBy:       "number",
	})
	if err == nil {
		d.apartments = apartmentsOutput.Apartments
	}
}

// buildUI створює інтерфейс діалогу.
func (d *OwnershipFormDialog) buildUI() {
	// Вибір власника
	ownerNames := make([]string, len(d.owners))
	for i, o := range d.owners {
		taxInfo := "немає ІПН"
		if o.TaxNumber != nil {
			taxInfo = *o.TaxNumber
		}
		ownerNames[i] = fmt.Sprintf("%s (ІПН: %s)", o.FullName, taxInfo)
	}
	d.ownerSelect = widget.NewSelect(ownerNames, func(selected string) {
		for i, name := range ownerNames {
			if name == selected {
				d.selectedOwnerID = d.owners[i].ID
				d.calculateRemaining()
				break
			}
		}
	})
	d.ownerSelect.PlaceHolder = "Оберіть власника"

	// Вибір квартири
	apartmentNames := make([]string, len(d.apartments))
	for i, a := range d.apartments {
		apartmentNames[i] = fmt.Sprintf("Кв. %s (%.1f м², %d пов.)",
			a.ApartmentNumber, a.AreaTotal, a.Floor)
	}
	d.apartmentSelect = widget.NewSelect(apartmentNames, func(selected string) {
		for i, name := range apartmentNames {
			if name == selected {
				d.selectedApartmentID = d.apartments[i].ID
				d.calculateRemaining()
				break
			}
		}
	})
	d.apartmentSelect.PlaceHolder = "Оберіть квартиру"

	// Чисельник
	d.numeratorEntry = widget.NewEntry()
	d.numeratorEntry.SetPlaceHolder("1")
	d.numeratorEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		num, err := strconv.Atoi(s)
		if err != nil || num <= 0 {
			return fmt.Errorf("має бути числом > 0")
		}
		return nil
	}
	d.numeratorEntry.OnChanged = func(string) {
		d.updateSharePercentage()
	}

	// Знаменник
	d.denominatorEntry = widget.NewEntry()
	d.denominatorEntry.SetPlaceHolder("1")
	d.denominatorEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		denom, err := strconv.Atoi(s)
		if err != nil || denom <= 0 {
			return fmt.Errorf("має бути числом > 0")
		}
		if d.numeratorEntry.Text != "" {
			num, _ := strconv.Atoi(d.numeratorEntry.Text)
			if denom < num {
				return fmt.Errorf("має бути >= чисельника")
			}
		}
		return nil
	}
	d.denominatorEntry.OnChanged = func(string) {
		d.updateSharePercentage()
	}

	// Відсоток
	d.sharePercentLabel = widget.NewLabel("0%")
	d.sharePercentLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Залишок
	d.remainingLabel = widget.NewLabel("")

	// Тип власності
	d.ownershipTypeSelect = widget.NewSelect([]string{
		"Повна власність",
		"Часткова власність",
		"Оренда",
	}, nil)
	d.ownershipTypeSelect.SetSelected("Повна власність")

	// Дата початку
	d.startDateEntry = widget.NewEntry()
	d.startDateEntry.SetPlaceHolder("ДД.ММ.РРРР")
	d.startDateEntry.SetText(time.Now().Format("02.01.2006"))
	d.startDateEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		_, err := time.Parse("02.01.2006", s)
		if err != nil {
			return fmt.Errorf("невірний формат дати")
		}
		return nil
	}

	// Дата завершення
	d.endDateEntry = widget.NewEntry()
	d.endDateEntry.SetPlaceHolder("ДД.ММ.РРРР (необов'язково)")
	d.endDateEntry.Validator = func(s string) error {
		if s != "" {
			endDate, err := time.Parse("02.01.2006", s)
			if err != nil {
				return fmt.Errorf("невірний формат дати")
			}
			// Перевірка що дата завершення > дата початку
			if d.startDateEntry.Text != "" {
				startDate, _ := time.Parse("02.01.2006", d.startDateEntry.Text)
				if endDate.Before(startDate) || endDate.Equal(startDate) {
					return fmt.Errorf("має бути після дати початку")
				}
			}
		}
		return nil
	}

	// Тип документа
	d.docTypeSelect = widget.NewSelect([]string{
		"",
		"Договір",
		"Свідоцтво",
		"Довідка",
		"Рішення суду",
		"Інше",
	}, nil)
	d.docTypeSelect.PlaceHolder = "Тип документа (необов'язково)"

	// Номер документа
	d.docNumberEntry = widget.NewEntry()
	d.docNumberEntry.SetPlaceHolder("Номер документа")

	// Дата документа
	d.docDateEntry = widget.NewEntry()
	d.docDateEntry.SetPlaceHolder("ДД.ММ.РРРР (необов'язково)")
	d.docDateEntry.Validator = func(s string) error {
		if s != "" {
			_, err := time.Parse("02.01.2006", s)
			if err != nil {
				return fmt.Errorf("невірний формат дати")
			}
		}
		return nil
	}

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
	if d.existingShare != nil {
		d.prefillExistingData()
	}
}

// prefillExistingData заповнює форму існуючими даними.
func (d *OwnershipFormDialog) prefillExistingData() {
	// Знаходимо власника
	for i, o := range d.owners {
		if o.ID == d.existingShare.OwnerID {
			d.ownerSelect.SetSelectedIndex(i)
			d.selectedOwnerID = o.ID
			break
		}
	}

	// Знаходимо квартиру
	for i, a := range d.apartments {
		if a.ID == d.existingShare.ApartmentID {
			d.apartmentSelect.SetSelectedIndex(i)
			d.selectedApartmentID = a.ID
			break
		}
	}

	// При редагуванні власник та квартира не змінюються
	d.ownerSelect.Disable()
	d.apartmentSelect.Disable()

	// Частка
	d.numeratorEntry.SetText(fmt.Sprintf("%d", d.existingShare.ShareNumerator))
	d.denominatorEntry.SetText(fmt.Sprintf("%d", d.existingShare.ShareDenominator))

	// Тип власності
	d.ownershipTypeSelect.SetSelected(d.existingShare.OwnershipTypeName)

	// Дати
	d.startDateEntry.SetText(d.existingShare.StartDate.Format("02.01.2006"))
	d.startDateEntry.Disable() // Дата початку не змінюється при редагуванні

	if d.existingShare.EndDate != nil {
		d.endDateEntry.SetText(d.existingShare.EndDate.Format("02.01.2006"))
	}

	// Документи
	if d.existingShare.DocumentType != nil {
		d.docTypeSelect.SetSelected(*d.existingShare.DocumentType)
	}
	if d.existingShare.DocumentNumber != nil {
		d.docNumberEntry.SetText(*d.existingShare.DocumentNumber)
	}
	if d.existingShare.DocumentDate != nil {
		d.docDateEntry.SetText(d.existingShare.DocumentDate.Format("02.01.2006"))
	}

	// Примітки
	if d.existingShare.Notes != nil {
		d.notesEntry.SetText(*d.existingShare.Notes)
	}
}

// Show показує діалог.
func (d *OwnershipFormDialog) Show() {
	title := "Додати частку власності"
	if d.existingShare != nil {
		title = "Редагувати частку власності"
	}

	// Ліва колонка
	leftColumn := container.NewVBox(
		widget.NewLabel("Основна інформація:"),
		widget.NewForm(
			widget.NewFormItem("* Власник", d.ownerSelect),
			widget.NewFormItem("* Квартира", d.apartmentSelect),
		),
		widget.NewSeparator(),
		widget.NewLabel("Частка власності:"),
		widget.NewForm(
			widget.NewFormItem("* Чисельник", d.numeratorEntry),
			widget.NewFormItem("* Знаменник", d.denominatorEntry),
		),
		container.NewHBox(
			widget.NewLabel("Відсоток:"),
			d.sharePercentLabel,
		),
		d.remainingLabel,
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("* Тип власності", d.ownershipTypeSelect),
		),
	)

	// Права колонка
	rightColumn := container.NewVBox(
		widget.NewLabel("Дати:"),
		widget.NewForm(
			widget.NewFormItem("* Дата початку", d.startDateEntry),
			widget.NewFormItem("Дата завершення", d.endDateEntry),
		),
		widget.NewSeparator(),
		widget.NewLabel("Документи:"),
		widget.NewForm(
			widget.NewFormItem("Тип", d.docTypeSelect),
			widget.NewFormItem("Номер", d.docNumberEntry),
			widget.NewFormItem("Дата", d.docDateEntry),
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
		container.NewVScroll(formContainer),
	)

	d.dialog = dialog.NewCustom(title, "", content, d.window)
	d.dialog.Resize(fyne.NewSize(900, 650))
	d.dialog.Show()
}

// updateSharePercentage оновлює відсоток.
func (d *OwnershipFormDialog) updateSharePercentage() {
	if d.numeratorEntry.Text == "" || d.denominatorEntry.Text == "" {
		d.sharePercentLabel.SetText("0%")
		return
	}

	num, err1 := strconv.Atoi(d.numeratorEntry.Text)
	denom, err2 := strconv.Atoi(d.denominatorEntry.Text)

	if err1 != nil || err2 != nil || denom == 0 {
		d.sharePercentLabel.SetText("???")
		return
	}

	percentage := float64(num) / float64(denom) * 100
	d.sharePercentLabel.SetText(fmt.Sprintf("%.2f%%", percentage))

	d.calculateRemaining()
}

// calculateRemaining розраховує скільки ще можна додати.
func (d *OwnershipFormDialog) calculateRemaining() {
	if d.selectedApartmentID == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Отримуємо залишок
	var excludeID *int64
	if d.existingShare != nil {
		excludeID = &d.existingShare.ID
	}

	result, err := d.ownershipService.CalculateRemaining(ctx, ownership.CalculateRemainingShareInput{
		CurrentUserID:  d.authManager.GetCurrentUserID(),
		ApartmentID:    d.selectedApartmentID,
		ExcludeShareID: excludeID,
	})

	if err != nil {
		d.remainingLabel.SetText("Помилка розрахунку")
		return
	}

	// Парсимо поточний відсоток
	var currentPercent float64
	fmt.Sscanf(d.sharePercentLabel.Text, "%f%%", &currentPercent)

	newTotal := result.TotalSharePercentage + currentPercent

	if newTotal > 100.01 {
		d.remainingLabel.SetText(fmt.Sprintf(
			"⚠️ УВАГА: Сума часток перевищить 100%% (буде %.2f%%)",
			newTotal,
		))
	} else {
		d.remainingLabel.SetText(fmt.Sprintf(
			"✅ Залишок після додавання: %.2f%%",
			100.0-newTotal,
		))
	}
}

// handleSave обробляє збереження.
func (d *OwnershipFormDialog) handleSave() {
	// Валідація
	if d.selectedOwnerID == 0 {
		dialog.ShowError(fmt.Errorf("Оберіть власника"), d.window)
		return
	}
	if d.selectedApartmentID == 0 {
		dialog.ShowError(fmt.Errorf("Оберіть квартиру"), d.window)
		return
	}
	if err := d.numeratorEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Чисельник: %v", err), d.window)
		return
	}
	if err := d.denominatorEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Знаменник: %v", err), d.window)
		return
	}
	if err := d.startDateEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Дата початку: %v", err), d.window)
		return
	}
	if d.endDateEntry.Text != "" {
		if err := d.endDateEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Дата завершення: %v", err), d.window)
			return
		}
	}

	d.performSave()
}

// performSave виконує збереження.
func (d *OwnershipFormDialog) performSave() {
	d.saveButton.Disable()
	defer d.saveButton.Enable()

	progress := dialog.NewCustomWithoutButtons(
		"Збереження...",
		widget.NewProgressBarInfinite(),
		d.window,
	)
	progress.Show()
	defer progress.Hide()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Парсинг даних
	numerator, _ := strconv.Atoi(d.numeratorEntry.Text)
	denominator, _ := strconv.Atoi(d.denominatorEntry.Text)
	startDate, _ := time.Parse("02.01.2006", d.startDateEntry.Text)

	// Тип власності
	var ownershipType entity.OwnershipType
	switch d.ownershipTypeSelect.Selected {
	case "Повна власність":
		ownershipType = entity.OwnershipTypeFull
	case "Часткова власність":
		ownershipType = entity.OwnershipTypeShared
	case "Оренда":
		ownershipType = entity.OwnershipTypeRent
	default:
		ownershipType = entity.OwnershipTypeFull
	}

	// Опціональні поля
	var endDate, docDate *time.Time
	var docType, docNumber, notes *string

	if d.endDateEntry.Text != "" {
		ed, _ := time.Parse("02.01.2006", d.endDateEntry.Text)
		endDate = &ed
	}
	if d.docDateEntry.Text != "" {
		dd, _ := time.Parse("02.01.2006", d.docDateEntry.Text)
		docDate = &dd
	}
	if d.docTypeSelect.Selected != "" {
		docType = &d.docTypeSelect.Selected
	}
	if d.docNumberEntry.Text != "" {
		docNumber = &d.docNumberEntry.Text
	}
	if d.notesEntry.Text != "" {
		notes = &d.notesEntry.Text
	}

	var err error

	if d.existingShare == nil {
		// Створення
		_, err = d.ownershipService.Create(ctx, ownership.CreateOwnershipShareInput{
			CurrentUserID:    d.authManager.GetCurrentUserID(),
			OwnerID:          d.selectedOwnerID,
			ApartmentID:      d.selectedApartmentID,
			ShareNumerator:   numerator,
			ShareDenominator: denominator,
			OwnershipType:    ownershipType,
			StartDate:        startDate,
			EndDate:          endDate,
			DocumentType:     docType,
			DocumentNumber:   docNumber,
			DocumentDate:     docDate,
			Notes:            notes,
		})
	} else {
		// Оновлення
		_, err = d.ownershipService.Update(ctx, ownership.UpdateOwnershipShareInput{
			CurrentUserID:    d.authManager.GetCurrentUserID(),
			ShareID:          d.existingShare.ID,
			ShareNumerator:   numerator,
			ShareDenominator: denominator,
			OwnershipType:    ownershipType,
			EndDate:          endDate,
			DocumentType:     docType,
			DocumentNumber:   docNumber,
			DocumentDate:     docDate,
			Notes:            notes,
		})
	}

	if err != nil {
		d.handleError(err)
		return
	}

	// Успіх
	d.dialog.Hide()

	successMsg := "Частку власності успішно створено!"
	if d.existingShare != nil {
		successMsg = "Дані частки успішно оновлено!"
	}

	dialog.ShowInformation("Успіх", successMsg, d.window)

	if d.OnSuccess != nil {
		d.OnSuccess()
	}
}

// handleError обробляє помилки.
func (d *OwnershipFormDialog) handleError(err error) {
	var message string

	var domainErr *domainErrors.DomainError
	if domainErrors.As(err, &domainErr) {
		if domainErr.Code == domainErrors.CodeValidationFailed {
			if domainErr.Err == entity.ErrOwnershipShareExceedsTotal {
				message = "Сума часток власності перевищує 100%!\n\n" + domainErr.Message
			} else {
				message = "Помилка валідації: " + domainErr.Message
			}
		} else if domainErr.Code == domainErrors.CodeDuplicateEntry {
			message = "Власник вже має активну частку у цій квартирі"
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
