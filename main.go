package main

import (
	"fmt"
	"log"
	"osbb-accounting/domain"
	"osbb-accounting/repository/sqlite"
	"osbb-accounting/service"
)

// main - точка входу в додаток ОСББ Обліку.
// Демонструє правильну ініціалізацію архітектури та використання сервісів.
func main() {
	// ============================================================
	// КРОК 1: Ініціалізація Data Access Layer (Repository)
	// ============================================================

	// Отримуємо конфігурацію БД за замовчуванням
	dbConfig := sqlite.DefaultDBConfig()

	// Ініціалізуємо з'єднання з БД та виконуємо міграції
	db, err := sqlite.InitDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Критична помилка: не вдалося ініціалізувати БД: %v", err)
	}
	defer sqlite.CloseDatabase(db)

	fmt.Println("✓ База даних успішно ініціалізована")

	// Створюємо SQLite репозиторій користувачів
	// ВАЖЛИВО: Це ЄДИНЕ місце в коді, де використовується конкретна
	// реалізація (SQLite). Весь інший код працює через інтерфейс.
	userRepo := sqlite.NewSQLiteUserRepository(db)

	// ============================================================
	// КРОК 2: Ініціалізація Business Logic Layer (Services)
	// ============================================================

	// Створюємо сервіс аутентифікації, передаючи йому репозиторій
	authService := service.NewAuthService(userRepo)

	fmt.Println("✓ Сервіси успішно ініціалізовані")

	// ============================================================
	// КРОК 3: Демонстрація Використання (Приклад Аутентифікації)
	// ============================================================

	// Приклад 1: Спроба входу з дефолтним адміном
	fmt.Println("\n--- Тестування Аутентифікації ---")

	user, err := authService.Authenticate("admin", "admin123")
	if err != nil {
		log.Printf("Помилка аутентифікації: %v", err)
	} else {
		fmt.Printf("✓ Успішний вхід!\n")
		fmt.Printf("  Користувач: %s\n", user.Username)
		fmt.Printf("  Повне ім'я: %s\n", user.FullName)
		fmt.Printf("  Роль: %s\n", user.Role)
		fmt.Printf("  Є адміном: %v\n", user.IsAdmin())
		fmt.Printf("  Може керувати фінансами: %v\n", user.CanManageFinances())
	}

	// Приклад 2: Створення нового користувача (Бухгалтер)
	fmt.Println("\n--- Створення Нового Користувача ---")

	// Хешуємо пароль перед збереженням
	hashedPassword, err := authService.HashPassword("password123")
	if err != nil {
		log.Printf("Помилка хешування пароля: %v", err)
	} else {
		// Створюємо нового користувача
		newUser := &domain.User{
			Username:       "accountant1",
			HashedPassword: hashedPassword,
			Role:           domain.RoleAccountant,
			FullName:       "Олена Петренко",
		}

		// Зберігаємо через репозиторій
		if err := userRepo.SaveUser(newUser); err != nil {
			log.Printf("Помилка створення користувача: %v", err)
		} else {
			fmt.Printf("✓ Користувач створений з ID: %d\n", newUser.ID)
		}

		// Тестуємо вхід з новим користувачем
		accountant, err := authService.Authenticate("accountant1", "password123")
		if err != nil {
			log.Printf("Помилка входу: %v", err)
		} else {
			fmt.Printf("✓ Бухгалтер успішно увійшов: %s\n", accountant.FullName)
		}
	}

	// Приклад 3: Зміна пароля
	fmt.Println("\n--- Зміна Пароля ---")

	if user != nil {
		err := authService.ChangePassword(user.ID, "admin123", "newSecurePass456")
		if err != nil {
			log.Printf("Помилка зміни пароля: %v", err)
		} else {
			fmt.Println("✓ Пароль успішно змінено")

			// Перевіряємо вхід з новим паролем
			_, err := authService.Authenticate("admin", "newSecurePass456")
			if err != nil {
				log.Printf("Помилка входу з новим паролем: %v", err)
			} else {
				fmt.Println("✓ Вхід з новим паролем успішний")
			}
		}
	}

	// Приклад 4: Отримання всіх користувачів
	fmt.Println("\n--- Список Користувачів ---")

	users, err := userRepo.FindAll()
	if err != nil {
		log.Printf("Помилка отримання користувачів: %v", err)
	} else {
		fmt.Printf("Всього користувачів: %d\n", len(users))
		for i, u := range users {
			fmt.Printf("%d. %s (%s) - %s\n", i+1, u.FullName, u.Username, u.Role)
		}
	}

	// ============================================================
	// КРОК 4: Запуск UI (Fyne)
	// ============================================================

	// TODO: Ініціалізація Fyne додатку
	// app := fyne.NewApp()
	// window := app.NewWindow("ОСББ Облік")
	//
	// loginScreen := ui.NewLoginScreen(authService)
	// window.SetContent(loginScreen)
	// window.ShowAndRun()

	fmt.Println("\n✓ Демонстрація завершена")
	fmt.Println("\nНаступні кроки:")
	fmt.Println("1. Реалізувати UI з Fyne (екран входу)")
	fmt.Println("2. Додати модулі для квартир та платежів")
	fmt.Println("3. Реалізувати звітність та розрахунки")
	fmt.Println("4. Додати експорт у Excel")
}
