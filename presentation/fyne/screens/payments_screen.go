// presentation/fyne/screens/payments_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/payment"
	"osbb-accounting/presentation/fyne/common"
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
	createButton  *widget.Button
	refreshButton *widget.Button
	table         *widget.Table
	searchEntry   *widget.Entry
	filterSelect  *widget.Select
	statsLabel    *widget.Label

	// Data
	payments         []*payment.PaymentOutput
	filteredPayments []*payment.PaymentOutput
	currentPage      int
	pageSize         int
	totalCount       int64
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
	// Buttons
	s.createButton = newCreateButton("Додати платіж", func() {
		s.showPaymentDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadPayments()
	})

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.payments) + 1, 8 // +1 for header
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell content")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				headers := []string{"ID", "Дата", "Період", "Метод", "Сума", "Підтверджено", "Квитанція", "Дії"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			// Data
			if id.Row-1 >= len(s.payments) {
				label.SetText("")
				return
			}
			p := s.payments[id.Row-1]

			// Reset style
			label.TextStyle = fyne.TextStyle{}

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
		if id.Row == 0 {
			return
		}
		if id.Col == 7 {
			s.showActionsMenu(id.Row - 1)
		}
	}

	// Filters
	s.searchEntry = newSearchEntry("Пошук...", func(_ string) { s.applyFilters() })

	s.filterSelect = newFilterSelect([]string{
		"Всі",
		"Поточний місяць",
		"Минулий місяць",
		"Непідтверджені",
		"Підтверджені",
	}, func(selected string) {
		s.loadPayments()
	})

	// Stats
	s.statsLabel = widget.NewLabel("")

	// Trigger initial load
	s.filterSelect.SetSelected("Всі")
}

func (s *PaymentsScreen) Render() fyne.CanvasObject {
	// Toolbar container
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(s.createButton, s.refreshButton),
		nil,
		container.NewVBox(
			s.searchEntry,
			s.filterSelect,
		),
	)

	// Layout
	content := container.NewBorder(
		toolbar,
		s.statsLabel,
		nil, nil,
		s.table,
	)
	return content
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
	s.totalCount = output.Total
	s.table.Refresh()
	s.updateStats()
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

	actions := []common.Action{
		{
			Label: "✏️ Редагувати",
			OnTap: func() {
				if p.IsApproved {
					common.ShowInformation(s.window, "Інформація", "Не можна редагувати підтверджений платіж")
					return
				}
				s.showPaymentDialog(p)
			},
		},
		{
			Label: "🗑️ Видалити",
			OnTap: func() {
				if p.IsApproved {
					common.ShowInformation(s.window, "Інформація", "Не можна видалити підтверджений платіж")
					return
				}
				s.confirmDelete(p)
			},
			Importance: widget.DangerImportance,
		},
	}

	if p.IsApproved {
		actions = append(actions, common.Action{
			Label: "↩️ Скасувати підтвердження",
			OnTap: func() { s.unapprovePayment(p.ID) },
		})
	} else {
		actions = append(actions, common.Action{
			Label: "✅ Підтвердити",
			OnTap: func() { s.approvePayment(p.ID) },
		})
	}

	common.ShowActionsMenu(s.window, fmt.Sprintf("Платіж #%d", p.ID), actions)
}

func (s *PaymentsScreen) confirmDelete(p *payment.PaymentOutput) {
	common.ShowDeleteConfirmation(
		s.window,
		"Підтвердження видалення",
		fmt.Sprintf("Ви впевнені, що хочете видалити платіж #%d?", p.ID),
		func() {
			s.deletePayment(p.ID)
		},
	)
}

func (s *PaymentsScreen) deletePayment(id int64) {
	ctx := context.Background()
	_, err := s.paymentService.Delete(ctx, payment.DeletePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Платіж успішно видалено")
	s.loadPayments()
}

func (s *PaymentsScreen) approvePayment(id int64) {
	ctx := context.Background()
	_, err := s.paymentService.Approve(ctx, payment.ApprovePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Платіж підтверджено")
	s.loadPayments()
}

func (s *PaymentsScreen) unapprovePayment(id int64) {
	ctx := context.Background()
	_, err := s.paymentService.Unapprove(ctx, payment.UnapprovePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, "Підтвердження платежу скасовано")
	s.loadPayments()
}

func (s *PaymentsScreen) applyFilters() {
	s.filteredPayments = make([]*payment.PaymentOutput, 0)

	filter := s.filterSelect.Selected
	searchQuery := s.searchEntry.Text

	for _, p := range s.payments {
		// написати фільтри
		if filter == "Всі" {
			if "" != searchQuery {
				continue
			}
		}

		s.filteredPayments = append(s.filteredPayments, p)
	}
}

// getStatsText повертає текст статистики.
func (s *PaymentsScreen) getStatsText() string {
	totalAmount := 0.0
	for _, p := range s.payments {
		totalAmount += p.Amount
	}
	return fmt.Sprintf("Показано: %d платежів | Загальна сума: %.2f грн", len(s.payments), totalAmount)
}

// updateStats оновлює статистику.
func (s *PaymentsScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
