package main

import (
	"database/sql"
	"log"
	"osbb-accounting/config"
	"osbb-accounting/domain"
	"osbb-accounting/repository"
	"osbb-accounting/repository/sqlite"
	"osbb-accounting/service"
	"osbb-accounting/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

// App представляє основний додаток ОСББ Обліку.
type App struct {
	// Конфігурація додатку
	config *config.AppConfig

	// Data Access Layer
	db            *sql.DB
	userRepo      repository.UserRepository
	apartmentRepo repository.ApartmentRepository

	// Business Logic Layer
	authService *service.AuthService

	// Поточний авторизований користувач
	currentUser *domain.User

	// UI Layer
	fyneApp fyne.App
	window  fyne.Window
}

// main - точка входу в додаток.
func main() {
	// Створюємо екземпляр додатку
	application := &App{}

	// Ініціалізуємо та запускаємо
	if err := application.Initialize(); err != nil {
		log.Fatalf("Критична помилка ініціалізації: %v", err)
	}

	application.Run()
}

// Initialize виконує ініціалізацію всіх компонентів додатку.
func (a *App) Initialize() error {
	// ============================================================
	// КРОК 1: Ініціалізація Конфігурації
	// ============================================================

	a.config = config.DefaultConfig()

	// Створюємо директорію для даних
	if err := a.config.EnsureDataDir(); err != nil {
		return err
	}

	log.Printf("✓ Конфігурація завантажена")

	// ============================================================
	// КРОК 2: Ініціалізація Data Access Layer
	// ============================================================

	// Створюємо конфігурацію БД
	dbConfig := &sqlite.DBConfig{
		DBPath:         a.config.DBPath,
		MigrationsPath: a.config.MigrationsPath,
	}

	// Ініціалізуємо базу даних
	db, err := sqlite.InitDatabase(dbConfig)
	if err != nil {
		return err
	}
	a.db = db

	log.Printf("✓ База даних ініціалізована: %s", a.config.DBPath)

	// Створюємо репозиторій користувачів
	a.userRepo = sqlite.NewSQLiteUserRepository(db)

	// ============================================================
	// КРОК 3: Ініціалізація Business Logic Layer
	// ============================================================

	// Створюємо сервіс аутентифікації
	a.authService = service.NewAuthService(a.userRepo)

	log.Printf("✓ Сервіси ініціалізовані")

	// ============================================================
	// КРОК 4: Ініціалізація UI Layer
	// ============================================================

	// Створюємо Fyne додаток
	a.fyneApp = app.NewWithID(a.config.AppID)
	//a.fyneApp.Settings().SetTheme(&customTheme{})

	// Створюємо головне вікно
	a.window = a.fyneApp.NewWindow(a.config.AppName)
	a.window.Resize(fyne.NewSize(a.config.WindowWidth, a.config.WindowHeight))
	a.window.SetFixedSize(false)
	a.window.CenterOnScreen()

	// Встановлюємо мінімальні розміри
	a.window.SetContent(container.NewMax()) // Тимчасовий контент

	log.Printf("✓ UI ініціалізовано")

	return nil
}

// Run запускає додаток і відображає екран входу.
func (a *App) Run() {
	// Показуємо екран входу
	a.showLoginScreen()

	// Запускаємо Fyne event loop
	a.window.ShowAndRun()

	// Очищення ресурсів після закриття
	a.cleanup()
}

// showLoginScreen відображає екран входу в систему.
func (a *App) showLoginScreen() {
	loginScreen := ui.NewLoginScreen(
		a.authService,
		a.window,
		a.onLoginSuccess,
	)

	a.window.SetContent(loginScreen.Render())
	a.window.SetTitle(a.config.AppName + " - Вхід")

	// Встановлюємо фокус ПІСЛЯ того, як контент додано до canvas
	loginScreen.SetFocus()
}

// onLoginSuccess викликається після успішного входу користувача.
func (a *App) onLoginSuccess(user *domain.User) {
	a.currentUser = user

	log.Printf("✓ Користувач увійшов: %s (%s)", user.Username, user.Role)

	// Перевіряємо, чи це перший вхід з дефолтним паролем
	if user.Username == "admin" {
		// Можна запропонувати змінити пароль
		// loginScreen.ShowPasswordChangeDialog(user)
	}

	// Показуємо головне вікно
	a.showMainWindow()
}

// showMainWindow відображає головне вікно додатку.
func (a *App) showMainWindow() {
	mainWindow := ui.NewMainWindow(
		a.window,
		a.currentUser,
		a.authService,
		a.onLogout,
	)

	a.window.SetContent(mainWindow.Render())
	a.window.SetTitle(a.config.AppName + " - " + a.currentUser.FullName)
}

// onLogout викликається при виході користувача з системи.
func (a *App) onLogout() {
	log.Printf("✓ Користувач вийшов: %s", a.currentUser.Username)

	a.currentUser = nil

	// Повертаємося до екрану входу
	a.showLoginScreen()
}

// cleanup виконує очищення ресурсів при закритті додатку.
func (a *App) cleanup() {
	log.Println("Закриття додатку...")

	// Закриваємо з'єднання з БД
	if a.db != nil {
		if err := sqlite.CloseDatabase(a.db); err != nil {
			log.Printf("Помилка закриття БД: %v", err)
		}
	}

	log.Println("✓ Додаток закрито")
}

// customTheme - кастомна тема для додатку (опціонально).
//type customTheme struct{}

// Реалізація інтерфейсу fyne.Theme буде додана пізніше
// для кастомізації кольорів та шрифтів
