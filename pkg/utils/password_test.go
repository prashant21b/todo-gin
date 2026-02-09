package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "normal password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-that-should-still-work-fine",
			wantErr:  false,
		},
		{
			name:     "special characters",
			password: "P@$$w0rd!#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Hash should not be empty
				if hash == "" {
					t.Error("HashPassword() returned empty hash")
				}

				// Hash should not equal the original password
				if hash == tt.password {
					t.Error("HashPassword() hash equals original password")
				}
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	// Create some hashed passwords for testing
	hash1, _ := HashPassword("password123")
	hash2, _ := HashPassword("different-password")

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: "password123",
			hash:     hash1,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     hash1,
			want:     false,
		},
		{
			name:     "password for different hash",
			password: "password123",
			hash:     hash2,
			want:     false,
		},
		{
			name:     "empty password with non-empty hash",
			password: "",
			hash:     hash1,
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: "password123",
			hash:     "not-a-valid-hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, tt.hash)

			if result != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	password := "same-password"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// bcrypt should generate different hashes for the same password (due to salt)
	if hash1 == hash2 {
		t.Error("HashPassword() should generate unique hashes for the same password")
	}

	// But both should still validate correctly
	if !CheckPassword(password, hash1) {
		t.Error("CheckPassword() should validate hash1")
	}
	if !CheckPassword(password, hash2) {
		t.Error("CheckPassword() should validate hash2")
	}
}
