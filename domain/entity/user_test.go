package entity_test

import (
	"strings"
	"testing"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		user, err := entity.NewUser("jdoe", "jdoe@example.com", "John", "Doe", nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "jdoe", user.Username)
		assert.True(t, user.IsActive)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name     string
			username string
			email    string
			wantErr  error
		}{
			{"Empty Username", "", "test@example.com", entity.ErrUsernameRequired},
			{"Short Username", "ab", "test@example.com", entity.ErrUsernameTooShort},
			{"Invalid Email", "user", "invalid-email", entity.ErrEmailInvalidFormat},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := entity.NewUser(tt.username, tt.email, "First", "Last", nil, nil)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"Valid", "Password123", nil},
		{"Too Short", "Pass1", entity.ErrPasswordTooShort},
		{"No Upper", "password123", entity.ErrPasswordTooWeak},
		{"No Lower", "PASSWORD123", entity.ErrPasswordTooWeak},
		{"No Number", "Password", entity.ErrPasswordTooWeak},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entity.ValidatePassword(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_SetPasswordHash(t *testing.T) {
	user, _ := entity.NewUser("test", "test@example.com", "Test", "User", nil, nil)

	t.Run("Valid Hash", func(t *testing.T) {
		validHash := strings.Repeat("a", 60)
		err := user.SetPasswordHash(validHash)
		require.NoError(t, err)
		assert.Equal(t, validHash, user.PasswordHash)
	})

	t.Run("Invalid Hash Length", func(t *testing.T) {
		err := user.SetPasswordHash("short")
		assert.Error(t, err)
	})
}

func TestUser_ChangeEmail(t *testing.T) {
	user, _ := entity.NewUser("test", "test@example.com", "Test", "User", nil, nil)

	t.Run("Success", func(t *testing.T) {
		newEmail := "new@example.com"
		err := user.ChangeEmail(newEmail)
		require.NoError(t, err)
		assert.Equal(t, newEmail, user.Email)
	})

	t.Run("Invalid Format", func(t *testing.T) {
		err := user.ChangeEmail("invalid")
		assert.ErrorIs(t, err, entity.ErrEmailInvalidFormat)
	})
}
