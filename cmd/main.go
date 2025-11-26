// cmd/main.go
package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	// Keep for type reference if needed, though bootstrap handles creation
	"osbb-accounting/config"
	"osbb-accounting/infrastructure/bootstrap"
	"osbb-accounting/presentation/fyne/auth"
	"osbb-accounting/presentation/fyne/screens"
	appTheme "osbb-accounting/presentation/fyne/theme"
)

func main() {
	// 0. Завантаження конфігурації
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("Config file not found or invalid, using defaults: %v", err)
		cfg = config.Default()
	}

	// 1. Ініціалізація Fyne додатку
	myApp := app.NewWithID("com.osbb.management")
	myApp.Settings().SetTheme(&appTheme.CustomTheme{})

	mainWindow := myApp.NewWindow(cfg.App.WindowTitle)
	mainWindow.Resize(fyne.NewSize(1200, 800))
	mainWindow.CenterOnScreen()

	// 2. Bootstrap Application (DB, Repos, Services)
	application, err := bootstrap.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// 3. Створення AuthManager (Presentation Layer)
	authManager := auth.NewAuthManager(myApp, application.Services.AuthService)

	// 4. Налаштування callbacks навігації
	authManager.OnLoginSuccess(func() {
		// Передаємо контейнер сервісів у головний екран
		screens.ShowMainScreen(mainWindow, authManager, application.Services)
	})

	authManager.OnLogout(func() {
		authManager.ShowLoginScreen(mainWindow)
	})

	// 5. Старт
	authManager.ShowLoginScreen(mainWindow)

	mainWindow.SetOnClosed(func() {
		authManager.Cleanup()
	})

	mainWindow.ShowAndRun()
}
