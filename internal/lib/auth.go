package lib

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const TokenExp = time.Hour * 1
const SecretKey = "supersecretkey"
const CookieName = "auth"

type contextKey string

const UserContextKey contextKey = "userID"

func GenerateAuthCookie(userID string) (*http.Cookie, error) {
	token, err := generateJWTString(userID)
	if err != nil {
		return nil, fmt.Errorf("generateCookie: failed to generate, %s", err)
	}

	return &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	}, nil
}

func generateJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})
	return token.SignedString([]byte(SecretKey))
}
