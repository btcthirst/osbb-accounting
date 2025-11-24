// presentation/fyne/screens/payments_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/payment"
)

// PaymentsScreen - екран списку платежів
type PaymentsScreen struct {
	window           fyne.Window
	paymentService   *service.PaymentService
	ownershipService *service.OwnershipService
	authManager      interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	}

	// UI
	content *fyne.Container
	table   *widget.Table

	// Data
	payments []*payment.PaymentOutput

	// Filters
	searchEntry  *widget.Entry
	filterSelect *widget.Select
}

// NewPaymentsScreen створює новий екран
func NewPaymentsScreen(
	window fyne.Window,
	paymentService *service.PaymentService,
	ownershipService *service.OwnershipService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *PaymentsScreen {
	s := &PaymentsScreen{
		window:           window,
		paymentService:   paymentService,
		ownershipService: ownershipService,
		authManager:      authManager,
	}
	s.buildUI()
	return s
}

func (s *PaymentsScreen) buildUI() {
	// Toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			s.showPaymentDialog(nil)
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			s.loadPayments()
		}),
	)

	// Filters
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("Пошук...")
	s.searchEntry.OnSubmitted = func(_ string) { s.loadPayments() }

	s.filterSelect = widget.NewSelect([]string{
		"Всі",
		"Поточний місяць",
		"Минулий місяць",
		"Непідтверджені",
		"Підтверджені",
	}, func(selected string) {
		s.loadPayments()
	})
	s.filterSelect.SetSelected("Всі")

	filterContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(widget.NewLabel("Фільтр:"), s.filterSelect, widget.NewButtonWithIcon("", theme.SearchIcon(), func() { s.loadPayments() })),
		s.searchEntry,
	)

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.payments), 8 // Cols: ID, Date, Period, Method, Amount, Approved, Receipt, Actions
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell content")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			p := s.payments[id.Row]
			label := cell.(*widget.Label)

			switch id.Col {
			case 0: // ID
				label.SetText(fmt.Sprintf("%d", p.ID))
			case 1: // Date
				label.SetText(p.PaymentDate.Format("02.01.2006"))
			case 2: // Period
				if p.PeriodMonth != nil && p.PeriodYear != nil {
					label.SetText(fmt.Sprintf("%02d/%d", *p.PeriodMonth, *p.PeriodYear))
				} else {
					label.SetText("-")
				}
			case 3: // Method
				label.SetText(p.MethodName)
			case 4: // Amount
				label.SetText(fmt.Sprintf("%.2f грн", p.Amount))
			case 5: // Approved
				if p.IsApproved {
					label.SetText("Так")
				} else {
					label.SetText("Ні")
				}
			case 6: // Receipt
				if p.ReceiptNumber != nil {
					label.SetText(*p.ReceiptNumber)
				} else {
					label.SetText("")
				}
			case 7: // Actions
				label.SetText("⚙️") // Placeholder for actions
			}
		},
	)

	// Column widths
	s.table.SetColumnWidth(0, 50)  // ID
	s.table.SetColumnWidth(1, 100) // Date
	s.table.SetColumnWidth(2, 80)  // Period
	s.table.SetColumnWidth(3, 150) // Method
	s.table.SetColumnWidth(4, 100) // Amount
	s.table.SetColumnWidth(5, 80)  // Approved
	s.table.SetColumnWidth(6, 100) // Receipt
	s.table.SetColumnWidth(7, 50)  // Actions

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id) // Don't keep selection
		if id.Col == 7 {
			s.showActionsMenu(id.Row)
		}
	}

	// Layout
	s.content = container.NewBorder(
		container.NewVBox(toolbar, filterContainer),
		nil, nil, nil,
		s.table,
	)

	s.loadPayments()
}

func (s *PaymentsScreen) Render() fyne.CanvasObject {
	return s.content
}

func (s *PaymentsScreen) loadPayments() {
	ctx := context.Background()
	input := payment.ListPaymentsInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
		OrderDesc:     true,
		OrderBy:       "payment_date",
	}

	// Apply filters
	now := time.Now()
	switch s.filterSelect.Selected {
	case "Поточний місяць":
		m := int(now.Month())
		y := now.Year()
		input.PeriodMonth = &m
		input.PeriodYear = &y
	case "Минулий місяць":
		prev := now.AddDate(0, -1, 0)
		m := int(prev.Month())
		y := prev.Year()
		input.PeriodMonth = &m
		input.PeriodYear = &y
	case "Непідтверджені":
		approved := false
		input.IsApproved = &approved
	case "Підтверджені":
		approved := true
		input.IsApproved = &approved
	}

	output, err := s.paymentService.List(ctx, input)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.payments = output.Payments
	s.table.Refresh()
}

func (s *PaymentsScreen) showPaymentDialog(existing *payment.PaymentOutput) {
	d := NewPaymentFormDialog(
		s.window,
		s.paymentService,
		s.ownershipService,
		s.authManager,
		existing,
	)
	d.SetOnSaved(func() {
		s.loadPayments()
	})
	d.Show()
}

func (s *PaymentsScreen) showActionsMenu(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(s.payments) {
		return
	}
	p := s.payments[rowIndex]

	menu := fyne.NewMenu("Дії",
		fyne.NewMenuItem("Редагувати", func() {
			if p.IsApproved {
				dialog.ShowInformation("Інформація", "Не можна редагувати підтверджений платіж", s.window)
				return
			}
			s.showPaymentDialog(p)
		}),
		fyne.NewMenuItem("Видалити", func() {
			if p.IsApproved {
				dialog.ShowInformation("Інформація", "Не можна видалити підтверджений платіж", s.window)
				return
			}
			dialog.ShowConfirm("Видалення", "Ви впевнені, що хочете видалити цей платіж?", func(ok bool) {
				if ok {
					s.deletePayment(p.ID)
				}
			}, s.window)
		}),
		fyne.NewMenuItemSeparator(),
	)

	if p.IsApproved {
		menu.Items = append(menu.Items, fyne.NewMenuItem("Скасувати підтвердження", func() {
			s.unapprovePayment(p.ID)
		}))
	} else {
		menu.Items = append(menu.Items, fyne.NewMenuItem("Підтвердити", func() {
			s.approvePayment(p.ID)
		}))
	}

	widget.ShowPopUpMenuAtPosition(menu, s.window.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(s.table))
}

func (s *PaymentsScreen) deletePayment(id int64) {
	ctx := context.Background()
	_, err := s.paymentService.Delete(ctx, payment.DeletePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadPayments()
}

func (s *PaymentsScreen) approvePayment(id int64) {
	ctx := context.Background()
	_, err := s.paymentService.Approve(ctx, payment.ApprovePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadPayments()
}

func (s *PaymentsScreen) unapprovePayment(id int64) {
	ctx := context.Background()
	_, err := s.paymentService.Unapprove(ctx, payment.UnapprovePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	s.loadPayments()
}
