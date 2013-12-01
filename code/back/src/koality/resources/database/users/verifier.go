package users

import (
	"database/sql"
	"errors"
	"koality/resources"
	"regexp"
)

const (
	emailRegex     = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$"
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
	} else if err := verifier.verifyUserDoesNotExistWithEmail(email); err != nil {
		return err
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
	} else if err := verifier.verifyKeyDoesNotExistWithAlias(userId, alias); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyPublicKey(publicKey string) error {
	if len(publicKey) > 1024 {
		return errors.New("Public key must be less than 1024 characters long")
	} else if ok, err := regexp.MatchString(publicKeyRegex, publicKey); !ok || err != nil {
		return errors.New("SSH Public Key must match regex: " + publicKeyRegex)
	} else if err := verifier.verifyPublicKeyDoesNotExist(publicKey); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyUserDoesNotExistWithEmail(email string) error {
	query := "SELECT id FROM users WHERE email=$1 AND deleted=0"
	err := verifier.database.QueryRow(query, email).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := "User already exists with email: " + email
		return resources.UserAlreadyExistsError{errors.New(errorText)}
	}
	return nil
}

func (verifier *Verifier) verifyKeyDoesNotExistWithAlias(userId uint64, alias string) error {
	query := "SELECT id FROM ssh_keys WHERE user_id=$1 and alias=$2"
	err := verifier.database.QueryRow(query, userId, alias).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := "SSH Public key already exists with alias: " + alias
		return resources.KeyAlreadyExistsError{errors.New(errorText)}
	}
	return nil
}

func (verifier *Verifier) verifyPublicKeyDoesNotExist(publicKey string) error {
	query := "SELECT id FROM ssh_keys WHERE public_key=$1"
	err := verifier.database.QueryRow(query, publicKey).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		errorText := "SSH Public key already exists"
		return resources.KeyAlreadyExistsError{errors.New(errorText)}
	}
	return nil
}
