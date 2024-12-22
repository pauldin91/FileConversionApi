package utils

import "golang.org/x/crypto/bcrypt"

func HashedPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashed), err
}

func IsPasswordValid(providedPassword string, hashedPassword string) error {

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}
