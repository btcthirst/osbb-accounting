package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	paymentFiles := []string{
		"create_payment.go",
		"get_payment.go",
		"update_payment.go",
		"delete_payment.go",
		"list_payments.go",
		"approve_payment.go",
	}
	ownershipFiles := []string{
		"ownership_output.go",
		"create_ownership_share.go",
		"get_ownership_share.go",
		"update_ownership_share.go",
		"delete_ownership_share.go",
		"deactivate_ownership_share.go",
		"list_ownership_shares.go",
		"get_shares_by_apartment.go",
		"get_shares_by_owner.go",
	}
	/*expenseCategoryFiles := []string{
		"expense_category_output.go",
		"create_expense_category.go",
		"get_expense_category.go",
		"update_expense_category.go",
		"delete_expense_category.go",
		"deactivate_expense_category.go",
		"list_expense_categories.go",
		"get_category_tree.go",
		"move_expense_category.go",
	}
	*/
	createFiles("/home/min/git-workspace/osbb-accounting/application/usecase/payment/", paymentFiles)
	createFiles("/home/min/git-workspace/osbb-accounting/application/usecase/ownership/", ownershipFiles)
	//createFiles("/home/min/git-workspace/osbb-accounting/application/usecase/expense_category/", expenseCategoryFiles)
}

func createFiles(path string, files []string) {
	for _, file := range files {
		filePath := path + file
		err := createFile(filePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
		}
	}
}

func createFile(filePath string) error {
	// Валідація вхідних даних
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Нормалізація шляху (обробка .., ., тощо)
	cleanPath := filepath.Clean(filePath)

	// Перевірка, чи файл вже існує
	if _, err := os.Stat(cleanPath); err == nil {
		return fmt.Errorf("%w: file already exists: %s", os.ErrExist, cleanPath)
	} else if !os.IsNotExist(err) {
		// Інша помилка при перевірці (недостатньо прав, тощо)
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Отримання директорії з шляху
	dir := filepath.Dir(cleanPath)

	// Створення всіх необхідних директорій (0755 - rwxr-xr-x)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Створення файлу (0644 - rw-r--r--)
	file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	// Закриття файлу
	if err := file.Close(); err != nil {
		// Спроба видалити створений файл при помилці закриття
		_ = os.Remove(cleanPath)
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}
