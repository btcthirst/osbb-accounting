// presentation/fyne/screens/dashboard_screen.go
package screens

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/charge"
	"osbb-accounting/application/usecase/payment"
	"osbb-accounting/presentation/fyne/text"
)

// DashboardScreen - головний екран з дашбордом
type DashboardScreen struct {
	window         fyne.Window
	chargeService  *service.ChargeService
	paymentService *service.PaymentService
	authManager    interface {
		GetCurrentUserID() int64
	}

	// UI Elements
	content *fyne.Container

	// Summary Labels
	totalChargesLabel    *widget.Label
	totalPaymentsLabel   *widget.Label
	outstandingDebtLabel *widget.Label

	// Stats Containers
	paymentMethodsContainer *fyne.Container
	chargeTypesContainer    *fyne.Container
}

// NewDashboardScreen створює новий екран дашборду
func NewDashboardScreen(
	window fyne.Window,
	chargeService *service.ChargeService,
	paymentService *service.PaymentService,
	authManager interface {
		GetCurrentUserID() int64
	},
) *DashboardScreen {
	s := &DashboardScreen{
		window:         window,
		chargeService:  chargeService,
		paymentService: paymentService,
		authManager:    authManager,
	}
	s.buildUI()
	return s
}

func (s *DashboardScreen) buildUI() {
	// 1. Summary Cards
	s.totalChargesLabel = widget.NewLabelWithStyle("0.00 грн", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	s.totalPaymentsLabel = widget.NewLabelWithStyle("0.00 грн", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	s.outstandingDebtLabel = widget.NewLabelWithStyle("0.00 грн", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	summaryCard := container.NewGridWithColumns(3,
		s.createSummaryCard(text.LabelTotalCharges, s.totalChargesLabel, theme.ColorNamePrimary),
		s.createSummaryCard(text.LabelTotalPayments, s.totalPaymentsLabel, theme.ColorNameSuccess),
		s.createSummaryCard(text.LabelOutstandingDebt, s.outstandingDebtLabel, theme.ColorNameError),
	)

	// 2. Statistics Details
	s.paymentMethodsContainer = container.NewVBox()
	s.chargeTypesContainer = container.NewVBox()

	detailsContainer := container.NewGridWithColumns(2,
		widget.NewCard(text.TitlePaymentMethods, "", s.paymentMethodsContainer),
		widget.NewCard(text.TitleChargeTypes, "", s.chargeTypesContainer),
	)

	// Refresh Button
	refreshBtn := widget.NewButtonWithIcon(text.ActionRefresh+" дані", theme.ViewRefreshIcon(), func() {
		s.loadStatistics()
	})

	// Layout
	s.content = container.NewVBox(
		widget.NewLabelWithStyle(text.TitleFinanceOverview, fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		summaryCard,
		widget.NewSeparator(),
		detailsContainer,
		widget.NewSeparator(),
		container.NewCenter(refreshBtn),
	)

	s.loadStatistics()
}

func (s *DashboardScreen) createSummaryCard(title string, valueLabel *widget.Label, colorName fyne.ThemeColorName) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	bg.StrokeColor = theme.Color(colorName)
	bg.StrokeWidth = 2
	bg.CornerRadius = 5

	return container.NewStack(
		bg,
		container.NewPadded(
			container.NewVBox(
				widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{}),
				valueLabel,
			),
		),
	)
}

func (s *DashboardScreen) Render() fyne.CanvasObject {
	return container.NewPadded(s.content)
}

func (s *DashboardScreen) loadStatistics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID := s.authManager.GetCurrentUserID()

	// 1. Charge Statistics
	chargeStats, err := s.chargeService.GetStatistics(ctx, charge.GetChargeStatisticsInput{
		CurrentUserID: userID,
	})
	if err != nil {
		fmt.Printf("Error loading charge stats: %v\n", err)
	}

	// 2. Payment Statistics
	paymentStats, err := s.paymentService.GetStatistics(ctx, payment.GetPaymentStatisticsInput{
		CurrentUserID: userID,
	})
	if err != nil {
		fmt.Printf("Error loading payment stats: %v\n", err)
	}

	// Update UI
	if chargeStats != nil && paymentStats != nil {
		// Summary
		s.totalChargesLabel.SetText(fmt.Sprintf("%.2f грн", chargeStats.TotalAmount))
		s.totalPaymentsLabel.SetText(fmt.Sprintf("%.2f грн", paymentStats.TotalAmount))

		debt := chargeStats.TotalAmount - paymentStats.TotalAmount
		s.outstandingDebtLabel.SetText(fmt.Sprintf("%.2f грн", debt))

		// Color logic for debt
		if debt > 0 {
			s.outstandingDebtLabel.TextStyle = fyne.TextStyle{Bold: true}
			// Note: Fyne labels don't support direct color change easily without custom renderer or markup,
			// but we used a colored border card container.
		}

		// Payment Methods
		s.paymentMethodsContainer.Objects = nil
		s.paymentMethodsContainer.Add(s.createStatRow(text.LabelMethodCash, paymentStats.AmountByCash, paymentStats.TotalAmount))
		s.paymentMethodsContainer.Add(s.createStatRow(text.LabelMethodCard, paymentStats.AmountByCard, paymentStats.TotalAmount))
		s.paymentMethodsContainer.Add(s.createStatRow(text.LabelMethodBank, paymentStats.AmountByBankTransfer, paymentStats.TotalAmount))
		s.paymentMethodsContainer.Add(s.createStatRow(text.LabelMethodOther, paymentStats.AmountByOther, paymentStats.TotalAmount))
		s.paymentMethodsContainer.Refresh()

		// Charge Types
		s.chargeTypesContainer.Objects = nil
		// Iterate over map (order is random, so maybe convert to slice if order matters, but for now ok)
		for cType, stats := range chargeStats.ByType {
			s.chargeTypesContainer.Add(s.createStatRow(cType.GetDisplayName(), stats.TotalAmount, chargeStats.TotalAmount))
		}
		s.chargeTypesContainer.Refresh()
	}
}

func (s *DashboardScreen) createStatRow(name string, amount, total float64) fyne.CanvasObject {
	percentage := 0.0
	if total > 0 {
		percentage = (amount / total) * 100
	}

	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.SetValue(percentage)
	progressBar.TextFormatter = func() string { return "" } // Hide text on bar

	label := widget.NewLabel(fmt.Sprintf("%s: %.2f грн (%.1f%%)", name, amount, percentage))

	return container.NewVBox(
		label,
		progressBar,
	)
}
