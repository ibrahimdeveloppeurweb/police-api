package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordService_HashPassword(t *testing.T) {
	logger := zap.NewNop()
	service := NewPasswordService(logger)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mySecurePassword123",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "123",
			wantErr:  false, // bcrypt doesn't enforce length, that's validation layer
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-with-many-characters-to-test-bcrypt-limits",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "p@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: "motDePasse123éàü",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash) // Hash should be different from password
				assert.True(t, len(hash) > 50)        // bcrypt hashes are long

				// Verify it's a valid bcrypt hash
				cost, err := bcrypt.Cost([]byte(hash))
				assert.NoError(t, err)
				assert.Equal(t, bcrypt.DefaultCost, cost)
			}
		})
	}
}

func TestPasswordService_CheckPassword(t *testing.T) {
	logger := zap.NewNop()
	service := NewPasswordService(logger)

	// Generate a known hash for testing
	password := "testPassword123"
	hash, err := service.HashPassword(password)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			wantErr:  true,
		},
		{
			name:     "invalid hash format",
			password: password,
			hash:     "invalid-hash-format",
			wantErr:  true,
		},
		{
			name:     "case sensitive password",
			password: "TESTPASSWORD123",
			hash:     hash,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CheckPassword(tt.password, tt.hash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordService_HashAndCheck_Integration(t *testing.T) {
	logger := zap.NewNop()
	service := NewPasswordService(logger)

	passwords := []string{
		"simplePassword",
		"ComplexP@ssw0rd!",
		"123456789",
		"motDePasse123éàü",
		"very-long-password-with-multiple-words-and-special-characters-!@#$%",
	}

	for _, password := range passwords {
		t.Run("hash_and_check_"+password, func(t *testing.T) {
			// Hash the password
			hash, err := service.HashPassword(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Verify correct password
			err = service.CheckPassword(password, hash)
			assert.NoError(t, err)

			// Verify incorrect password fails
			err = service.CheckPassword(password+"wrong", hash)
			assert.Error(t, err)

			// Verify each hash is unique (even for same password)
			hash2, err := service.HashPassword(password)
			assert.NoError(t, err)
			assert.NotEqual(t, hash, hash2) // Salt makes each hash unique
		})
	}
}

func TestPasswordService_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	service := NewPasswordService(logger)

	password := "concurrentTestPassword"
	
	// Test concurrent hashing
	const numGoroutines = 5
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			hash, err := service.HashPassword(password)
			if err != nil {
				errors <- err
				return
			}
			results <- hash
		}()
	}

	// Collect results
	var hashes []string
	for i := 0; i < numGoroutines; i++ {
		select {
		case hash := <-results:
			hashes = append(hashes, hash)
		case err := <-errors:
			t.Errorf("Concurrent hashing failed: %v", err)
		}
	}

	// Verify all hashes are valid and unique
	assert.Len(t, hashes, numGoroutines)
	for i, hash := range hashes {
		assert.NotEmpty(t, hash)
		
		// Check password verification works
		err := service.CheckPassword(password, hash)
		assert.NoError(t, err, "Hash %d should verify correctly", i)
		
		// Ensure each hash is unique (due to salt)
		for j, otherHash := range hashes {
			if i != j {
				assert.NotEqual(t, hash, otherHash, "Hashes %d and %d should be different", i, j)
			}
		}
	}
}