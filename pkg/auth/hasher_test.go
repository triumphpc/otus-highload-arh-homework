package auth

import (
	"errors"
	"testing"
)

func TestBcryptHasher(t *testing.T) {
	hasher := NewBcryptHasher(10)

	t.Run("Hash and Check", func(t *testing.T) {
		password := "test-password-123"
		hash, err := hasher.Hash(password)
		if err != nil {
			t.Fatalf("Hash failed: %v", err)
		}

		if !hasher.Check(password, hash) {
			t.Error("Check failed for correct password")
		}

		if hasher.Check("wrong-password", hash) {
			t.Error("Check passed for wrong password")
		}
	})

	t.Run("Long password", func(t *testing.T) {
		longPass := string(make([]byte, 100))
		_, err := hasher.Hash(longPass)
		if !errors.Is(err, ErrPasswordTooLong) {
			t.Errorf("Expected ErrPasswordTooLong, got %v", err)
		}
	})

	t.Run("IsHashed", func(t *testing.T) {
		tests := []struct {
			input string
			want  bool
		}{
			{"$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqK3X6R/ztUpB7WQyJ2sFjJNDjY95a", true},
			{"plaintext", false},
			{"", false},
		}

		for _, tt := range tests {
			if got := hasher.IsHashed(tt.input); got != tt.want {
				t.Errorf("IsHashed(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})
}
