package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	result, err := HashPassword("Something")
	if err != nil {
		t.Errorf("failed")
	}
	fmt.Println(result)
}

func TestCheckPassword(t *testing.T) {
	hashedPassword, _ := HashPassword("Something else")

	result, err := CheckPassword("Szevasztok3", hashedPassword)
	if err != nil {
		t.Errorf("failed")
	}
	fmt.Println(result)
}

func TestMakeJWT(t *testing.T) {
	duration := 2 * time.Second
	fmt.Println(MakeJWT(uuid.New(), "mySecret", time.Duration(duration)))
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		want        string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer abc123",
			want:        "abc123",
			expectError: false,
		},
		{
			name:        "missing authorization header",
			authHeader:  "",
			want:        "",
			expectError: true,
		},
		{
			name:        "longer valid token",
			authHeader:  "Bearer this-is-my-secret-token",
			want:        "this-is-my-secret-token",
			expectError: false,
		},
		{
			name:        "authorization header without Bearer prefix",
			authHeader:  "abc123",
			want:        "",
			expectError: true,
		},
		{
			name:        "wrong authorization scheme",
			authHeader:  "Basic abc123",
			want:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.authHeader != "" {
				headers.Set("Authorization", tt.authHeader)
			}

			got, err := GetBearerToken(headers)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
