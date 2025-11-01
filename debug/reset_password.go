package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"osbb-accounting/repository/sqlite"
	"osbb-accounting/service"
	"strings"
)

// Скрипт для скидання пароля користувача
func main() {
	fmt.Println("=== Скидання Пароля Користувача ===\n")

	// Ініціалізуємо БД
	dbConfig := sqlite.DefaultDBConfig()
	db, err := sqlite.InitDatabase(dbConfig)
	if err != nil {
		log.Fatalf("❌ Помилка БД: %v", err)
	}
	defer db.Close()

	fmt.Printf("✓ БД підключена: %s\n", dbConfig.DBPath)

	// Створюємо репозиторій та сервіс
	userRepo := sqlite.NewSQLiteUserRepository(db)
	authService := service.NewAuthService(userRepo)

	// Отримуємо список користувачів
	users, err := userRepo.FindAll()
	if err != nil {
		log.Fatalf("❌ Помилка отримання користувачів: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("⚠️ Користувачів не знайдено в БД!")
		return
	}

	// Показуємо список
	fmt.Println("\nСписок користувачів:")
	for i, user := range users {
		fmt.Printf("%d. %s (ID: %d, Роль: %s)\n", i+1, user.Username, user.ID, user.Role)
	}

	// Запитуємо, кому скинути пароль
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\nВведіть ім'я користувача для скидання пароля: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Знаходимо користувача
	user, err := userRepo.FindByUsername(username)
	if err != nil {
		log.Fatalf("❌ Користувача '%s' не знайдено: %v", username, err)
	}

	fmt.Printf("\n✓ Знайдено: %s (%s)\n", user.FullName, user.Role)

	// Запитуємо новий пароль
	fmt.Print("\nВведіть новий пароль (мінімум 6 символів): ")
	newPassword, _ := reader.ReadString('\n')
	newPassword = strings.TrimSpace(newPassword)

	if len(newPassword) < 6 {
		log.Fatal("❌ Пароль має бути не менше 6 символів!")
	}

	// Підтвердження
	fmt.Print("\nПідтвердіть новий пароль: ")
	confirmPassword, _ := reader.ReadString('\n')
	confirmPassword = strings.TrimSpace(confirmPassword)

	if newPassword != confirmPassword {
		log.Fatal("❌ Паролі не співпадають!")
	}

	// Хешуємо новий пароль
	hashedPassword, err := authService.HashPassword(newPassword)
	if err != nil {
		log.Fatalf("❌ Помилка хешування: %v", err)
	}

	// Оновлюємо пароль
	err = userRepo.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		log.Fatalf("❌ Помилка оновлення: %v", err)
	}

	fmt.Printf("\n✓ Пароль успішно змінено для користувача '%s'\n", username)
	fmt.Println("\nТепер ви можете увійти з новим паролем.")

	// Тестуємо новий пароль
	fmt.Println("\n--- Тестування нового пароля ---")
	testUser, err := authService.Authenticate(username, newPassword)
	if err != nil {
		fmt.Printf("❌ Помилка тесту: %v\n", err)
	} else {
		fmt.Printf("✓ Аутентифікація успішна! (User: %s)\n", testUser.FullName)
	}
}
