// presentation/fyne/screens/payments_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/payment"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/dialogs"
	"osbb-accounting/presentation/fyne/text"
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

type PaymentFilterType string

const (
	PaymentFilterAll         PaymentFilterType = text.FilterPaymentAll
	PaymentFilterCurrent     PaymentFilterType = text.FilterPaymentCurrent
	PaymentFilterPrevious    PaymentFilterType = text.FilterPaymentPrevious
	PaymentFilterNotApproved PaymentFilterType = text.FilterPaymentNotApproved
	PaymentFilterApproved    PaymentFilterType = text.FilterPaymentApproved
)

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
	s.createButton = newCreateButton(text.ActionAdd, func() {
		s.showPaymentDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadPayments()
	})

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.filteredPayments) + 1, 8 // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				renderTableHeader(label, text.PaymentsTableHeaders, id.Col)
				return
			}

			// Data
			if id.Row-1 >= len(s.filteredPayments) {
				label.SetText("")
				return
			}
			p := s.filteredPayments[id.Row-1]

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
					label.SetText(text.LabelYes)
				} else {
					label.SetText(text.LabelNo)
				}
			case 6: // Receipt
				if p.ReceiptNumber != nil {
					label.SetText(*p.ReceiptNumber)
				} else {
					label.SetText("")
				}
			case 7: // Actions
				renderActionColumn(label)
			}
		},
	)

	// Column widths
	setupTableColumnWidths(s.table, map[int]float32{
		0: 50,  // ID
		1: 100, // Date
		2: 80,  // Period
		3: 150, // Method
		4: 100, // Amount
		5: 80,  // Approved
		6: 100, // Receipt
		7: 50,  // Actions
	})

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
	s.searchEntry = newSearchEntry(text.SearchPlaceholder, func(_ string) { s.applyFilters() })

	s.filterSelect = newFilterSelect([]string{
		string(PaymentFilterAll),
		string(PaymentFilterCurrent),
		string(PaymentFilterPrevious),
		string(PaymentFilterNotApproved),
		string(PaymentFilterApproved),
	}, func(selected string) {
		s.applyFilters()
	})

	// Stats
	s.statsLabel = widget.NewLabel("")

	// Trigger initial load
	s.filterSelect.SetSelected(string(PaymentFilterAll))
	s.loadPayments()
}

func (s *PaymentsScreen) Render() fyne.CanvasObject {
	toolbar := buildStandardToolbar(s.createButton, s.refreshButton, s.searchEntry, s.filterSelect)
	return buildStandardLayout(toolbar, s.table, s.statsLabel)
}

func (s *PaymentsScreen) loadPayments() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	input := payment.ListPaymentsInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
		OrderDesc:     true,
		OrderBy:       "payment_date",
	}

	output, err := s.paymentService.List(ctx, input)
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorPaymentLoading, err))
		return
	}

	s.payments = output.Payments
	s.totalCount = output.Total
	s.applyFilters()
}

func (s *PaymentsScreen) showPaymentDialog(existing *payment.PaymentOutput) {
	d := dialogs.NewPaymentFormDialog(
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

	actions := buildStandardActions(
		func() {
			if p.IsApproved {
				common.ShowInformation(s.window, "Інформація", text.MsgErrorEditApproved)
				return
			}
			s.showPaymentDialog(p)
		},
		func() {
			if p.IsApproved {
				common.ShowInformation(s.window, "Інформація", text.MsgErrorDeleteApproved)
				return
			}
			s.confirmDelete(p)
		},
		nil, // No details view yet
		nil,
	)

	if p.IsApproved {
		actions = append(actions, common.Action{
			Label: text.ActionUnapprove,
			OnTap: func() { s.unapprovePayment(p.ID) },
		})
	} else {
		actions = append(actions, common.Action{
			Label: text.ActionApprove,
			OnTap: func() { s.approvePayment(p.ID) },
		})
	}

	common.ShowActionsMenu(s.window, fmt.Sprintf(text.TitlePaymentActions, p.ID), actions)
}

func (s *PaymentsScreen) confirmDelete(p *payment.PaymentOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf(text.LabelPaymentID, p.ID), func() {
		s.deletePayment(p.ID)
	})
}

func (s *PaymentsScreen) deletePayment(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.paymentService.Delete(ctx, payment.DeletePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorDeleting, err))
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessPaymentDeleted)
	s.loadPayments()
}

func (s *PaymentsScreen) approvePayment(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.paymentService.Approve(ctx, payment.ApprovePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessPaymentApproved)
	s.loadPayments()
}

func (s *PaymentsScreen) unapprovePayment(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.paymentService.Unapprove(ctx, payment.UnapprovePaymentInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		PaymentID:     id,
	})
	if err != nil {
		common.ShowError(s.window, err)
		return
	}
	common.ShowSuccess(s.window, text.MsgSuccessPaymentUnapproved)
	s.loadPayments()
}

func (s *PaymentsScreen) applyFilters() {
	s.filteredPayments = make([]*payment.PaymentOutput, 0)

	filter := s.filterSelect.Selected
	searchQuery := s.searchEntry.Text

	s.filteredPayments = FilterList(
		s.payments,
		searchQuery,
		filter,
		func(p *payment.PaymentOutput, query string) bool {
			return contains(p.PaymentPurpose, query) ||
				contains(fmt.Sprintf("%.2f", p.Amount), query) ||
				(p.ReceiptNumber != nil && contains(*p.ReceiptNumber, query)) ||
				(p.Notes != nil && contains(*p.Notes, query)) ||
				contains(p.MethodName, query)
		},
		func(p *payment.PaymentOutput, filter string) bool {
			switch PaymentFilterType(filter) {
			case PaymentFilterAll:
				return true
			case PaymentFilterApproved:
				return p.IsApproved
			case PaymentFilterNotApproved:
				return !p.IsApproved
			// TODO: Реалізувати фільтрацію за датою для PaymentFilterCurrent та PaymentFilterPrevious
			// Це вимагає парсингу дати платежу, що краще робити на рівні сервісу або тут, якщо є доступ до time.Time
			case PaymentFilterCurrent:
				// Приклад: якщо p.PaymentDate доступна як time.Time
				now := time.Now()
				return p.PaymentDate.Month() == now.Month() && p.PaymentDate.Year() == now.Year()
			case PaymentFilterPrevious:
				prev := time.Now().AddDate(0, -1, 0)
				return p.PaymentDate.Month() == prev.Month() && p.PaymentDate.Year() == prev.Year()
			}
			return true
		},
	)
	s.table.Refresh()
	s.updateStats()
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
