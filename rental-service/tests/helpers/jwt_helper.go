package helpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	SecretKey    = "test-secret-key-for-hmac-256"
	WrongSecret  = "wrong-secret-key-for-hmac-256"
)

func CreateToken(sub string, role string, expired bool) string {
	now := time.Now()
	exp := now.Add(1 * time.Hour)
	if expired {
		exp = now.Add(-1 * time.Hour)
	}

	claims := jwt.MapClaims{
		"sub":  sub,
		"role": role,
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		panic(err)
	}
	return signed
}

func CreateTokenWithKey(sub string, role string, secret string) string {
	claims := jwt.MapClaims{
		"sub":  sub,
		"role": role,
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signed
}
