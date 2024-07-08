package auth

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dawumnam/token-trader/config"
	"github.com/dawumnam/token-trader/db"
	"github.com/dawumnam/token-trader/types"
	"github.com/dawumnam/token-trader/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/net/context"
)

func CreateJWT(secret []byte, userID int) (string, error) {
	expiration := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(userID),
		"expiredAt": time.Now().Add(expiration).Unix(),
	})
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetTokenFromRequest(r *http.Request) string {
	authToken := r.Header.Get("Authorization")

	if authToken != "" {
		return authToken
	}

	return ""
}

func ValidateToken(t string) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid singing method: %v", t.Header["alg"])
		}

		return []byte(config.Envs.JWTSecret), nil
	})
}

func WithJWTAuth(handlerFunc http.HandlerFunc, repo types.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := GetTokenFromRequest(r)

		t, err := ValidateToken(token)

		if err != nil || !t.Valid {
			log.Printf("validation failed %v", err)
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
			return
		}

		blacklisted, err := db.IsTokenBlacklisted(token)
		if err != nil {
			log.Printf("failed to check token blacklist: %v", err)
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			return
		}
		if blacklisted {
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("token has been invalidated"))
			return
		}

		claims := t.Claims.(jwt.MapClaims)

		str := claims["userID"].(string)

		userID, _ := strconv.Atoi(str)

		u, err := repo.GetUserById(userID)

		if err != nil {
			log.Printf("failed to get user by id during jwt validation: %v", err)
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", u.ID)
		r = r.WithContext(ctx)

		handlerFunc(w, r)
	}
}

func GetUserIdFromContext(ctx context.Context) int {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return -1
	}
	return userID
}
