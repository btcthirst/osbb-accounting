package main

import (
	"database/sql"
	"fmt"
	"log"
	"osbb-accounting/domain"
	"osbb-accounting/repository/sqlite"
	"osbb-accounting/service"

	"golang.org/x/crypto/bcrypt"
)

// Тестовий скрипт для діагностики проблем з аутентифікацією
func main() {
	fmt.Println("=== Діагностика Аутентифікації ===\n")

	// Ініціалізуємо БД
	dbConfig := sqlite.DefaultDBConfig()
	db, err := sqlite.InitDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Помилка БД: %v", err)
	}
	defer db.Close()

	fmt.Printf("✓ БД підключена: %s\n\n", dbConfig.DBPath)

	// Перевіряємо наявність користувачів
	fmt.Println("--- Крок 1: Перевірка користувачів в БД ---")
	checkUsers(db)

	// Створюємо репозиторій та сервіс
	userRepo := sqlite.NewSQLiteUserRepository(db)
	authService := service.NewAuthService(userRepo)

	// Тестуємо хешування
	fmt.Println("\n--- Крок 2: Тест хешування BCrypt ---")
	testBCrypt()

	// Пробуємо знайти адміна
	fmt.Println("\n--- Крок 3: Пошук користувача 'admin' ---")
	user, err := userRepo.FindByUsername("admin")
	if err != nil {
		fmt.Printf("❌ Помилка: %v\n", err)
		fmt.Println("\n🔧 Спроба створити адміністратора...")
		createDefaultAdmin(db, authService)
		return
	}

	fmt.Printf("✓ Користувач знайдений:\n")
	fmt.Printf("  ID: %d\n", user.ID)
	fmt.Printf("  Username: %s\n", user.Username)
	fmt.Printf("  Role: %s\n", user.Role)
	fmt.Printf("  IsActive: %v\n", user.IsActive)
	fmt.Printf("  HashedPassword: %s\n", user.HashedPassword[:60])

	// Тестуємо аутентифікацію
	fmt.Println("\n--- Крок 4: Тест аутентифікації ---")
	testPasswords := []string{"admin123", "admin", "password", "Admin123"}

	for _, pass := range testPasswords {
		fmt.Printf("\nПароль: '%s'\n", pass)

		// Перевірка через BCrypt напряму
		err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(pass))
		if err == nil {
			fmt.Printf("  ✓ BCrypt: Співпадає\n")
		} else {
			fmt.Printf("  ❌ BCrypt: Не співпадає (%v)\n", err)
		}

		// Перевірка через AuthService
		authUser, err := authService.Authenticate("admin", pass)
		if err == nil {
			fmt.Printf("  ✓ AuthService: Успішно (User: %s)\n", authUser.FullName)
		} else {
			fmt.Printf("  ❌ AuthService: Помилка (%v)\n", err)
		}
	}

	fmt.Println("\n=== Діагностика завершена ===")
}

// checkUsers виводить список всіх користувачів з БД
func checkUsers(db *sql.DB) {
	rows, err := db.Query("SELECT id, username, role, is_active FROM users")
	if err != nil {
		fmt.Printf("❌ Помилка запиту: %v\n", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var username, role string
		var isActive bool

		err := rows.Scan(&id, &username, &role, &isActive)
		if err != nil {
			fmt.Printf("❌ Помилка сканування: %v\n", err)
			continue
		}

		count++
		fmt.Printf("%d. ID=%d, Username=%s, Role=%s, Active=%v\n",
			count, id, username, role, isActive)
	}

	if count == 0 {
		fmt.Println("⚠️ Користувачів не знайдено!")
	} else {
		fmt.Printf("✓ Знайдено користувачів: %d\n", count)
	}
}

// testBCrypt тестує функції хешування
func testBCrypt() {
	password := "admin123"

	// Генеруємо хеш
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Printf("❌ Помилка генерації хешу: %v\n", err)
		return
	}

	fmt.Printf("Пароль: %s\n", password)
	fmt.Printf("Хеш: %s\n", string(hash))

	// Перевіряємо
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == nil {
		fmt.Println("✓ Хеш перевірено успішно")
	} else {
		fmt.Printf("❌ Помилка перевірки: %v\n", err)
	}
}

// createDefaultAdmin створює дефолтного адміністратора
func createDefaultAdmin(db *sql.DB, authService *service.AuthService) {
	password := "admin123"

	// Генеруємо хеш
	hashedPassword, err := authService.HashPassword(password)
	if err != nil {
		fmt.Printf("❌ Помилка хешування: %v\n", err)
		return
	}

	// Створюємо користувача
	query := `
		INSERT INTO users (username, hashed_password, role, full_name, is_active)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query, "admin", hashedPassword, domain.RoleAdmin, "Системний Адміністратор", true)
	if err != nil {
		fmt.Printf("❌ Помилка створення: %v\n", err)
		return
	}

	id, _ := result.LastInsertId()
	fmt.Printf("✓ Адміністратора створено з ID: %d\n", id)
	fmt.Println("\nТепер спробуйте увійти: admin / admin123")
}
