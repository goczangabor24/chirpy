package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hashedPassword, nil
}

func CheckPassword(password string, hash string) (bool, error) {
	hashedPassoword, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	if hashedPassoword == hash {
		return true, nil
	} else {
		return false, nil
	}
}
