// presentation/fyne/auth/auth_manager.go
package auth

import (
	"context"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/auth"
	"osbb-accounting/presentation/fyne/screens"
)

// AuthManager управляє станом аутентифікації в додатку.
type AuthManager struct {
	app         fyne.App
	authService *service.AuthService

	// Стан аутентифікації
	mu            sync.RWMutex
	isLoggedIn    bool
	sessionToken  string
	currentUserID int64
	currentUser   *auth.UserInfo
	roles         []string
	permissions   []string

	// Callbacks
	onLoginSuccess func()
	onLogout       func()

	// Session validation
	validationTicker *time.Ticker
	stopValidation   chan bool
}

// NewAuthManager створює новий AuthManager.
func NewAuthManager(application fyne.App, authService *service.AuthService) *AuthManager {
	return &AuthManager{
		app:            application,
		authService:    authService,
		isLoggedIn:     false,
		stopValidation: make(chan bool),
	}
}

// ShowLoginScreen показує екран входу.
func (m *AuthManager) ShowLoginScreen(window fyne.Window) {
	loginScreen := screens.NewLoginScreen(window, m.authService)

	loginScreen.OnLoginSuccess(func(sessionToken string, userID int64) {
		m.handleLoginSuccess(sessionToken, userID)
	})

	loginScreen.OnRegisterClick(func() {
		m.ShowRegisterScreen(window)
	})

	window.SetContent(loginScreen.Render())
	loginScreen.Focus()
}

// ShowRegisterScreen показує екран реєстрації.
func (m *AuthManager) ShowRegisterScreen(window fyne.Window) {
	registerScreen := screens.NewRegisterScreen(window, m.authService)

	registerScreen.OnRegisterSuccess(func() {
		// Після успішної реєстрації повертаємось на екран входу
		m.ShowLoginScreen(window)
	})

	registerScreen.OnBackClick(func() {
		m.ShowLoginScreen(window)
	})

	window.SetContent(registerScreen.Render())
	registerScreen.Focus()
}

// handleLoginSuccess обробляє успішний вхід.
func (m *AuthManager) handleLoginSuccess(sessionToken string, userID int64) {
	m.mu.Lock()
	m.isLoggedIn = true
	m.sessionToken = sessionToken
	m.currentUserID = userID
	m.mu.Unlock()

	// Завантажуємо інформацію про користувача
	m.loadUserInfo()

	// Запускаємо періодичну перевірку сесії
	m.startSessionValidation()

	// Викликаємо callback
	if m.onLoginSuccess != nil {
		m.onLoginSuccess()
	}
}

// loadUserInfo завантажує інформацію про поточного користувача.
func (m *AuthManager) loadUserInfo() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := m.authService.ValidateSession(ctx, auth.ValidateSessionInput{
		SessionToken: m.GetSessionToken(),
	})

	if err != nil || !output.Valid {
		// Сесія невалідна - виконуємо logout
		m.Logout()
		return
	}

	m.mu.Lock()
	m.currentUser = output.User
	m.roles = output.Roles
	m.permissions = output.Permissions
	m.mu.Unlock()
}

// startSessionValidation запускає періодичну перевірку сесії.
func (m *AuthManager) startSessionValidation() {
	// Перевіряємо сесію кожні 5 хвилин
	m.validationTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-m.validationTicker.C:
				m.validateSession()
			case <-m.stopValidation:
				return
			}
		}
	}()
}

// validateSession перевіряє валідність поточної сесії.
func (m *AuthManager) validateSession() {
	if !m.IsLoggedIn() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := m.authService.ValidateSession(ctx, auth.ValidateSessionInput{
		SessionToken: m.GetSessionToken(),
	})

	if err != nil || !output.Valid {
		// Сесія невалідна - виконуємо logout
		m.Logout()
	}
}

// Logout виконує вихід з системи.
func (m *AuthManager) Logout() {
	if !m.IsLoggedIn() {
		return
	}

	// Зупиняємо валідацію сесії
	if m.validationTicker != nil {
		m.validationTicker.Stop()
		m.stopValidation <- true
	}

	// Викликаємо API logout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.authService.Logout(ctx, auth.LogoutUserInput{
		SessionToken: m.GetSessionToken(),
	})
	if err != nil {
		log.Fatal("logout fail: ", err)
	}

	// Очищуємо стан
	m.mu.Lock()
	m.isLoggedIn = false
	m.sessionToken = ""
	m.currentUserID = 0
	m.currentUser = nil
	m.roles = nil
	m.permissions = nil
	m.mu.Unlock()

	// Викликаємо callback
	if m.onLogout != nil {
		m.onLogout()
	}
}

// IsLoggedIn повертає чи користувач авторизований.
func (m *AuthManager) IsLoggedIn() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isLoggedIn
}

// GetSessionToken повертає токен поточної сесії.
func (m *AuthManager) GetSessionToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionToken
}

// GetCurrentUserID повертає ID поточного користувача.
func (m *AuthManager) GetCurrentUserID() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentUserID
}

// GetCurrentUser повертає інформацію про поточного користувача.
func (m *AuthManager) GetCurrentUser() *auth.UserInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentUser
}

// GetRoles повертає ролі поточного користувача.
func (m *AuthManager) GetRoles() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.roles
}

// GetPermissions повертає дозволи поточного користувача.
func (m *AuthManager) GetPermissions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.permissions
}

// HasPermission перевіряє чи має користувач конкретний дозвіл.
func (m *AuthManager) HasPermission(permissionCode string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, perm := range m.permissions {
		if perm == permissionCode || perm == "system.all" {
			return true
		}
	}
	return false
}

// HasRole перевіряє чи має користувач конкретну роль.
func (m *AuthManager) HasRole(roleName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, role := range m.roles {
		if role == roleName {
			return true
		}
	}
	return false
}

// OnLoginSuccess встановлює callback для успішного входу.
func (m *AuthManager) OnLoginSuccess(callback func()) {
	m.onLoginSuccess = callback
}

// OnLogout встановлює callback для виходу.
func (m *AuthManager) OnLogout(callback func()) {
	m.onLogout = callback
}

// RequireLogin перевіряє чи користувач авторизований.
// Якщо ні - показує екран входу.
func (m *AuthManager) RequireLogin(window fyne.Window) bool {
	if !m.IsLoggedIn() {
		m.ShowLoginScreen(window)
		return false
	}
	return true
}

// Cleanup очищує ресурси при закритті додатку.
func (m *AuthManager) Cleanup() {
	if m.validationTicker != nil {
		m.validationTicker.Stop()
	}
	if m.stopValidation != nil {
		close(m.stopValidation)
	}
}
