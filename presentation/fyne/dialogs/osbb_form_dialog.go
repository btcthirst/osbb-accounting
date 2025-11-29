package dialogs

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/osbb"
	"osbb-accounting/presentation/fyne/auth"
)

func ShowOSBBFormDialog(
	window fyne.Window,
	service *service.OSBBService,
	authManager *auth.AuthManager,
	existingData *osbb.GetOSBBOutput,
	onSaved func(),
) {
	isEdit := existingData != nil
	title := "Створити ОСББ"
	if isEdit {
		title = "Редагувати ОСББ"
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Назва ОСББ")

	edrpouEntry := widget.NewEntry()
	edrpouEntry.SetPlaceHolder("ЄДРПОУ (8 цифр)")

	legalAddressEntry := widget.NewEntry()
	legalAddressEntry.SetPlaceHolder("Юридична адреса")

	actualAddressEntry := widget.NewEntry()
	actualAddressEntry.SetPlaceHolder("Фактична адреса (необов'язково)")

	chairmanNameEntry := widget.NewEntry()
	chairmanNameEntry.SetPlaceHolder("ПІБ голови правління")

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Телефон (необов'язково)")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email (необов'язково)")

	websiteEntry := widget.NewEntry()
	websiteEntry.SetPlaceHolder("Веб-сайт (необов'язково)")

	if isEdit {
		nameEntry.SetText(existingData.Name)
		edrpouEntry.SetText(existingData.EDRPOU)
		edrpouEntry.Disable() // EDRPOU usually cannot be changed easily
		legalAddressEntry.SetText(existingData.LegalAddress)
		chairmanNameEntry.SetText(existingData.ChairmanName)

		if existingData.ActualAddress != nil {
			actualAddressEntry.SetText(*existingData.ActualAddress)
		}
		if existingData.Phone != nil {
			phoneEntry.SetText(*existingData.Phone)
		}
		if existingData.Email != nil {
			emailEntry.SetText(*existingData.Email)
		}
		if existingData.Website != nil {
			websiteEntry.SetText(*existingData.Website)
		}
	}

	items := []*widget.FormItem{
		widget.NewFormItem("Назва", nameEntry),
		widget.NewFormItem("ЄДРПОУ", edrpouEntry),
		widget.NewFormItem("Юридична адреса", legalAddressEntry),
		widget.NewFormItem("Фактична адреса", actualAddressEntry),
		widget.NewFormItem("Голова правління", chairmanNameEntry),
		widget.NewFormItem("Телефон", phoneEntry),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Веб-сайт", websiteEntry),
	}

	dialog.ShowForm(title, "Зберегти", "Скасувати", items, func(confirmed bool) {
		if !confirmed {
			return
		}

		ctx := context.Background()
		userID := authManager.GetCurrentUser().ID

		actualAddress := actualAddressEntry.Text
		phone := phoneEntry.Text
		email := emailEntry.Text
		website := websiteEntry.Text

		var actualAddressPtr, phonePtr, emailPtr, websitePtr *string
		if actualAddress != "" {
			actualAddressPtr = &actualAddress
		}
		if phone != "" {
			phonePtr = &phone
		}
		if email != "" {
			emailPtr = &email
		}
		if website != "" {
			websitePtr = &website
		}

		if isEdit {
			input := osbb.UpdateOSBBInput{
				CurrentUserID: userID,
				Name:          nameEntry.Text,
				LegalAddress:  legalAddressEntry.Text,
				ActualAddress: actualAddressPtr,
				Phone:         phonePtr,
				Email:         emailPtr,
				Website:       websitePtr,
				ChairmanName:  chairmanNameEntry.Text,
			}
			_, err := service.Update(ctx, input)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Помилка оновлення: %w", err), window)
				return
			}
		} else {
			input := osbb.CreateOSBBInput{
				CurrentUserID: userID,
				Name:          nameEntry.Text,
				EDRPOU:        edrpouEntry.Text,
				LegalAddress:  legalAddressEntry.Text,
				ActualAddress: actualAddressPtr,
				Phone:         phonePtr,
				Email:         emailPtr,
				Website:       websitePtr,
				ChairmanName:  chairmanNameEntry.Text,
			}
			_, err := service.Create(ctx, input)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Помилка створення: %w", err), window)
				return
			}
		}

		if onSaved != nil {
			onSaved()
		}
	}, window)
}
