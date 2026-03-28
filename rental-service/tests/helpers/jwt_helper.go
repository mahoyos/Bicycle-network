package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	PublicPEM  string
)

func init() {
	var err error
	PrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	PublicKey = &PrivateKey.PublicKey

	pubBytes, err := x509.MarshalPKIXPublicKey(PublicKey)
	if err != nil {
		panic(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	PublicPEM = string(pubPEM)
}

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
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(PrivateKey)
	if err != nil {
		panic(err)
	}
	return signed
}
