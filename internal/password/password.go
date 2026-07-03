package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	fmt.Printf("hash: %s\n", string(hash))
	fmt.Printf("plain: %s\n", plain)
	return string(hash), nil
}

func Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
