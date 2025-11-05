package main

import (
	"database/sql"
	"fmt"
	"log"
	"osbb-accounting/domain"
	"osbb-accounting/repository/sqlite"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestRepositories - тестовий код для перевірки роботи нових репозиторіїв.
// Цей код можна запустити окремо для тестування.
func TestRepositories() {
	// Відкриваємо тестову базу даних
	db, err := sql.Open("sqlite3", "./test_osbb.db")
	if err != nil {
		log.Fatalf("Помилка відкриття БД: %v", err)
	}
	defer db.Close()

	// Створюємо фабрику репозиторіїв
	factory := sqlite.NewRepositoryFactory(db)
	repos := factory.CreateAllRepositories()

	fmt.Println("=== Тестування Репозиторіїв ===\n")

	// 1. Тестуємо OSBB Repository
	fmt.Println("1. Тестування OSBB Repository...")
	testOSBBRepository(repos.OSBB)

	// 2. Тестуємо Owner Repository
	fmt.Println("\n2. Тестування Owner Repository...")
	ownerID := testOwnerRepository(repos.Owner)

	// 3. Тестуємо PersonalAccount Repository
	fmt.Println("\n3. Тестування PersonalAccount Repository...")
	testPersonalAccountRepository(repos.PersonalAccount, ownerID)

	fmt.Println("\n=== Всі тести завершено! ===")
}

// testOSBBRepository тестує OSBB Repository.
func testOSBBRepository(repo *sqlite.OSBBRepositoryImpl) {
	// Перевіряємо існування ОСББ
	exists, err := repo.Exists()
	if err != nil {
		log.Printf("Помилка перевірки існування: %v", err)
	}
	fmt.Printf("✓ ОСББ існує: %v\n", exists)

	if !exists {
		// Створюємо нове ОСББ
		osbb := &domain.OSBB{
			Name:          "ОСББ Тестовий Будинок",
			ShortName:     "ОСББ ТБ",
			Address:       "м. Київ, вул. Тестова, 1",
			EDRPOU:        "12345678",
			BaseRate:      15.50,
			ChairmanName:  "Тестовий Голова",
			ChairmanPhone: "+380501234567",
			ChairmanEmail: "chairman@test.ua",
			BankName:      "ТестБанк",
			BankAccount:   "UA123456789012345678901234567",
			MFO:           "305299",
			FoundedAt:     time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
		}

		err = repo.Save(osbb)
		if err != nil {
			log.Printf("Помилка збереження ОСББ: %v", err)
		} else {
			fmt.Printf("✓ ОСББ створено з ID: %d\n", osbb.ID)
		}
	}

	// Отримуємо дані ОСББ
	osbb, err := repo.Get()
	if err != nil {
		log.Printf("Помилка отримання ОСББ: %v", err)
	} else {
		fmt.Printf("✓ ОСББ отримано: %s (Тариф: %.2f грн/м²)\n", osbb.GetDisplayName(), osbb.BaseRate)
	}

	// Оновлюємо тариф
	if osbb != nil {
		osbb.BaseRate = 16.00
		err = repo.Update(osbb)
		if err != nil {
			log.Printf("Помилка оновлення ОСББ: %v", err)
		} else {
			fmt.Printf("✓ Тариф оновлено: %.2f грн/м²\n", osbb.BaseRate)
		}
	}

	// Отримуємо налаштування
	settings, err := repo.GetSettings()
	if err != nil {
		log.Printf("Помилка отримання налаштувань: %v", err)
	} else {
		fmt.Printf("✓ Статистика ОСББ:\n")
		fmt.Printf("  - Квартир: %d\n", settings.TotalApartments)
		fmt.Printf("  - Площа: %.2f м²\n", settings.TotalArea)
		fmt.Printf("  - Власників: %d\n", settings.TotalOwners)
		fmt.Printf("  - Активних рахунків: %d\n", settings.ActiveAccounts)
	}
}

// testOwnerRepository тестує Owner Repository.
func testOwnerRepository(repo *sqlite.OwnerRepositoryImpl) int {
	// Створюємо нового власника
	owner := &domain.Owner{
		FullName:         "Тестовий Власник Іванович",
		TaxID:            "1234567890",
		Phone:            "+380501234567",
		Email:            "test@owner.ua",
		AlternativePhone: "+380672345678",
		PassportSeries:   "АА",
		PassportNumber:   "123456",
		Notes:            "Тестовий власник для перевірки системи",
	}

	err := repo.Save(owner)
	if err != nil {
		log.Printf("Помилка збереження власника: %v", err)
		return 0
	}
	fmt.Printf("✓ Власника створено з ID: %d\n", owner.ID)

	// Перевіряємо пошук за ID
	foundOwner, err := repo.FindByID(owner.ID)
	if err != nil {
		log.Printf("Помилка пошуку за ID: %v", err)
	} else {
		fmt.Printf("✓ Власник знайдений: %s (ІПН: %s)\n", foundOwner.FullName, foundOwner.TaxID)
	}

	// Перевіряємо пошук за ІПН
	foundByTax, err := repo.FindByTaxID(owner.TaxID)
	if err != nil {
		log.Printf("Помилка пошуку за ІПН: %v", err)
	} else {
		fmt.Printf("✓ Пошук за ІПН успішний: %s\n", foundByTax.FullName)
	}

	// Оновлюємо дані
	owner.Phone = "+380991234567"
	err = repo.Update(owner)
	if err != nil {
		log.Printf("Помилка оновлення власника: %v", err)
	} else {
		fmt.Printf("✓ Телефон власника оновлено: %s\n", owner.Phone)
	}

	// Отримуємо список всіх власників
	allOwners, err := repo.FindAll()
	if err != nil {
		log.Printf("Помилка отримання списку: %v", err)
	} else {
		fmt.Printf("✓ Всього власників: %d\n", len(allOwners))
	}

	// Пошук
	searchResults, err := repo.Search("Тестовий")
	if err != nil {
		log.Printf("Помилка пошуку: %v", err)
	} else {
		fmt.Printf("✓ Знайдено за пошуком: %d\n", len(searchResults))
	}

	// Підрахунок
	count, err := repo.Count()
	if err != nil {
		log.Printf("Помилка підрахунку: %v", err)
	} else {
		fmt.Printf("✓ Кількість активних власників: %d\n", count)
	}

	return owner.ID
}

// testPersonalAccountRepository тестує PersonalAccount Repository.
func testPersonalAccountRepository(repo *sqlite.PersonalAccountRepositoryImpl, ownerID int) {
	// Генеруємо номер рахунку
	accountNumber, err := repo.GenerateAccountNumber()
	if err != nil {
		log.Printf("Помилка генерації номера: %v", err)
		return
	}
	fmt.Printf("✓ Згенеровано номер рахунку: %s\n", accountNumber)

	// Створюємо особистий рахунок (потрібен існуючий apartment_id)
	// Припускаємо, що квартира з ID=1 існує
	account := &domain.PersonalAccount{
		AccountNumber: accountNumber,
		ApartmentID:   1, // Має існувати!
		OwnerID:       ownerID,
		Notes:         "Тестовий особистий рахунок",
	}

	err = repo.Save(account)
	if err != nil {
		log.Printf("Помилка збереження рахунку: %v", err)
		return
	}
	fmt.Printf("✓ Особистий рахунок створено з ID: %d\n", account.ID)

	// Пошук за номером
	foundAccount, err := repo.FindByAccountNumber(accountNumber)
	if err != nil {
		log.Printf("Помилка пошуку рахунку: %v", err)
	} else {
		fmt.Printf("✓ Рахунок знайдений: %s (Баланс: %.2f грн)\n",
			foundAccount.AccountNumber, foundAccount.CurrentBalance)
	}

	// Оновлюємо баланс (симулюємо борг)
	err = repo.UpdateBalance(account.ID, -500.00)
	if err != nil {
		log.Printf("Помилка оновлення балансу: %v", err)
	} else {
		fmt.Printf("✓ Баланс оновлено: -500.00 грн (борг)\n")
	}

	// Перевіряємо рахунки з боргом
	debtAccounts, err := repo.FindWithDebt()
	if err != nil {
		log.Printf("Помилка пошуку боржників: %v", err)
	} else {
		fmt.Printf("✓ Рахунків з боргом: %d\n", len(debtAccounts))
		for _, acc := range debtAccounts {
			fmt.Printf("  - Рахунок %s: борг %.2f грн\n",
				acc.AccountNumber, acc.GetDebtAmount())
		}
	}

	// Загальний борг
	totalDebt, err := repo.GetTotalDebt()
	if err != nil {
		log.Printf("Помилка підрахунку боргу: %v", err)
	} else {
		fmt.Printf("✓ Загальний борг: %.2f грн\n", totalDebt)
	}

	// Спроба закриття рахунку з боргом (має видати помилку)
	err = repo.Close(account.ID)
	if err == domain.ErrCannotCloseAccountWithDebt {
		fmt.Printf("✓ Правильно: неможливо закрити рахунок з боргом\n")
	} else if err != nil {
		log.Printf("Неочікувана помилка закриття: %v", err)
	}

	// Погашаємо борг
	err = repo.UpdateBalance(account.ID, 0)
	if err != nil {
		log.Printf("Помилка погашення боргу: %v", err)
	} else {
		fmt.Printf("✓ Борг погашено\n")
	}

	// Тепер закриваємо рахунок
	err = repo.Close(account.ID)
	if err != nil {
		log.Printf("Помилка закриття рахунку: %v", err)
	} else {
		fmt.Printf("✓ Рахунок успішно закрито\n")
	}

	// Знову відкриваємо
	err = repo.Reopen(account.ID)
	if err != nil {
		log.Printf("Помилка відкриття рахунку: %v", err)
	} else {
		fmt.Printf("✓ Рахунок знову відкрито\n")
	}

	// Підрахунок рахунків
	count, err := repo.Count()
	if err != nil {
		log.Printf("Помилка підрахунку: %v", err)
	} else {
		fmt.Printf("✓ Кількість активних рахунків: %d\n", count)
	}
}

// Розкоментуйте функцію main для запуску тестів:
/*
func main() {
	TestRepositories()
}
*/
