package utils

import (
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

func TestMoonAlgorithm(t *testing.T) {
	tests := []struct {
		order string
		want  bool
	}{
		{
			order: "79927398713",
			want:  true,
		},
		{
			order: "49927398716",
			want:  true,
		},

		{
			order: "4111111111111112",
			want:  false,
		},
		{
			order: "4111111111211111",
			want:  false,
		},

		{
			order: "",
			want:  false,
		},
		{
			order: "411111111111111A",
			want:  false,
		},
		{
			order: "4111-1111-1111-1111",
			want:  false,
		},
		{
			order: "4111 1111 1111 1111",
			want:  false,
		},
		{
			order: "123",
			want:  false,
		},
		{
			order: "1",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.order, func(t *testing.T) {
			got := MoonAlgorithm(tt.order)
			if got != tt.want {
				t.Errorf("MoonAlgorithm(%q) = %v, want %v", tt.order, got, tt.want)
			}
		})
	}
}

func TestHasherHashPassword(t *testing.T) {
	hasher := NewHasher(8, 16)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "normal password",
			password: "password",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "very long password",
			password: strings.Repeat("a", 1000),
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "!@#$%^&*",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
			}
		})
	}
}

func TestHasherVerifyPassword(t *testing.T) {
	hasher := NewHasher(8, 16)

	password1 := "password1"
	password2 := "password2"

	hash1, err := hasher.HashPassword(password1)
	require.NoError(t, err)

	hash2, err := hasher.HashPassword(password2)
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		encodedHash string
		valid       bool
		wantErr     bool
	}{
		{
			name:        "correct password",
			password:    password1,
			encodedHash: hash1,
			valid:       true,
			wantErr:     false,
		},
		{
			name:        "wrong password",
			password:    "wrong_password",
			encodedHash: hash1,
			valid:       false,
			wantErr:     false,
		},
		{
			name:        "password from different hash",
			password:    password1,
			encodedHash: hash2,
			valid:       false,
			wantErr:     false,
		},
		{
			name:        "invalid hash format",
			password:    password1,
			encodedHash: "invalid_hash_format",
			valid:       false,
			wantErr:     true,
		},
		{
			name:        "malformed hash",
			password:    password1,
			encodedHash: "$argon2id$invalid",
			valid:       false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := hasher.VerifyPassword(tt.password, tt.encodedHash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.valid, valid)
		})
	}
}
