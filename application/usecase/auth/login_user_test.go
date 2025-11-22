package auth_test

import (
	"context"
	"errors"
	"testing"

	"osbb-accounting/application/usecase/auth"
	"osbb-accounting/application/usecase/mocks"
	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoginUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Setup mocks
		userRepo := new(mocks.UserRepositoryMock)
		sessionRepo := new(mocks.SessionRepositoryMock)
		roleRepo := new(mocks.RoleRepositoryMock)
		permissionRepo := new(mocks.PermissionRepositoryMock)
		hasher := new(mocks.PasswordHasherMock)

		useCase := auth.NewLoginUserUseCase(userRepo, sessionRepo, roleRepo, permissionRepo, hasher)

		// Test data
		input := auth.LoginUserInput{
			UsernameOrEmail: "testuser",
			Password:        "password123",
		}

		user := &entity.User{
			ID:           1,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			IsActive:     true,
		}

		// Expectations
		userRepo.On("GetByUsernameOrEmail", ctx, "testuser").Return(user, nil)
		hasher.On("Verify", "password123", "hashed_password").Return(nil)
		sessionRepo.On("Create", ctx, mock.AnythingOfType("*entity.Session")).Return(nil)
		userRepo.On("UpdateLastLogin", ctx, int64(1)).Return(nil)
		roleRepo.On("GetByUserID", ctx, int64(1)).Return([]*entity.Role{{Name: "admin"}}, nil)
		permissionRepo.On("GetByUserID", ctx, int64(1)).Return([]*entity.Permission{{Code: "read_users"}}, nil)

		// Execute
		output, err := useCase.Execute(ctx, input)

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, user.ID, output.User.ID)
		assert.NotEmpty(t, output.SessionToken)
		assert.Equal(t, []string{"admin"}, output.Roles)
		assert.Equal(t, []string{"read_users"}, output.Permissions)

		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
		hasher.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userRepo := new(mocks.UserRepositoryMock)
		useCase := auth.NewLoginUserUseCase(userRepo, nil, nil, nil, nil)

		userRepo.On("GetByUsernameOrEmail", ctx, "unknown").Return(nil, domainErrors.ErrNotFound)

		input := auth.LoginUserInput{UsernameOrEmail: "unknown", Password: "password"}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrInvalidCredentials))
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		userRepo := new(mocks.UserRepositoryMock)
		hasher := new(mocks.PasswordHasherMock)
		useCase := auth.NewLoginUserUseCase(userRepo, nil, nil, nil, hasher)

		user := &entity.User{ID: 1, PasswordHash: "hashed", IsActive: true}

		userRepo.On("GetByUsernameOrEmail", ctx, "user").Return(user, nil)
		hasher.On("Verify", "wrong", "hashed").Return(errors.New("mismatch"))

		input := auth.LoginUserInput{UsernameOrEmail: "user", Password: "wrong"}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.True(t, domainErrors.Is(err, domainErrors.ErrInvalidCredentials))
	})

	t.Run("UserInactive", func(t *testing.T) {
		userRepo := new(mocks.UserRepositoryMock)
		useCase := auth.NewLoginUserUseCase(userRepo, nil, nil, nil, nil)

		user := &entity.User{ID: 1, IsActive: false}

		userRepo.On("GetByUsernameOrEmail", ctx, "inactive").Return(user, nil)

		input := auth.LoginUserInput{UsernameOrEmail: "inactive", Password: "password"}
		output, err := useCase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "inactive")
	})
}
