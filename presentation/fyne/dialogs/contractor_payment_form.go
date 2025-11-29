// presentation/fyne/dialogs/contractor_payment_form.go
package dialogs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/contractor"
	"osbb-accounting/application/usecase/contractorpayment"
	"osbb-accounting/domain/entity"
	"osbb-accounting/presentation/fyne/common"
)

// ContractorPaymentFormDialog - діалог створення/редагування платежу контрагента
type ContractorPaymentFormDialog struct {
	window                   fyne.Window
	contractorPaymentService *service.ContractorPaymentService
	contractorService        *service.ContractorService
	authManager              interface {
		GetCurrentUserID() int64
	}
	existingPayment *contractorpayment.ContractorPaymentOutput
	onSaved         func()

	// Form fields
	contractorSelect *widget.Select
	amountEntry      *widget.Entry
	datePicker       *widget.Entry // TODO: Use proper date picker when available
	methodSelect     *widget.Select
	purposeEntry     *widget.Entry
	periodMonthEntry *widget.Entry
	periodYearEntry  *widget.Entry
	receiptEntry     *widget.Entry
	notesEntry       *widget.Entry

	// Data mapping
	contractors          []*contractor.ContractorOutput
	contractorMap        map[string]int64
	selectedContractorID int64
}

// NewContractorPaymentFormDialog створює новий діалог
func NewContractorPaymentFormDialog(
	window fyne.Window,
	contractorPaymentService *service.ContractorPaymentService,
	contractorService *service.ContractorService,
	authManager interface {
		GetCurrentUserID() int64
	},
	existingPayment *contractorpayment.ContractorPaymentOutput,
) *ContractorPaymentFormDialog {
	return &ContractorPaymentFormDialog{
		window:                   window,
		contractorPaymentService: contractorPaymentService,
		contractorService:        contractorService,
		authManager:              authManager,
		existingPayment:          existingPayment,
		contractorMap:            make(map[string]int64),
	}
}

// SetOnSaved встановлює callback функцію, яка викликається після успішного збереження
func (d *ContractorPaymentFormDialog) SetOnSaved(f func()) {
	d.onSaved = f
}

// Show показує діалог
func (d *ContractorPaymentFormDialog) Show() {
	d.loadContractors(func() {
		d.showDialog()
	})
}

func (d *ContractorPaymentFormDialog) loadContractors(onLoaded func()) {
	ctx := context.Background()
	input := contractor.ListContractorsInput{
		CurrentUserID: d.authManager.GetCurrentUserID(),
		Limit:         1000, // Load all (or reasonable limit)
	}

	output, err := d.contractorService.List(ctx, input)
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}

	d.contractors = output.Contractors
	d.contractorMap = make(map[string]int64)
	for _, c := range d.contractors {
		d.contractorMap[c.Name] = c.ID
	}

	onLoaded()
}

func (d *ContractorPaymentFormDialog) showDialog() {
	// Init fields
	contractorNames := make([]string, 0, len(d.contractors))
	for _, c := range d.contractors {
		contractorNames = append(contractorNames, c.Name)
	}

	d.contractorSelect = widget.NewSelect(contractorNames, func(s string) {
		d.selectedContractorID = d.contractorMap[s]
	})

	d.amountEntry = widget.NewEntry()
	d.amountEntry.PlaceHolder = "0.00"

	d.datePicker = widget.NewEntry()
	d.datePicker.PlaceHolder = "DD.MM.YYYY"
	d.datePicker.SetText(time.Now().Format("02.01.2006"))

	d.methodSelect = widget.NewSelect([]string{
		"Готівка", "Картка", "Банківський переказ", "Інше",
	}, nil)
	d.methodSelect.SetSelected("Банківський переказ")

	d.purposeEntry = widget.NewEntry()
	d.purposeEntry.PlaceHolder = "Призначення платежу"

	d.periodMonthEntry = widget.NewEntry()
	d.periodMonthEntry.PlaceHolder = "MM"

	d.periodYearEntry = widget.NewEntry()
	d.periodYearEntry.PlaceHolder = "YYYY"

	d.receiptEntry = widget.NewEntry()
	d.receiptEntry.PlaceHolder = "Номер квитанції"

	d.notesEntry = widget.NewEntry()
	d.notesEntry.PlaceHolder = "Примітки"

	// Fill if editing
	title := "Новий платіж контрагенту"
	if d.existingPayment != nil {
		title = "Редагування платежу"

		// Set contractor selection
		// Find contractor name by ID
		var contractorName string
		for name, id := range d.contractorMap {
			if id == d.existingPayment.ContractorID {
				contractorName = name
				break
			}
		}
		if contractorName != "" {
			d.contractorSelect.SetSelected(contractorName)
		}

		d.amountEntry.SetText(fmt.Sprintf("%.2f", d.existingPayment.Amount))
		d.datePicker.SetText(d.existingPayment.PaymentDate.Format("02.01.2006"))

		switch entity.PaymentMethod(d.existingPayment.PaymentMethod) {
		case entity.PaymentMethodCash:
			d.methodSelect.SetSelected("Готівка")
		case entity.PaymentMethodCard:
			d.methodSelect.SetSelected("Картка")
		case entity.PaymentMethodBankTransfer:
			d.methodSelect.SetSelected("Банківський переказ")
		case entity.PaymentMethodOther:
			d.methodSelect.SetSelected("Інше")
		}

		d.purposeEntry.SetText(d.existingPayment.Purpose)

		if d.existingPayment.PeriodMonth != nil {
			d.periodMonthEntry.SetText(fmt.Sprintf("%d", *d.existingPayment.PeriodMonth))
		}
		if d.existingPayment.PeriodYear != nil {
			d.periodYearEntry.SetText(fmt.Sprintf("%d", *d.existingPayment.PeriodYear))
		}

		if d.existingPayment.ReceiptNumber != nil {
			d.receiptEntry.SetText(*d.existingPayment.ReceiptNumber)
		}
		if d.existingPayment.Notes != nil {
			d.notesEntry.SetText(*d.existingPayment.Notes)
		}
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Контрагент", Widget: d.contractorSelect},
			{Text: "Сума (грн)", Widget: d.amountEntry},
			{Text: "Дата", Widget: d.datePicker},
			{Text: "Метод", Widget: d.methodSelect},
			{Text: "Призначення", Widget: d.purposeEntry},
			{Text: "Період (місяць)", Widget: d.periodMonthEntry},
			{Text: "Період (рік)", Widget: d.periodYearEntry},
			{Text: "Квитанція", Widget: d.receiptEntry},
			{Text: "Примітки", Widget: d.notesEntry},
		},
	}

	dialog.ShowCustomConfirm(title, "Зберегти", "Скасувати", form, func(ok bool) {
		if ok {
			d.save()
		}
	}, d.window)
}

func (d *ContractorPaymentFormDialog) save() {
	// Validation and parsing
	amount, err := strconv.ParseFloat(d.amountEntry.Text, 64)
	if err != nil {
		dialog.ShowError(fmt.Errorf("некоректна сума"), d.window)
		return
	}

	date, err := time.Parse("02.01.2006", d.datePicker.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("некоректна дата (формат DD.MM.YYYY)"), d.window)
		return
	}

	var method entity.PaymentMethod
	switch d.methodSelect.Selected {
	case "Готівка":
		method = entity.PaymentMethodCash
	case "Картка":
		method = entity.PaymentMethodCard
	case "Банківський переказ":
		method = entity.PaymentMethodBankTransfer
	default:
		method = entity.PaymentMethodOther
	}

	var periodMonth *int
	if d.periodMonthEntry.Text != "" {
		m, err := strconv.Atoi(d.periodMonthEntry.Text)
		if err == nil && m >= 1 && m <= 12 {
			periodMonth = &m
		}
	}

	var periodYear *int
	if d.periodYearEntry.Text != "" {
		y, err := strconv.Atoi(d.periodYearEntry.Text)
		if err == nil && y >= 2000 {
			periodYear = &y
		}
	}

	var receipt *string
	if d.receiptEntry.Text != "" {
		t := d.receiptEntry.Text
		receipt = &t
	}

	var notes *string
	if d.notesEntry.Text != "" {
		t := d.notesEntry.Text
		notes = &t
	}

	ctx := context.Background() // TODO: Use proper context

	if d.existingPayment == nil {
		// Create
		input := contractorpayment.CreateContractorPaymentInput{
			CurrentUserID: d.authManager.GetCurrentUserID(),
			ContractorID:  d.selectedContractorID, // TODO: Validate selected
			Amount:        amount,
			PaymentMethod: method,
			Purpose:       d.purposeEntry.Text,
			PaymentDate:   date,
			PeriodMonth:   periodMonth,
			PeriodYear:    periodYear,
			ReceiptNumber: receipt,
			Notes:         notes,
		}

		_, err := d.contractorPaymentService.Create(ctx, input)
		if err != nil {
			common.ShowError(d.window, err)
			return
		}
	} else {
		// Update
		input := contractorpayment.UpdateContractorPaymentInput{
			CurrentUserID: d.authManager.GetCurrentUserID(),
			PaymentID:     d.existingPayment.ID,
			Amount:        amount,
			PaymentMethod: method,
			Purpose:       d.purposeEntry.Text,
			PaymentDate:   date,
			PeriodMonth:   periodMonth,
			PeriodYear:    periodYear,
			ReceiptNumber: receipt,
			Notes:         notes,
		}

		_, err := d.contractorPaymentService.Update(ctx, input)
		if err != nil {
			common.ShowError(d.window, err)
			return
		}
	}

	if d.onSaved != nil {
		d.onSaved()
	}
}
