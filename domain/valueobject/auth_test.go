package valueobject_test

import (
	"testing"

	"osbb-accounting/domain/valueobject"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentials(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		creds, err := valueobject.NewCredentials("user", "pass")
		require.NoError(t, err)
		assert.Equal(t, "user", creds.UsernameOrEmail)
	})

	t.Run("Empty Username", func(t *testing.T) {
		_, err := valueobject.NewCredentials("", "pass")
		assert.ErrorIs(t, err, valueobject.ErrCredentialsUsernameEmpty)
	})

	t.Run("Empty Password", func(t *testing.T) {
		_, err := valueobject.NewCredentials("user", "")
		assert.ErrorIs(t, err, valueobject.ErrCredentialsPasswordEmpty)
	})
}

func TestNewPasswordChangeRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req, err := valueobject.NewPasswordChangeRequest(1, "old", "new")
		require.NoError(t, err)
		assert.Equal(t, "old", req.OldPassword)
	})

	t.Run("Same Password", func(t *testing.T) {
		_, err := valueobject.NewPasswordChangeRequest(1, "pass", "pass")
		assert.ErrorIs(t, err, valueobject.ErrPasswordChangeSame)
	})
}

func TestNewRegistrationRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req, err := valueobject.NewRegistrationRequest(
			"user", "email@example.com", "pass", "First", "Last", nil, nil,
		)
		require.NoError(t, err)
		assert.Equal(t, "user", req.Username)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		_, err := valueobject.NewRegistrationRequest(
			"", "email", "pass", "First", "Last", nil, nil,
		)
		assert.Error(t, err)
	})
}
