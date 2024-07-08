package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/dawumnam/token-trader/config"
	"github.com/golang-jwt/jwt/v5"
)

func TestCreateJWT(t *testing.T) {
	config.Envs.JWTExpirationInSeconds = 3600
	secret := []byte("test_secret")
	userID := 123

	token, err := CreateJWT(secret, userID)
	if err != nil {
		t.Fatalf("CreateJWT() error = %v", err)
	}
	if token == "" {
		t.Errorf("CreateJWT() returned empty token")
	}
}

func TestGetTokenFromRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "test_token")

	token := GetTokenFromRequest(req)
	if token != "test_token" {
		t.Errorf("GetTokenFromRequest() = %v, want %v", token, "test_token")
	}
}

func TestValidateToken(t *testing.T) {
	config.Envs.JWTSecret = "test_secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    "123",
		"expiredAt": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("test_secret"))

	validatedToken, err := ValidateToken(tokenString)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}
	if !validatedToken.Valid {
		t.Errorf("ValidateToken() returned invalid token")
	}
}
