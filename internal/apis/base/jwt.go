package base

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

const (
	TokenIssuer         = "hyperdot"
	TokenDefaultSubject = "hyperdot-fronted"
	todoSecret          = "hyperdot"
)

func TokenDefaultExpireTime() time.Time {
	return time.Now().Add(time.Hour * 24)
}

func GenerateJwtToken(claims *datamodel.UserClaims, expireAt time.Time) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    TokenIssuer,
		Subject:   TokenDefaultSubject,
		ExpiresAt: jwt.NewNumericDate(expireAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(todoSecret)) // TODO: use generate
}

func VerifyJwtToken(tokenString string) (*datamodel.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &datamodel.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(todoSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*datamodel.UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
