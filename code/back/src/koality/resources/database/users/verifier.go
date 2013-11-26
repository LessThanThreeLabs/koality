package users

import (
	"database/sql"
	"errors"
	"koality/resources"
	"regexp"
)

const (
	emailRegex     = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+(?:[a-z]{2}|com|org|net|edu|gov|mil|biz|info|mobi|name|aero|asia|jobs|museum)\\b$"
	firstNameRegex = "^[-'a-zA-Z ]+$"
	lastNameRegex  = "^[-'a-zA-Z ]+$"
	keyAliasRegex  = "^[-'a-zA-Z ]+$"
	publicKeyRegex = "^ssh-(?:dss|rsa) [A-Za-z0-9+/]+={0,2}"
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyEmail(email string) error {
	if len(email) > 256 {
		return errors.New("Email must be less than 256 characters long")
	} else if ok, err := regexp.MatchString(emailRegex, email); !ok || err != nil {
		return errors.New("Email must match regex: " + emailRegex)
	} else if verifier.doesUserExistWithEmail(email) {
		return resources.UserAlreadyExistsError{errors.New("User already exists with email: " + email)}
	}
	return nil
}

func (verifier *Verifier) verifyFirstName(firstName string) error {
	if len(firstName) > 64 {
		return errors.New("First name must be less than 64 characters long")
	} else if ok, err := regexp.MatchString(firstNameRegex, firstName); !ok || err != nil {
		return errors.New("First name must match regex: " + firstNameRegex)
	}
	return nil
}

func (verifier *Verifier) verifyLastName(lastName string) error {
	if len(lastName) > 64 {
		return errors.New("Last name must be less than 64 characters long")
	} else if ok, err := regexp.MatchString(lastNameRegex, lastName); !ok || err != nil {
		return errors.New("Last name must match regex: " + lastNameRegex)
	}
	return nil
}

func (verifier *Verifier) verifyKeyAlias(userId uint64, alias string) error {
	if len(alias) > 256 {
		return errors.New("Key alias must be less than 256 characters long")
	} else if ok, err := regexp.MatchString(keyAliasRegex, alias); !ok || err != nil {
		return errors.New("SSH Key alias must match regex: " + keyAliasRegex)
	} else if verifier.doesKeyExistWithUserAndAlias(userId, alias) {
		return resources.KeyAlreadyExistsError{errors.New("User already has SSH Key with alias: " + alias)}
	}
	return nil
}

func (verifier *Verifier) verifyPublicKey(publicKey string) error {
	if len(publicKey) > 1024 {
		return errors.New("Public key must be less than 1024 characters long")
	} else if ok, err := regexp.MatchString(publicKeyRegex, publicKey); !ok || err != nil {
		return errors.New("SSH Public Key must match regex: " + publicKeyRegex)
	} else if verifier.doesKeyExistWithPublicKey(publicKey) {
		return resources.KeyAlreadyExistsError{errors.New("SSH Public key already exists")}
	}
	return nil
}

func (verifier *Verifier) doesUserExistWithEmail(email string) bool {
	query := "SELECT id FROM users WHERE email=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, email).Scan()
	return err != sql.ErrNoRows
}

func (verifier *Verifier) doesKeyExistWithUserAndAlias(userId uint64, alias string) bool {
	query := "SELECT id FROM ssh_keys WHERE user_id=$1 and alias=$2"
	err := verifier.database.QueryRow(query, userId, alias).Scan()
	return err != sql.ErrNoRows
}

func (verifier *Verifier) doesKeyExistWithPublicKey(publicKey string) bool {
	query := "SELECT id FROM ssh_keys WHERE public_key=$1"
	err := verifier.database.QueryRow(query, publicKey).Scan()
	return err != sql.ErrNoRows
}
