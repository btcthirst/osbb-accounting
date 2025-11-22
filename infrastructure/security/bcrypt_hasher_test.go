package security_test

import (
	"testing"

	"osbb-accounting/infrastructure/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewBCryptHasher(t *testing.T) {
	t.Run("Default Cost", func(t *testing.T) {
		hasher := security.NewBCryptHasher(0)
		assert.NotNil(t, hasher)
	})

	t.Run("Valid Cost", func(t *testing.T) {
		hasher := security.NewBCryptHasher(12)
		assert.NotNil(t, hasher)
	})

	t.Run("Invalid Cost Low", func(t *testing.T) {
		hasher := security.NewBCryptHasher(1)
		assert.NotNil(t, hasher)
	})

	t.Run("Invalid Cost High", func(t *testing.T) {
		hasher := security.NewBCryptHasher(50)
		assert.NotNil(t, hasher)
	})
}

func TestBCryptHasher_Hash(t *testing.T) {
	hasher := security.NewBCryptHasher(bcrypt.MinCost) // Use min cost for speed

	t.Run("Success", func(t *testing.T) {
		hash, err := hasher.Hash("password123")
		require.NoError(t, err)
		assert.Len(t, hash, 60)
	})

	t.Run("Empty Password", func(t *testing.T) {
		_, err := hasher.Hash("")
		assert.Error(t, err)
	})
}

func TestBCryptHasher_Verify(t *testing.T) {
	hasher := security.NewBCryptHasher(bcrypt.MinCost)
	password := "secret"
	hash, _ := hasher.Hash(password)

	t.Run("Success", func(t *testing.T) {
		err := hasher.Verify(password, hash)
		assert.NoError(t, err)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		err := hasher.Verify("wrong", hash)
		assert.Error(t, err)
		assert.Equal(t, "invalid password", err.Error())
	})

	t.Run("Invalid Hash Format", func(t *testing.T) {
		err := hasher.Verify(password, "invalid-hash")
		assert.Error(t, err)
	})

	t.Run("Empty Inputs", func(t *testing.T) {
		assert.Error(t, hasher.Verify("", hash))
		assert.Error(t, hasher.Verify(password, ""))
	})
}
