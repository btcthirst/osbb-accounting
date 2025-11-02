package main

import (
	"fmt"
	"log"
	"osbb-accounting/repository/sqlite"
)

// Тестовий скрипт для перевірки ініціалізації репозиторіїв
func main() {
	fmt.Println("=== Тест Ініціалізації Репозиторіїв ===\n")

	// Ініціалізуємо БД
	dbConfig := sqlite.DefaultDBConfig()
	fmt.Printf("Шлях до БД: %s\n", dbConfig.DBPath)

	db, err := sqlite.InitDatabase(dbConfig)
	if err != nil {
		log.Fatalf("❌ Помилка БД: %v", err)
	}
	defer db.Close()

	fmt.Println("✓ База даних підключена\n")

	// Тестуємо UserRepository
	fmt.Println("--- Тест UserRepository ---")
	userRepo := sqlite.NewSQLiteUserRepository(db)
	if userRepo == nil {
		log.Fatal("❌ UserRepository повернув nil!")
	}
	fmt.Printf("✓ UserRepository створено: %T\n", userRepo)

	users, err := userRepo.FindAll()
	if err != nil {
		log.Fatalf("❌ Помилка FindAll: %v", err)
	}
	fmt.Printf("✓ Знайдено користувачів: %d\n\n", len(users))

	// Тестуємо ApartmentRepository
	fmt.Println("--- Тест ApartmentRepository ---")
	apartmentRepo := sqlite.NewSQLiteApartmentRepository(db)
	if apartmentRepo == nil {
		log.Fatal("❌ ApartmentRepository повернув nil!")
	}
	fmt.Printf("✓ ApartmentRepository створено: %T\n", apartmentRepo)

	// Перевіряємо наявність таблиці
	count, err := apartmentRepo.Count()
	if err != nil {
		log.Fatalf("❌ Помилка Count: %v\nМожлива причина: таблиця apartments не існує!", err)
	}
	fmt.Printf("✓ Кількість квартир: %d\n", count)

	// Спробуємо завантажити всі квартири
	apartments, err := apartmentRepo.FindAll()
	if err != nil {
		log.Fatalf("❌ Помилка FindAll: %v", err)
	}
	fmt.Printf("✓ Завантажено квартир: %d\n\n", len(apartments))

	// Показуємо перші 3 квартири
	if len(apartments) > 0 {
		fmt.Println("Перші квартири:")
		for i, apt := range apartments {
			if i >= 3 {
				break
			}
			fmt.Printf("  %d. Кв. %s - %s\n", i+1, apt.ApartmentNumber, apt.OwnerName)
		}
	} else {
		fmt.Println("⚠️ Квартир не знайдено. Можливо, не застосована міграція 002.")
	}

	fmt.Println("\n=== Тест завершено успішно! ===")
}
