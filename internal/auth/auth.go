package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	params := argon2id.DefaultParams
	hash, err := argon2id.CreateHash(password, params)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	if match {
		return match, nil
	} else {
		return match, nil
	}
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {

	currentTime := time.Now().UTC()
	newToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy-access",
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(1 * time.Hour)),
			Subject:   userID.String(),
		},
	)
	secretInBytes := []byte(tokenSecret)

	newJWT, err := newToken.SignedString(secretInBytes)
	if err != nil {
		return "", err
	}

	return newJWT, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	regClaims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &regClaims, func(token *jwt.Token) (any, error) {
		secretInBytes := []byte(tokenSecret)
		return secretInBytes, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); !ok {
		return uuid.UUID{}, fmt.Errorf("cannot get claims")
	} else {
		if time.Now().UTC().After(claims.ExpiresAt.Time.UTC()) {
			return uuid.UUID{}, fmt.Errorf("expired")
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.UUID{}, err
		}
		return userID, nil
	}
}

func GetBearerToken(headers http.Header) (string, error) {

	authHeader := headers.Get("Authorization")
	fields := strings.Fields(authHeader)
	var token string
	for i, field := range fields {
		if field == "Bearer" {
			token = fields[i+1]
			break
		}
	}
	if token == "" {
		return "", fmt.Errorf("could not find token")
	}
	return token, nil

	// tokenString := headers.Get("Authorization")
	// if tokenString == "" {
	// 	return "", fmt.Errorf("no authorization token found")
	// }
	// tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	// return tokenString, nil
}

func MakeRefreshToken() string {

	ranBytes := make([]byte, 32)
	rand.Read(ranBytes)
	return hex.EncodeToString(ranBytes)
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	fields := strings.Fields(authHeader)
	var APIKey string
	for i, field := range fields {
		if field == "ApiKey" {
			APIKey = fields[i+1]
			break
		}
	}
	if APIKey == "" {
		return "", fmt.Errorf("could not find ApiKey")
	}
	return APIKey, nil
}
