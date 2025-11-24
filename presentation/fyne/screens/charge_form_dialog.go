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
	"osbb-accounting/application/usecase/charge"
	"osbb-accounting/application/usecase/ownership"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/auth"
)

// ChargeFormDialog - діалог для створення/редагування нарахування.
type ChargeFormDialog struct {
	window           fyne.Window
	chargeService    service.ChargeServiceInterface
	ownershipService service.OwnershipServiceInterface
	authManager      *auth.AuthManager
	existingCharge   *charge.ChargeOutput

	// UI елементи
	dialog               dialog.Dialog
	ownershipShareSelect *widget.Select
	chargeTypeSelect     *widget.Select
	chargeDateEntry      *widget.Entry
	periodMonthEntry     *widget.Entry
	periodYearEntry      *widget.Entry
	amountEntry          *widget.Entry
	descriptionEntry     *widget.Entry
	saveButton           *widget.Button
	cancelButton         *widget.Button

	// Дані
	ownershipShares []*ownership.OwnershipShareDetailsOutput
	selectedShareID int64

	// Callbacks
	OnSuccess func()
}

// NewChargeFormDialog створює новий діалог.
func NewChargeFormDialog(
	window fyne.Window,
	chargeService service.ChargeServiceInterface,
	ownershipService service.OwnershipServiceInterface,
	authManager *auth.AuthManager,
	existingCharge *charge.ChargeOutput,
) *ChargeFormDialog {
	d := &ChargeFormDialog{
		window:           window,
		chargeService:    chargeService,
		ownershipService: ownershipService,
		authManager:      authManager,
		existingCharge:   existingCharge,
	}

	d.loadData()
	d.buildUI()
	return d
}

// loadData завантажує список часток власності.
func (d *ChargeFormDialog) loadData() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завантажуємо активні частки власності
	// TODO: Можливо, варто додати фільтрацію або пагінацію, якщо часток багато
	sharesOutput, err := d.ownershipService.List(ctx, ownership.ListOwnershipSharesInput{
		CurrentUserID: d.authManager.GetCurrentUserID(),
		Limit:         1000,
		// Можна додати фільтр IsActive=true, якщо такий є
	})
	if err == nil {
		d.ownershipShares = sharesOutput.Shares
	}
}

// buildUI створює інтерфейс діалогу.
func (d *ChargeFormDialog) buildUI() {
	// Вибір частки власності
	shareNames := make([]string, len(d.ownershipShares))
	for i, s := range d.ownershipShares {
		shareNames[i] = fmt.Sprintf("%s - Кв. %s", s.OwnerName, s.ApartmentNumber)
	}
	d.ownershipShareSelect = widget.NewSelect(shareNames, func(selected string) {
		for i, name := range shareNames {
			if name == selected {
				d.selectedShareID = d.ownershipShares[i].ID
				break
			}
		}
	})
	d.ownershipShareSelect.PlaceHolder = "Оберіть власника/квартиру"

	// Тип нарахування
	chargeTypes := []string{
		entity.ChargeTypeMaintenance.GetDisplayName(),
		entity.ChargeTypeUtility.GetDisplayName(),
		entity.ChargeTypeRepair.GetDisplayName(),
		entity.ChargeTypePenalty.GetDisplayName(),
		entity.ChargeTypeOther.GetDisplayName(),
	}
	d.chargeTypeSelect = widget.NewSelect(chargeTypes, nil)
	d.chargeTypeSelect.PlaceHolder = "Оберіть тип нарахування"

	// Дата нарахування
	d.chargeDateEntry = widget.NewEntry()
	d.chargeDateEntry.SetPlaceHolder("ДД.ММ.РРРР")
	d.chargeDateEntry.SetText(time.Now().Format("02.01.2006"))
	d.chargeDateEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		_, err := time.Parse("02.01.2006", s)
		if err != nil {
			return fmt.Errorf("невірний формат дати")
		}
		return nil
	}

	// Період (Місяць/Рік)
	now := time.Now()
	d.periodMonthEntry = widget.NewEntry()
	d.periodMonthEntry.SetText(fmt.Sprintf("%d", now.Month()))
	d.periodMonthEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		m, err := strconv.Atoi(s)
		if err != nil || m < 1 || m > 12 {
			return fmt.Errorf("місяць від 1 до 12")
		}
		return nil
	}

	d.periodYearEntry = widget.NewEntry()
	d.periodYearEntry.SetText(fmt.Sprintf("%d", now.Year()))
	d.periodYearEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		y, err := strconv.Atoi(s)
		if err != nil || y < 2000 || y > 2100 {
			return fmt.Errorf("невірний рік")
		}
		return nil
	}

	// Сума
	d.amountEntry = widget.NewEntry()
	d.amountEntry.SetPlaceHolder("0.00")
	d.amountEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("має бути числом")
		}
		return nil
	}

	// Опис
	d.descriptionEntry = widget.NewMultiLineEntry()
	d.descriptionEntry.SetPlaceHolder("Опис нарахування...")
	d.descriptionEntry.SetMinRowsVisible(3)

	// Кнопки
	d.saveButton = widget.NewButton("Зберегти", d.handleSave)
	d.saveButton.Importance = widget.HighImportance

	d.cancelButton = widget.NewButton("Скасувати", func() {
		d.dialog.Hide()
	})

	// Заповнення даних при редагуванні
	if d.existingCharge != nil {
		d.prefillExistingData()
	}
}

// prefillExistingData заповнює форму існуючими даними.
func (d *ChargeFormDialog) prefillExistingData() {
	// Знаходимо частку
	for i, s := range d.ownershipShares {
		if s.ID == d.existingCharge.OwnershipShareID {
			d.ownershipShareSelect.SetSelectedIndex(i)
			d.selectedShareID = s.ID
			break
		}
	}
	d.ownershipShareSelect.Disable() // Не можна змінювати частку при редагуванні

	// Тип
	d.chargeTypeSelect.SetSelected(d.existingCharge.TypeName)

	// Дата
	d.chargeDateEntry.SetText(d.existingCharge.ChargeDate.Format("02.01.2006"))

	// Період
	d.periodMonthEntry.SetText(fmt.Sprintf("%d", d.existingCharge.PeriodMonth))
	d.periodYearEntry.SetText(fmt.Sprintf("%d", d.existingCharge.PeriodYear))

	// Сума
	d.amountEntry.SetText(fmt.Sprintf("%.2f", d.existingCharge.Amount))

	// Опис
	if d.existingCharge.Description != nil {
		d.descriptionEntry.SetText(*d.existingCharge.Description)
	}
}

// Show показує діалог.
func (d *ChargeFormDialog) Show() {
	title := "Створити нарахування"
	if d.existingCharge != nil {
		title = "Редагувати нарахування"
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Платник (Власник/Квартира)", d.ownershipShareSelect),
		widget.NewFormItem("Тип нарахування", d.chargeTypeSelect),
		widget.NewFormItem("Дата нарахування", d.chargeDateEntry),
		widget.NewFormItem("Період (Місяць)", d.periodMonthEntry),
		widget.NewFormItem("Період (Рік)", d.periodYearEntry),
		widget.NewFormItem("Сума (грн)", d.amountEntry),
		widget.NewFormItem("Опис", d.descriptionEntry),
	}

	form := widget.NewForm(formItems...)

	// Кнопки
	buttons := container.NewGridWithColumns(2, d.cancelButton, d.saveButton)

	content := container.NewBorder(
		nil,
		buttons,
		nil,
		nil,
		container.NewVScroll(form),
	)

	d.dialog = dialog.NewCustom(title, "Скасувати", content, d.window)
	d.dialog.Resize(fyne.NewSize(500, 600))
	d.dialog.Show()
}

// handleSave обробляє збереження.
func (d *ChargeFormDialog) handleSave() {
	// Валідація
	if d.selectedShareID == 0 {
		dialog.ShowError(fmt.Errorf("Оберіть платника"), d.window)
		return
	}
	if d.chargeTypeSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("Оберіть тип нарахування"), d.window)
		return
	}
	if err := d.chargeDateEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Дата: %v", err), d.window)
		return
	}
	if err := d.periodMonthEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Місяць: %v", err), d.window)
		return
	}
	if err := d.periodYearEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Рік: %v", err), d.window)
		return
	}
	if err := d.amountEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Сума: %v", err), d.window)
		return
	}

	d.performSave()
}

// performSave виконує збереження.
func (d *ChargeFormDialog) performSave() {
	d.saveButton.Disable()
	defer d.saveButton.Enable()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Парсинг даних
	chargeDate, _ := time.Parse("02.01.2006", d.chargeDateEntry.Text)
	periodMonth, _ := strconv.Atoi(d.periodMonthEntry.Text)
	periodYear, _ := strconv.Atoi(d.periodYearEntry.Text)
	amount, _ := strconv.ParseFloat(d.amountEntry.Text, 64)

	var description *string
	if d.descriptionEntry.Text != "" {
		desc := d.descriptionEntry.Text
		description = &desc
	}

	// Визначення типу нарахування
	var chargeType entity.ChargeType
	switch d.chargeTypeSelect.Selected {
	case entity.ChargeTypeMaintenance.GetDisplayName():
		chargeType = entity.ChargeTypeMaintenance
	case entity.ChargeTypeUtility.GetDisplayName():
		chargeType = entity.ChargeTypeUtility
	case entity.ChargeTypeRepair.GetDisplayName():
		chargeType = entity.ChargeTypeRepair
	case entity.ChargeTypePenalty.GetDisplayName():
		chargeType = entity.ChargeTypePenalty
	case entity.ChargeTypeOther.GetDisplayName():
		chargeType = entity.ChargeTypeOther
	default:
		chargeType = entity.ChargeTypeOther
	}

	var err error

	if d.existingCharge == nil {
		// Створення
		_, err = d.chargeService.Create(ctx, charge.CreateChargeInput{
			CurrentUserID:    d.authManager.GetCurrentUserID(),
			OwnershipShareID: d.selectedShareID,
			ChargeType:       chargeType,
			ChargeDate:       chargeDate,
			PeriodMonth:      periodMonth,
			PeriodYear:       periodYear,
			Amount:           amount,
			Description:      description,
		})
	} else {
		// Оновлення
		_, err = d.chargeService.Update(ctx, charge.UpdateChargeInput{
			CurrentUserID: d.authManager.GetCurrentUserID(),
			ChargeID:      d.existingCharge.ID,
			Amount:        amount,
			Description:   description,
			// TODO: Додати можливість оновлення інших полів, якщо потрібно
		})
	}

	if err != nil {
		dialog.ShowError(fmt.Errorf("Помилка збереження: %v", err), d.window)
		return
	}

	// Успіх
	d.dialog.Hide()
	dialog.ShowInformation("Успіх", "Нарахування успішно збережено", d.window)

	if d.OnSuccess != nil {
		d.OnSuccess()
	}
}
