package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"osbb-accounting/infrastructure/persistence/sqlite"
	"osbb-accounting/infrastructure/security"
)

func main() {
	// 1. Connect to database
	dbConfig := sqlite.DefaultConfig()
	dbConfig.Path = "./osbb.db" // Assume running from root

	fmt.Printf("Connecting to database at %s...\n", dbConfig.Path)
	db, err := sqlite.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlite.Close(db)

	// 2. Initialize components
	userRepo := sqlite.NewUserRepository(db)
	passwordHasher := security.DefaultBCryptHasher()

	// 3. Get arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run debug/reset_password.go <username> <new_password>")
		return
	}
	username := os.Args[1]
	newPassword := os.Args[2]

	ctx := context.Background()

	// 4. Find user
	fmt.Printf("Looking for user '%s'...\n", username)
	user, err := userRepo.GetByUsername(ctx, username)
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}

	// 5. Hash new password
	fmt.Println("Hashing new password...")
	hash, err := passwordHasher.Hash(newPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 6. Update password
	fmt.Println("Updating password...")
	if err := userRepo.UpdatePasswordHash(ctx, user.ID, hash); err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("PASSWORD RESET SUCCESSFUL")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("User: %s (ID: %d)\n", user.Username, user.ID)
	fmt.Println("New password has been set.")
	fmt.Println("--------------------------------------------------")
}
