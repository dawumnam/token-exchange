package auth

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
			name:     "Valid password",
			password: "securePassword123!",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: "longpasswordlongpasswordlongpasswordlongpasswordlongpassword",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == tt.password {
				t.Errorf("HashPassword() returned unhashed password")
			}
		})
	}
}

func TestComparePassword(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		hashedPassword string
		want           bool
	}{
		{
			name:     "Correct password",
			password: "correctPassword123!",
			want:     true,
		},
		{
			name:     "Incorrect password",
			password: "incorrectPassword456!",
			want:     false,
		},
		{
			name:     "Empty password",
			password: "",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			if tt.name == "Incorrect password" {
				hashedPassword, _ = HashPassword("differentPassword789!")
			}

			got := ComparePassword([]byte(hashedPassword), []byte(tt.password))
			if got != tt.want {
				t.Errorf("ComparePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
