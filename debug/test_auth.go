package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/auth"
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

	// 2. Initialize repositories
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	roleRepo := sqlite.NewRoleRepository(db)
	permissionRepo := sqlite.NewPermissionRepository(db)

	// 3. Initialize service
	passwordHasher := security.DefaultBCryptHasher()
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		roleRepo,
		permissionRepo,
		passwordHasher,
	)

	// 4. Get credentials
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run debug/test_auth.go <username> <password>")
		return
	}
	username := os.Args[1]
	password := os.Args[2]

	// 5. Attempt login
	ctx := context.Background()
	ip := "127.0.0.1"
	userAgent := "DebugScript"

	fmt.Printf("Attempting login for user '%s'...\n", username)

	input := auth.LoginUserInput{
		UsernameOrEmail: username,
		Password:        password,
		IPAddress:       &ip,
		UserAgent:       &userAgent,
		SessionDuration: 24 * time.Hour,
	}

	output, err := authService.Login(ctx, input)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// 6. Print success
	fmt.Println("--------------------------------------------------")
	fmt.Println("LOGIN SUCCESSFUL")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("User ID:      %d\n", output.User.ID)
	fmt.Printf("Username:     %s\n", output.User.Username)
	fmt.Printf("Full Name:    %s\n", output.User.FullName)
	fmt.Printf("Email:        %s\n", output.User.Email)
	fmt.Printf("Roles:        %v\n", output.Roles)
	fmt.Printf("SessionToken: %s\n", output.SessionToken)
	fmt.Printf("Expires At:   %s\n", output.ExpiresAt)
	fmt.Println("--------------------------------------------------")
}
