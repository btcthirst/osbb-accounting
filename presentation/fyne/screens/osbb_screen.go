package screens

import (
	"context"
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/osbb"
	domainErrors "osbb-accounting/domain/errors"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/dialogs"
	"osbb-accounting/presentation/fyne/text"
)

type OSBBScreen struct {
	window      fyne.Window
	service     *service.OSBBService
	authManager *auth.AuthManager
	content     *fyne.Container
}

func NewOSBBScreen(window fyne.Window, service *service.OSBBService, authManager *auth.AuthManager) *OSBBScreen {
	return &OSBBScreen{
		window:      window,
		service:     service,
		authManager: authManager,
		content:     container.NewStack(),
	}
}

func (s *OSBBScreen) Render() fyne.CanvasObject {
	s.refreshData()
	return s.content
}

func (s *OSBBScreen) refreshData() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID := s.authManager.GetCurrentUser().ID

	data, err := s.service.Get(ctx, osbb.GetOSBBInput{CurrentUserID: userID})
	if err != nil {
		// Check if it's a NotFound error
		var domainErr *domainErrors.DomainError
		if errors.As(err, &domainErr) && domainErr.Code == domainErrors.CodeNotFound {
			s.showCreateView()
			return
		}

		if domainErrors.Is(err, domainErrors.ErrNotFound) {
			s.showCreateView()
			return
		}

		dialog.ShowError(err, s.window)
		s.content.Objects = []fyne.CanvasObject{
			container.NewCenter(widget.NewLabel("Помилка завантаження даних: " + err.Error())),
		}
		s.content.Refresh()
		return
	}

	s.showDetailsView(data)
}

func (s *OSBBScreen) showCreateView() {
	icon := widget.NewIcon(theme.InfoIcon())
	label := widget.NewLabel("Організація ОСББ ще не налаштована")
	label.TextStyle = fyne.TextStyle{Bold: true}

	createBtn := widget.NewButtonWithIcon("Створити ОСББ", theme.ContentAddIcon(), func() {
		dialogs.ShowOSBBFormDialog(
			s.window,
			s.service,
			s.authManager,
			nil,
			func() {
				s.refreshData()
			},
		)
	})
	createBtn.Importance = widget.HighImportance

	// Only admin can create
	if !s.authManager.HasPermission("system.all") {
		createBtn.Disable()
		createBtn.SetText("Створити ОСББ (потрібні права адміністратора)")
	}

	s.content.Objects = []fyne.CanvasObject{
		container.NewCenter(container.NewVBox(
			container.NewCenter(icon),
			container.NewCenter(label),
			container.NewCenter(createBtn),
		)),
	}
	s.content.Refresh()
}

func (s *OSBBScreen) showDetailsView(data *osbb.GetOSBBOutput) {
	// Header
	title := widget.NewLabel(text.TitleOSBBInfo)
	title.TextStyle = fyne.TextStyle{Bold: true, Monospace: true} // Larger?

	// Details Form
	form := widget.NewForm(
		widget.NewFormItem(text.LabelOSBBName, widget.NewLabel(data.Name)),
		widget.NewFormItem(text.LabelOSBBEDRPOU, widget.NewLabel(data.EDRPOU)),
		widget.NewFormItem(text.LabelOSBBAddress, widget.NewLabel(data.LegalAddress)),
		widget.NewFormItem(text.LabelOSBBChairman, widget.NewLabel(data.ChairmanName)),
	)

	if data.ActualAddress != nil {
		form.Append(text.LabelOSBBActualAddr, widget.NewLabel(*data.ActualAddress))
	}
	if data.Phone != nil {
		form.Append(text.LabelOSBBPhone, widget.NewLabel(*data.Phone))
	}
	if data.Email != nil {
		form.Append(text.LabelOSBBEmail, widget.NewLabel(*data.Email))
	}
	if data.Website != nil {
		form.Append(text.LabelOSBBWebsite, widget.NewLabel(*data.Website)) // Make it a link?
	}

	// Toolbar
	toolbar := container.NewHBox()

	if s.authManager.HasPermission("system.all") {
		editBtn := widget.NewButtonWithIcon(text.ActionEdit, theme.DocumentCreateIcon(), func() {
			dialogs.ShowOSBBFormDialog(s.window, s.service, s.authManager, data, func() {
				s.refreshData()
			})
		})
		toolbar.Add(editBtn)
	}

	refreshBtn := widget.NewButtonWithIcon(text.ActionRefresh, theme.ViewRefreshIcon(), func() {
		s.refreshData()
	})
	toolbar.Add(refreshBtn)

	mainContent := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(form),
	)

	s.content.Objects = []fyne.CanvasObject{
		container.NewBorder(toolbar, nil, nil, nil, mainContent),
	}
	s.content.Refresh()
}
