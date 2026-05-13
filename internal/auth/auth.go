package auth

import (
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const task = `
func HashPassword(password string) (string, error): Hash the password using the argon2id.CreateHash 
function.

func CheckPasswordHash(password, hash string) (bool, error): Use the argon2id.ComparePasswordAndHash 
function to compare the password that the user entered in the HTTP request with the password that is 
stored in the database.
`

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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	time_UTC := time.Now().UTC()
	newToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy-access",
			IssuedAt:  jwt.NewNumericDate(time_UTC),
			ExpiresAt: jwt.NewNumericDate(time_UTC.Add(expiresIn)),
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

	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		secretInBytes := []byte(tokenSecret)
		return secretInBytes, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); !ok {
		return uuid.UUID{}, fmt.Errorf("cannot get claims")
	} else {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.UUID{}, err
		}
		return userID, nil
	}
}
