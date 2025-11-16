// cmd/migrate/main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/persistence/sqlite/migrations"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := flag.String("db", "", "Path to database file")
	flag.Parse()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Визначаємо шлях до БД
	path := *dbPath
	if path == "" {
		config := sqlite.DefaultConfig()
		path = config.Path
	}

	// Відкриваємо БД напряму (без автоматичних міграцій)
	dsn := fmt.Sprintf("file:%s?_foreign_keys=ON", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Обробка команд
	switch os.Args[1] {
	case "migrate":
		if err := migrations.RunMigrations(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case "status":
		if err := migrations.ShowMigrationStatus(db); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}

	case "version":
		version, err := migrations.GetDatabaseVersion(db)
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Current database version: %s\n", version)

	case "repair":
		fmt.Println("⚠ This will remove orphaned migration records. Continue? (y/n)")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Cancelled")
			os.Exit(0)
		}
		if err := migrations.RepairMigrations(db); err != nil {
			log.Fatalf("Repair failed: %v", err)
		}

	case "reset":
		fmt.Println("⚠⚠⚠ WARNING: This will DELETE ALL DATA! Continue? (type 'yes' to confirm)")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("Cancelled")
			os.Exit(0)
		}
		if err := migrations.ResetDatabase(db); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
OSBB Migration Tool

Usage:
  migrate [command] [flags]

Commands:
  migrate   - Apply all pending migrations
  status    - Show migration status
  version   - Show current database version
  repair    - Remove orphaned migration records
  reset     - Reset database (DELETE ALL DATA!)

Flags:
  -db string  Path to database file (optional)

Examples:
  go run ./cmd/migrate/main.go migrate
  go run ./cmd/migrate/main.go status
  go run ./cmd/migrate/main.go repair
	`)
}
