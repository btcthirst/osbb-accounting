// presentation/fyne/screens/contractor_payment_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/contractorpayment"
	"osbb-accounting/presentation/fyne/common"
	"osbb-accounting/presentation/fyne/dialogs"
	"osbb-accounting/presentation/fyne/text"
)

// ContractorPaymentScreen - екран списку платежів від контрагентів
type ContractorPaymentScreen struct {
	window                   fyne.Window
	contractorPaymentService *service.ContractorPaymentService
	contractorService        *service.ContractorService
	authManager              interface {
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
	payments         []*contractorpayment.ContractorPaymentOutput
	filteredPayments []*contractorpayment.ContractorPaymentOutput
	currentPage      int
	pageSize         int
	totalCount       int64
}

type ContractorPaymentFilterType string

const (
	ContractorPaymentFilterAll      ContractorPaymentFilterType = text.FilterContractorPaymentAll
	ContractorPaymentFilterCurrent  ContractorPaymentFilterType = text.FilterContractorPaymentCurrent
	ContractorPaymentFilterPrevious ContractorPaymentFilterType = text.FilterContractorPaymentPrevious
)

// NewContractorPaymentScreen створює новий екран
func NewContractorPaymentScreen(
	window fyne.Window,
	contractorPaymentService *service.ContractorPaymentService,
	contractorService *service.ContractorService,
	authManager interface {
		GetCurrentUserID() int64
		HasPermission(permission string) bool
	},
) *ContractorPaymentScreen {
	s := &ContractorPaymentScreen{
		window:                   window,
		contractorPaymentService: contractorPaymentService,
		contractorService:        contractorService,
		authManager:              authManager,
	}
	s.buildUI()
	return s
}

func (s *ContractorPaymentScreen) buildUI() {
	// Buttons
	s.createButton = newCreateButton(text.ActionAddPayment, func() {
		s.showPaymentDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadPayments()
	})

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.filteredPayments) + 1, len(text.PaymentsTableHeaders) // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers "ID", "Контрагент", "Дата", "Період", "Метод", "Сума", "Дії"
				renderTableHeader(label, text.ContractorPaymentsTableHeaders, id.Col)
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
			case 1: // Contractor (будемо заповнювати)
				label.SetText(fmt.Sprintf(text.LabelContractorID, p.ContractorID))
			case 2: // Date
				label.SetText(p.PaymentDate.Format("02.01.2006"))
			case 3: // Period
				if p.PeriodMonth != nil && p.PeriodYear != nil {
					label.SetText(fmt.Sprintf("%02d/%d", *p.PeriodMonth, *p.PeriodYear))
				} else {
					label.SetText("-")
				}
			case 4: // Method
				label.SetText(p.MethodName)
			case 5: // Amount
				label.SetText(fmt.Sprintf("%.2f грн", p.Amount))
				label.TextStyle = fyne.TextStyle{Bold: true}
			case 6: // Actions
				renderActionColumn(label)
			}
		},
	)

	// Column widths
	setupTableColumnWidths(s.table, map[int]float32{
		0: 50,  // ID
		1: 200, // Contractor
		2: 100, // Date
		3: 80,  // Period
		4: 150, // Method
		5: 100, // Amount
		6: 50,  // Actions
	})

	s.table.OnSelected = func(id widget.TableCellID) {
		s.table.Unselect(id) // Don't keep selection
		if id.Row == 0 {
			return
		}
		if id.Col == 6 {
			s.showActionsMenu(id.Row - 1)
		}
	}

	// Filters
	s.searchEntry = newSearchEntry(text.SearchPlaceholder, func(_ string) { s.applyFilters() })

	s.filterSelect = newFilterSelect([]string{
		string(ContractorPaymentFilterAll),
		string(ContractorPaymentFilterCurrent),
		string(ContractorPaymentFilterPrevious),
	}, func(selected string) {
		s.applyFilters()
	})

	// Stats
	s.statsLabel = widget.NewLabel("")

	// Trigger initial load
	s.filterSelect.SetSelected(string(ContractorPaymentFilterAll))
	s.loadPayments()
}

func (s *ContractorPaymentScreen) Render() fyne.CanvasObject {
	toolbar := buildStandardToolbar(s.createButton, s.refreshButton, s.searchEntry, s.filterSelect)
	return buildStandardLayout(toolbar, s.table, s.statsLabel)
}

func (s *ContractorPaymentScreen) loadPayments() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	input := contractorpayment.ListContractorPaymentsInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         1000, // Load all for client-side filtering
		SearchQuery:   "",   // We filter on client side for now to match other screens
		OrderDesc:     true,
		OrderBy:       "payment_date",
	}

	output, err := s.contractorPaymentService.List(ctx, input)
	if err != nil {
		common.ShowError(s.window, fmt.Errorf(text.MsgErrorLoading, err))
		return
	}

	s.payments = output.Payments
	s.totalCount = output.Total
	s.applyFilters()
}

func (s *ContractorPaymentScreen) showPaymentDialog(existing *contractorpayment.ContractorPaymentOutput) {
	d := dialogs.NewContractorPaymentFormDialog(
		s.window,
		s.contractorPaymentService,
		s.contractorService,
		s.authManager,
		existing,
	)
	d.SetOnSaved(func() {
		s.loadPayments()
	})
	d.Show()
}

func (s *ContractorPaymentScreen) showActionsMenu(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(s.payments) {
		return
	}
	p := s.payments[rowIndex]

	actions := buildStandardActions(
		func() {
			s.showPaymentDialog(p)
		},
		func() {
			s.confirmDelete(p)
		},
		nil, // No details view yet
	)

	common.ShowActionsMenu(s.window, fmt.Sprintf(text.TitlePaymentActions, p.ID), actions)
}

func (s *ContractorPaymentScreen) confirmDelete(p *contractorpayment.ContractorPaymentOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf(text.LabelPaymentID, p.ID), func() {
		s.deletePayment(p.ID)
	})
}

func (s *ContractorPaymentScreen) deletePayment(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.contractorPaymentService.Delete(ctx, contractorpayment.DeleteContractorPaymentInput{
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

func (s *ContractorPaymentScreen) applyFilters() {
	s.filteredPayments = make([]*contractorpayment.ContractorPaymentOutput, 0)

	filter := s.filterSelect.Selected
	searchQuery := s.searchEntry.Text

	s.filteredPayments = FilterList(
		s.payments,
		searchQuery,
		filter,
		func(p *contractorpayment.ContractorPaymentOutput, query string) bool {
			return contains(p.Purpose, query) ||
				contains(fmt.Sprintf("%.2f", p.Amount), query) ||
				(p.ReceiptNumber != nil && contains(*p.ReceiptNumber, query)) ||
				(p.Notes != nil && contains(*p.Notes, query)) ||
				contains(p.MethodName, query)
		},
		func(p *contractorpayment.ContractorPaymentOutput, filter string) bool {
			switch ContractorPaymentFilterType(filter) {
			case ContractorPaymentFilterAll:
				return true
			case ContractorPaymentFilterCurrent:
				now := time.Now()
				return p.PaymentDate.Month() == now.Month() && p.PaymentDate.Year() == now.Year()
			case ContractorPaymentFilterPrevious:
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
func (s *ContractorPaymentScreen) getStatsText() string {
	totalAmount := 0.0
	for _, p := range s.payments {
		totalAmount += p.Amount
	}
	return fmt.Sprintf(text.MsgContractorPaymentStats, len(s.payments), totalAmount)
}

// updateStats оновлює статистику.
func (s *ContractorPaymentScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
