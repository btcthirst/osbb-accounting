package service_test

import (
	"context"
	"testing"

	"osbb-accounting/application/service"
	"osbb-accounting/application/usecase/auth"
	"osbb-accounting/application/usecase/mocks"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := new(mocks.UserRepositoryMock)
		sessionRepo := new(mocks.SessionRepositoryMock)
		roleRepo := new(mocks.RoleRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		hasher := new(mocks.PasswordHasherMock)

		svc := service.NewAuthService(userRepo, sessionRepo, roleRepo, permRepo, hasher)

		ip := "127.0.0.1"
		userAgent := "TestAgent"
		input := auth.LoginUserInput{
			UsernameOrEmail: "testuser",
			Password:        "password123",
			IPAddress:       &ip,
			UserAgent:       &userAgent,
		}

		user := &entity.User{
			ID:           1,
			Username:     "testuser",
			PasswordHash: "hashed_password",
			IsActive:     true,
		}

		userRepo.On("GetByUsernameOrEmail", ctx, "testuser").Return(user, nil)
		hasher.On("Verify", "password123", "hashed_password").Return(nil)
		sessionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)
		userRepo.On("UpdateLastLogin", ctx, int64(1)).Return(nil)
		roleRepo.On("GetByUserID", ctx, int64(1)).Return([]*entity.Role{}, nil)
		permRepo.On("GetByUserID", ctx, int64(1)).Return([]*entity.Permission{}, nil)

		output, err := svc.Login(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(1), output.User.ID)

		userRepo.AssertExpectations(t)
		hasher.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		userRepo := new(mocks.UserRepositoryMock)
		sessionRepo := new(mocks.SessionRepositoryMock)
		roleRepo := new(mocks.RoleRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		hasher := new(mocks.PasswordHasherMock)

		svc := service.NewAuthService(userRepo, sessionRepo, roleRepo, permRepo, hasher)

		input := auth.LoginUserInput{
			UsernameOrEmail: "testuser",
			Password:        "wrongpassword",
		}

		user := &entity.User{
			ID:           1,
			Username:     "testuser",
			PasswordHash: "hashed_password",
			IsActive:     true,
		}

		userRepo.On("GetByUsernameOrEmail", ctx, "testuser").Return(user, nil)
		hasher.On("Verify", "wrongpassword", "hashed_password").Return(domainErrors.NewDomainError(domainErrors.CodeInvalidCredentials, "invalid password", nil))

		output, err := svc.Login(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrInvalidCredentials))
	})
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userRepo := new(mocks.UserRepositoryMock)
		sessionRepo := new(mocks.SessionRepositoryMock)
		roleRepo := new(mocks.RoleRepositoryMock)
		permRepo := new(mocks.PermissionRepositoryMock)
		hasher := new(mocks.PasswordHasherMock)

		svc := service.NewAuthService(userRepo, sessionRepo, roleRepo, permRepo, hasher)
		input := auth.RegisterUserInput{
			Username:  "newuser",
			Email:     "test@example.com",
			Password:  "Password123",
			FirstName: "Test",
			LastName:  "User",
		}

		userRepo.On("ExistsByUsername", ctx, "newuser").Return(false, nil)
		userRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
		hasher.On("Hash", "Password123").Return("$2a$10$TestHashTestHashTestHashTestHashTestHashTestHashTestH", nil)
		roleRepo.On("GetByName", ctx, "viewer").Return(&entity.Role{ID: 2, Name: "viewer"}, nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Run(func(args mock.Arguments) {
			user := args.Get(1).(*entity.User)
			user.ID = 1
		}).Return(nil)
		roleRepo.On("AssignToUser", ctx, mock.AnythingOfType("*entity.UserRole")).Return(nil)

		output, err := svc.Register(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "newuser", output.Username)
	})
}
