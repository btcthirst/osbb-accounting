// presentation/fyne/screens/apartment_form_dialog.go
package dialogs

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
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/presentation/fyne/auth"
)

// ApartmentFormDialog - діалог для створення/редагування квартири.
type ApartmentFormDialog struct {
	window            fyne.Window
	apartmentService  service.ApartmentServiceInterface
	authManager       *auth.AuthManager
	existingApartment *apartment.ApartmentOutput

	// UI елементи
	dialog               dialog.Dialog
	apartmentNumberEntry *widget.Entry
	floorEntry           *widget.Entry
	entranceEntry        *widget.Entry
	areaTotalEntry       *widget.Entry
	areaLivingEntry      *widget.Entry
	roomsCountEntry      *widget.Entry
	cadastralNumberEntry *widget.Entry
	notesEntry           *widget.Entry
	saveButton           *widget.Button
	cancelButton         *widget.Button

	// Callbacks
	OnSuccess func()
}

// NewApartmentFormDialog створює новий діалог.
func NewApartmentFormDialog(
	window fyne.Window,
	apartmentService service.ApartmentServiceInterface,
	authManager *auth.AuthManager,
	existingApartment *apartment.ApartmentOutput,
) *ApartmentFormDialog {
	d := &ApartmentFormDialog{
		window:            window,
		apartmentService:  apartmentService,
		authManager:       authManager,
		existingApartment: existingApartment,
	}

	d.buildUI()
	return d
}

func (d *ApartmentFormDialog) buildUI() {
	// Номер квартири
	d.apartmentNumberEntry = widget.NewEntry()
	d.apartmentNumberEntry.SetPlaceHolder("Номер квартири")
	d.apartmentNumberEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		return nil
	}

	// Поверх
	d.floorEntry = widget.NewEntry()
	d.floorEntry.SetPlaceHolder("Поверх (0-99)")
	d.floorEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		floor, err := strconv.Atoi(s)
		if err != nil || floor < 0 {
			return fmt.Errorf("має бути числом >= 0")
		}
		return nil
	}

	// Під'їзд
	d.entranceEntry = widget.NewEntry()
	d.entranceEntry.SetPlaceHolder("Під'їзд (необов'язково)")
	d.entranceEntry.Validator = func(s string) error {
		if s != "" {
			entrance, err := strconv.Atoi(s)
			if err != nil || entrance <= 0 {
				return fmt.Errorf("має бути числом > 0")
			}
		}
		return nil
	}

	// Загальна площа
	d.areaTotalEntry = widget.NewEntry()
	d.areaTotalEntry.SetPlaceHolder("Загальна площа (м²)")
	d.areaTotalEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("обов'язкове поле")
		}
		area, err := strconv.ParseFloat(s, 64)
		if err != nil || area <= 0 {
			return fmt.Errorf("має бути числом > 0")
		}
		return nil
	}

	// Житлова площа
	d.areaLivingEntry = widget.NewEntry()
	d.areaLivingEntry.SetPlaceHolder("Житлова площа (необов'язково)")
	d.areaLivingEntry.Validator = func(s string) error {
		if s != "" {
			area, err := strconv.ParseFloat(s, 64)
			if err != nil || area <= 0 {
				return fmt.Errorf("має бути числом > 0")
			}
			// Перевірка що не більше загальної
			if d.areaTotalEntry.Text != "" {
				totalArea, _ := strconv.ParseFloat(d.areaTotalEntry.Text, 64)
				if area > totalArea {
					return fmt.Errorf("не може бути більше загальної площі")
				}
			}
		}
		return nil
	}

	// Кількість кімнат
	d.roomsCountEntry = widget.NewEntry()
	d.roomsCountEntry.SetPlaceHolder("Кількість кімнат")
	d.roomsCountEntry.Validator = func(s string) error {
		if s != "" {
			rooms, err := strconv.Atoi(s)
			if err != nil || rooms <= 0 {
				return fmt.Errorf("має бути числом > 0")
			}
		}
		return nil
	}

	// Кадастровий номер
	d.cadastralNumberEntry = widget.NewEntry()
	d.cadastralNumberEntry.SetPlaceHolder("Кадастровий номер (необов'язково)")

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
	if d.existingApartment != nil {
		d.apartmentNumberEntry.SetText(d.existingApartment.ApartmentNumber)
		d.apartmentNumberEntry.Disable() // Номер не можна змінювати

		d.floorEntry.SetText(fmt.Sprintf("%d", d.existingApartment.Floor))

		if d.existingApartment.Entrance != nil {
			d.entranceEntry.SetText(fmt.Sprintf("%d", *d.existingApartment.Entrance))
		}

		d.areaTotalEntry.SetText(fmt.Sprintf("%.1f", d.existingApartment.AreaTotal))

		if d.existingApartment.AreaLiving != nil {
			d.areaLivingEntry.SetText(fmt.Sprintf("%.1f", *d.existingApartment.AreaLiving))
		}

		if d.existingApartment.RoomsCount != nil {
			d.roomsCountEntry.SetText(fmt.Sprintf("%d", *d.existingApartment.RoomsCount))
		}

		if d.existingApartment.CadastralNumber != nil {
			d.cadastralNumberEntry.SetText(*d.existingApartment.CadastralNumber)
		}

		if d.existingApartment.Notes != nil {
			d.notesEntry.SetText(*d.existingApartment.Notes)
		}
	}
}

func (d *ApartmentFormDialog) Show() {
	title := "Додати квартиру"
	if d.existingApartment != nil {
		title = "Редагувати квартиру"
	}

	// Форма у дві колонки
	leftColumn := container.NewVBox(
		widget.NewLabel("Основна інформація:"),
		widget.NewForm(
			widget.NewFormItem("* Номер квартири", d.apartmentNumberEntry),
			widget.NewFormItem("* Поверх", d.floorEntry),
			widget.NewFormItem("Під'їзд", d.entranceEntry),
		),
	)

	rightColumn := container.NewVBox(
		widget.NewLabel("Площі та характеристики:"),
		widget.NewForm(
			widget.NewFormItem("* Загальна площа (м²)", d.areaTotalEntry),
			widget.NewFormItem("Житлова площа (м²)", d.areaLivingEntry),
			widget.NewFormItem("Кількість кімнат", d.roomsCountEntry),
			widget.NewFormItem("Кадастровий номер", d.cadastralNumberEntry),
		),
	)

	formContainer := container.NewGridWithColumns(2, leftColumn, rightColumn)

	// Примітки внизу
	notesContainer := container.NewVBox(
		widget.NewLabel("Примітки:"),
		d.notesEntry,
	)

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
		container.NewVScroll(
			container.NewVBox(
				formContainer,
				notesContainer,
			),
		),
	)

	d.dialog = dialog.NewCustom(title, "", content, d.window)
	d.dialog.Resize(fyne.NewSize(700, 550))
	d.dialog.Show()
}

func (d *ApartmentFormDialog) handleSave() {
	// Валідація обов'язкових полів
	if err := d.apartmentNumberEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Номер квартири: %v", err), d.window)
		return
	}
	if err := d.floorEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Поверх: %v", err), d.window)
		return
	}
	if err := d.areaTotalEntry.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("Загальна площа: %v", err), d.window)
		return
	}

	// Валідація опціональних полів
	if d.entranceEntry.Text != "" {
		if err := d.entranceEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Під'їзд: %v", err), d.window)
			return
		}
	}
	if d.areaLivingEntry.Text != "" {
		if err := d.areaLivingEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Житлова площа: %v", err), d.window)
			return
		}
	}
	if d.roomsCountEntry.Text != "" {
		if err := d.roomsCountEntry.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Кількість кімнат: %v", err), d.window)
			return
		}
	}

	d.performSave()
}

func (d *ApartmentFormDialog) performSave() {
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

	// Парсинг обов'язкових полів
	floor, _ := strconv.Atoi(d.floorEntry.Text)
	areaTotal, _ := strconv.ParseFloat(d.areaTotalEntry.Text, 64)

	// Парсинг опціональних полів
	var entrance, roomsCount *int
	var areaLiving *float64
	var cadastralNumber, notes *string

	if d.entranceEntry.Text != "" {
		e, _ := strconv.Atoi(d.entranceEntry.Text)
		entrance = &e
	}
	if d.areaLivingEntry.Text != "" {
		al, _ := strconv.ParseFloat(d.areaLivingEntry.Text, 64)
		areaLiving = &al
	}
	if d.roomsCountEntry.Text != "" {
		rc, _ := strconv.Atoi(d.roomsCountEntry.Text)
		roomsCount = &rc
	}
	if d.cadastralNumberEntry.Text != "" {
		cadastralNumber = &d.cadastralNumberEntry.Text
	}
	if d.notesEntry.Text != "" {
		notes = &d.notesEntry.Text
	}

	var err error
	var result *apartment.ApartmentOutput

	if d.existingApartment == nil {
		// Створення
		result, err = d.apartmentService.Create(ctx, apartment.CreateApartmentInput{
			CurrentUserID:   d.authManager.GetCurrentUserID(),
			ApartmentNumber: d.apartmentNumberEntry.Text,
			Floor:           floor,
			Entrance:        entrance,
			AreaTotal:       areaTotal,
			AreaLiving:      areaLiving,
			RoomsCount:      roomsCount,
			CadastralNumber: cadastralNumber,
			Notes:           notes,
		})
	} else {
		// Оновлення
		result, err = d.apartmentService.Update(ctx, apartment.UpdateApartmentInput{
			CurrentUserID:   d.authManager.GetCurrentUserID(),
			ApartmentID:     d.existingApartment.ID,
			Floor:           floor,
			Entrance:        entrance,
			AreaTotal:       areaTotal,
			AreaLiving:      areaLiving,
			RoomsCount:      roomsCount,
			CadastralNumber: cadastralNumber,
			Notes:           notes,
		})
	}

	if err != nil {
		d.handleError(err)
		return
	}

	// Успіх
	d.dialog.Hide()

	successMsg := "Квартиру успішно створено!"
	if d.existingApartment != nil {
		successMsg = "Дані квартири успішно оновлено!"
	}

	dialog.ShowInformation("Успіх",
		fmt.Sprintf("%s\n\nКвартира №%s", successMsg, result.ApartmentNumber),
		d.window)

	if d.OnSuccess != nil {
		d.OnSuccess()
	}
}

func (d *ApartmentFormDialog) handleError(err error) {
	var message string

	var domainErr *domainErrors.DomainError
	if domainErrors.As(err, &domainErr) {
		if domainErr.Code == domainErrors.CodeDuplicateEntry {
			message = "Квартира з таким номером вже існує"
			d.apartmentNumberEntry.FocusGained()
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
