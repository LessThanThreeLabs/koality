package resources

import (
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"io"
)

type PasswordHasher struct {
	hasher hash.Hash
}

func NewPasswordHasher() (*PasswordHasher, error) {
	hasher := sha256.New()
	return &PasswordHasher{hasher}, nil
}

func (passwordHasher *PasswordHasher) GenerateHashAndSalt(password string) ([]byte, []byte, error) {
	passwordHasher.hasher.Reset()

	salt, err := passwordHasher.generateSalt()
	if err != nil {
		return nil, nil, err
	}

	_, err = passwordHasher.hasher.Write(salt)
	if err != nil {
		return nil, nil, err
	}

	_, err = io.WriteString(passwordHasher.hasher, password)
	if err != nil {
		return nil, nil, err
	}

	hash := passwordHasher.hasher.Sum(nil)
	return hash, salt, nil
}

func (passwordHasher *PasswordHasher) generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	return salt, err
}

func (passwordHasher *PasswordHasher) ComputeHash(password string, salt []byte) ([]byte, error) {
	passwordHasher.hasher.Reset()

	_, err := passwordHasher.hasher.Write(salt)
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(passwordHasher.hasher, password)
	if err != nil {
		return nil, err
	}

	hash := passwordHasher.hasher.Sum(nil)
	return hash, nil
}
