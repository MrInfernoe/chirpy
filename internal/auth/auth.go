package auth

import "github.com/alexedwards/argon2id"

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
