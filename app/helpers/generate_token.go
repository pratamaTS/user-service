package helpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

func GenerateJWT(
	subject, jwtSecret string,
	ttl time.Duration,
	tokenType TokenType,
	sessionID string,
) (token string, exp int64, err error) {
	now := time.Now()
	expTS := now.Add(ttl).Unix()

	claims := jwt.MapClaims{
		"subject":   subject,
		"type":      string(tokenType),
		"sessionId": sessionID,
		"exp":       expTS,
		"iat":       now.Unix(),
	}

	j := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := j.SignedString([]byte(jwtSecret))
	return tok, expTS, err
}
