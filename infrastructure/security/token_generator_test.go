package security_test

import (
	"testing"

	"osbb-accounting/infrastructure/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoTokenGenerator_Generate(t *testing.T) {
	generator := security.NewCryptoTokenGenerator()

	t.Run("Success", func(t *testing.T) {
		token, err := generator.Generate(32)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		// Base64 encoding increases length by ~1.33
		assert.Greater(t, len(token), 32)
	})

	t.Run("Invalid Length", func(t *testing.T) {
		_, err := generator.Generate(0)
		assert.Error(t, err)

		_, err = generator.Generate(-1)
		assert.Error(t, err)
	})

	t.Run("Uniqueness", func(t *testing.T) {
		t1, _ := generator.Generate(16)
		t2, _ := generator.Generate(16)
		assert.NotEqual(t, t1, t2)
	})
}

func TestCryptoTokenGenerator_GenerateHex(t *testing.T) {
	generator := security.NewCryptoTokenGenerator()

	t.Run("Success", func(t *testing.T) {
		length := 16
		token, err := generator.GenerateHex(length)
		require.NoError(t, err)
		assert.Len(t, token, length*2) // Hex is 2 chars per byte
	})

	t.Run("Invalid Length", func(t *testing.T) {
		_, err := generator.GenerateHex(0)
		assert.Error(t, err)
	})
}
