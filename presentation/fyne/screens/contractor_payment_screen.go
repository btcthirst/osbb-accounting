// presentation/fyne/screens/contractor_payment_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/contractorpayment"
	"osbb-accounting/presentation/fyne/common"
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
	ContractorPaymentFilterAll      ContractorPaymentFilterType = "Всі платежі"
	ContractorPaymentFilterCurrent  ContractorPaymentFilterType = "Поточний місяць"
	ContractorPaymentFilterPrevious ContractorPaymentFilterType = "Минулий місяць"
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
	s.createButton = newCreateButton("Додати платіж", func() {
		s.showPaymentDialog(nil)
	})

	s.refreshButton = newRefreshButton(func() {
		s.loadPayments()
	})

	// Table
	s.table = widget.NewTable(
		func() (int, int) {
			return len(s.filteredPayments) + 1, 7 // +1 for header
		},
		func() fyne.CanvasObject {
			return newTableTemplate()
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				// Headers
				renderTableHeader(label,
					[]string{"ID", "Контрагент", "Дата", "Період", "Метод", "Сума", "Дії"}, id.Col)
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
				label.SetText(fmt.Sprintf("Контрагент #%d", p.ContractorID))
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
	s.searchEntry = newSearchEntry("Пошук...", func(_ string) { s.applyFilters() })

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
	ctx := context.Background()
	input := contractorpayment.ListContractorPaymentsInput{
		CurrentUserID: s.authManager.GetCurrentUserID(),
		Limit:         100,
		SearchQuery:   s.searchEntry.Text,
		OrderDesc:     true,
		OrderBy:       "payment_date",
	}

	output, err := s.contractorPaymentService.List(ctx, input)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.payments = output.Payments
	s.totalCount = output.Total
	s.applyFilters()
}

func (s *ContractorPaymentScreen) showPaymentDialog(existing *contractorpayment.ContractorPaymentOutput) {
	// TODO: Створити ContractorPaymentFormDialog
	common.ShowInformation(s.window, "TODO", "ContractorPaymentFormDialog ще не реалізовано")
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

	common.ShowActionsMenu(s.window, fmt.Sprintf("Платіж #%d", p.ID), actions)
}

func (s *ContractorPaymentScreen) confirmDelete(p *contractorpayment.ContractorPaymentOutput) {
	showConfirmDeleteDialog(s.window, fmt.Sprintf("платіж #%d", p.ID), func() {
		s.deletePayment(p.ID)
	})
}

func (s *ContractorPaymentScreen) deletePayment(id int64) {
	ctx := context.Background()
	_, err := s.contractorPaymentService.Delete(ctx, contractorpayment.DeleteContractorPaymentInput{
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
	return fmt.Sprintf("Показано: %d платежів | Загальна сума: %.2f грн", len(s.payments), totalAmount)
}

// updateStats оновлює статистику.
func (s *ContractorPaymentScreen) updateStats() {
	s.statsLabel.SetText(s.getStatsText())
}
