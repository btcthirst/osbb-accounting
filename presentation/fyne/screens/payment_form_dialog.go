// presentation/fyne/screens/payment_form_dialog.go
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
	"osbb-accounting/application/usecase/ownership"
	"osbb-accounting/application/usecase/payment"
	"osbb-accounting/domain/entity"
)

// PaymentFormDialog - діалог створення/редагування платежу
type PaymentFormDialog struct {
	window           fyne.Window
	paymentService   *service.PaymentService
	ownershipService *service.OwnershipService
	authManager      interface {
		GetCurrentUserID() int64
	}

	// Data
	existingPayment *payment.PaymentOutput
	ownershipShares []*ownership.OwnershipShareDetailsOutput

	// UI Elements
	ownershipShareSelect *widget.Select
	amountEntry          *widget.Entry
	paymentDateEntry     *widget.Entry // Format: DD.MM.YYYY
	periodMonthEntry     *widget.Entry // 1-12
	periodYearEntry      *widget.Entry // YYYY
	methodSelect         *widget.Select
	receiptNumberEntry   *widget.Entry
	notesEntry           *widget.Entry

	// Callbacks
	onSaved func()
}

// NewPaymentFormDialog створює новий діалог
func NewPaymentFormDialog(
	window fyne.Window,
	paymentService *service.PaymentService,
	ownershipService *service.OwnershipService,
	authManager interface {
		GetCurrentUserID() int64
	},
	existingPayment *payment.PaymentOutput,
) *PaymentFormDialog {
	return &PaymentFormDialog{
		window:           window,
		paymentService:   paymentService,
		ownershipService: ownershipService,
		authManager:      authManager,
		existingPayment:  existingPayment,
	}
}

// Show показує діалог
func (d *PaymentFormDialog) Show() {
	d.loadData()

	title := "Новий платіж"
	if d.existingPayment != nil {
		title = "Редагування платежу"
	}

	// Ownership Share Select
	var shareOptions []string
	shareMap := make(map[string]int64)
	for _, s := range d.ownershipShares {
		label := fmt.Sprintf("%s - %s (кв. %s)", s.OwnerName, s.ApartmentNumber, s.ShareFraction)
		shareOptions = append(shareOptions, label)
		shareMap[label] = s.ID
	}

	d.ownershipShareSelect = widget.NewSelect(shareOptions, nil)
	d.ownershipShareSelect.PlaceHolder = "Оберіть платника (Власник - Квартира)"

	// Amount
	d.amountEntry = widget.NewEntry()
	d.amountEntry.PlaceHolder = "0.00"

	// Date
	d.paymentDateEntry = widget.NewEntry()
	d.paymentDateEntry.PlaceHolder = "DD.MM.YYYY"
	d.paymentDateEntry.SetText(time.Now().Format("02.01.2006"))

	// Period
	d.periodMonthEntry = widget.NewEntry()
	d.periodMonthEntry.PlaceHolder = "MM"
	d.periodYearEntry = widget.NewEntry()
	d.periodYearEntry.PlaceHolder = "YYYY"

	// Set current period by default
	now := time.Now()
	d.periodMonthEntry.SetText(fmt.Sprintf("%d", now.Month()))
	d.periodYearEntry.SetText(fmt.Sprintf("%d", now.Year()))

	// Method
	methodOptions := []string{
		entity.PaymentMethodCash.GetDisplayName(),
		entity.PaymentMethodCard.GetDisplayName(),
		entity.PaymentMethodBankTransfer.GetDisplayName(),
		entity.PaymentMethodOther.GetDisplayName(),
	}
	d.methodSelect = widget.NewSelect(methodOptions, nil)
	d.methodSelect.SetSelected(entity.PaymentMethodCash.GetDisplayName())

	// Receipt Number
	d.receiptNumberEntry = widget.NewEntry()
	d.receiptNumberEntry.PlaceHolder = "Номер квитанції (опціонально)"

	// Notes
	d.notesEntry = widget.NewMultiLineEntry()
	d.notesEntry.PlaceHolder = "Примітки..."
	d.notesEntry.SetMinRowsVisible(3)

	// Pre-fill if editing
	if d.existingPayment != nil {
		// Find share label
		for label, id := range shareMap {
			if id == d.existingPayment.OwnershipShareID {
				d.ownershipShareSelect.SetSelected(label)
				break
			}
		}
		d.ownershipShareSelect.Disable() // Cannot change payer when editing

		d.amountEntry.SetText(fmt.Sprintf("%.2f", d.existingPayment.Amount))
		d.paymentDateEntry.SetText(d.existingPayment.PaymentDate.Format("02.01.2006"))

		if d.existingPayment.PeriodMonth != nil {
			d.periodMonthEntry.SetText(fmt.Sprintf("%d", *d.existingPayment.PeriodMonth))
		} else {
			d.periodMonthEntry.SetText("")
		}

		if d.existingPayment.PeriodYear != nil {
			d.periodYearEntry.SetText(fmt.Sprintf("%d", *d.existingPayment.PeriodYear))
		} else {
			d.periodYearEntry.SetText("")
		}

		d.methodSelect.SetSelected(d.existingPayment.MethodName)

		if d.existingPayment.ReceiptNumber != nil {
			d.receiptNumberEntry.SetText(*d.existingPayment.ReceiptNumber)
		}

		if d.existingPayment.Notes != nil {
			d.notesEntry.SetText(*d.existingPayment.Notes)
		}
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Платник", d.ownershipShareSelect),
		widget.NewFormItem("Сума (грн)", d.amountEntry),
		widget.NewFormItem("Дата оплати", d.paymentDateEntry),
		widget.NewFormItem("Період (Місяць/Рік)", container.NewGridWithColumns(2, d.periodMonthEntry, d.periodYearEntry)),
		widget.NewFormItem("Метод оплати", d.methodSelect),
		widget.NewFormItem("№ Квитанції", d.receiptNumberEntry),
		widget.NewFormItem("Примітки", d.notesEntry),
	}

	dialog.ShowForm(title, "Зберегти", "Скасувати", formItems, func(confirm bool) {
		if !confirm {
			return
		}

		// Validation & Save

		// 1. Share
		var shareID int64
		if d.existingPayment != nil {
			shareID = d.existingPayment.OwnershipShareID
		} else {
			selectedShare := d.ownershipShareSelect.Selected
			if selectedShare == "" {
				dialog.ShowError(fmt.Errorf("оберіть платника"), d.window)
				return
			}
			shareID = shareMap[selectedShare]
		}

		// 2. Amount
		amount, err := strconv.ParseFloat(d.amountEntry.Text, 64)
		if err != nil || amount <= 0 {
			dialog.ShowError(fmt.Errorf("некоректна сума"), d.window)
			return
		}

		// 3. Date
		pDate, err := time.Parse("02.01.2006", d.paymentDateEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("некоректна дата (DD.MM.YYYY)"), d.window)
			return
		}

		// 4. Period (Optional)
		var pMonth, pYear *int
		if d.periodMonthEntry.Text != "" {
			m, err := strconv.Atoi(d.periodMonthEntry.Text)
			if err != nil || m < 1 || m > 12 {
				dialog.ShowError(fmt.Errorf("некоректний місяць (1-12)"), d.window)
				return
			}
			pMonth = &m
		}
		if d.periodYearEntry.Text != "" {
			y, err := strconv.Atoi(d.periodYearEntry.Text)
			if err != nil || y < 2000 || y > 2100 {
				dialog.ShowError(fmt.Errorf("некоректний рік"), d.window)
				return
			}
			pYear = &y
		}

		// 5. Method
		var method entity.PaymentMethod
		switch d.methodSelect.Selected {
		case entity.PaymentMethodCash.GetDisplayName():
			method = entity.PaymentMethodCash
		case entity.PaymentMethodCard.GetDisplayName():
			method = entity.PaymentMethodCard
		case entity.PaymentMethodBankTransfer.GetDisplayName():
			method = entity.PaymentMethodBankTransfer
		case entity.PaymentMethodOther.GetDisplayName():
			method = entity.PaymentMethodOther
		default:
			method = entity.PaymentMethodOther
		}

		// 6. Optional fields
		var receiptNum *string
		if d.receiptNumberEntry.Text != "" {
			val := d.receiptNumberEntry.Text
			receiptNum = &val
		}

		var notes *string
		if d.notesEntry.Text != "" {
			val := d.notesEntry.Text
			notes = &val
		}

		// Execute
		ctx := context.Background()
		userID := d.authManager.GetCurrentUserID()

		if d.existingPayment == nil {
			// Create
			input := payment.CreatePaymentInput{
				CurrentUserID:    userID,
				OwnershipShareID: shareID,
				Amount:           amount,
				PaymentMethod:    method,
				PaymentPurpose:   "Оплата внесків", // Default purpose
				PaymentDate:      pDate,
				PeriodMonth:      pMonth,
				PeriodYear:       pYear,
				ReceiptNumber:    receiptNum,
				Notes:            notes,
			}
			_, err := d.paymentService.Create(ctx, input)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
		} else {
			// Update
			input := payment.UpdatePaymentInput{
				CurrentUserID:  userID,
				PaymentID:      d.existingPayment.ID,
				Amount:         amount,
				PaymentMethod:  method,
				PaymentPurpose: d.existingPayment.PaymentPurpose,
				PaymentDate:    pDate,
				PeriodMonth:    pMonth,
				PeriodYear:     pYear,
				ReceiptNumber:  receiptNum,
				Notes:          notes,
			}
			_, err := d.paymentService.Update(ctx, input)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
		}

		if d.onSaved != nil {
			d.onSaved()
		}
	}, d.window)
}

func (d *PaymentFormDialog) loadData() {
	ctx := context.Background()
	// Load ownership shares
	sharesOutput, err := d.ownershipService.List(ctx, ownership.ListOwnershipSharesInput{
		CurrentUserID: d.authManager.GetCurrentUserID(),
		Limit:         1000,
	})
	if err == nil {
		d.ownershipShares = sharesOutput.Shares
	}
}

// SetOnSaved sets the callback
func (d *PaymentFormDialog) SetOnSaved(f func()) {
	d.onSaved = f
}
