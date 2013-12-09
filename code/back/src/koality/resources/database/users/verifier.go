package users

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"regexp"
)

const (
	userMaxEmailLength     = 256
	userMaxFirstNameLength = 64
	userMaxLastNameLength  = 64
	userMaxKeyAliasLength  = 256
	userMaxPublicKeyLength = 1024
	emailRegex             = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$"
	firstNameRegex         = "^[-'a-zA-Z ]+$"
	lastNameRegex          = "^[-'a-zA-Z ]+$"
	keyAliasRegex          = "^[-'a-zA-Z ]+$"
	publicKeyRegex         = "^ssh-(?:dss|rsa) [A-Za-z0-9+/]+={0,2}"
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyEmail(email string) error {
	if len(email) > userMaxEmailLength {
		return fmt.Errorf("Email cannot exceed %d characters long", userMaxEmailLength)
	} else if ok, err := regexp.MatchString(emailRegex, email); !ok || err != nil {
		return errors.New("Email must match regex: " + emailRegex)
	} else if err := verifier.verifyUserDoesNotExistWithEmail(email); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyFirstName(firstName string) error {
	if len(firstName) > userMaxFirstNameLength {
		return fmt.Errorf("First name exceed %d characters long", userMaxFirstNameLength)
	} else if ok, err := regexp.MatchString(firstNameRegex, firstName); !ok || err != nil {
		return errors.New("First name must match regex: " + firstNameRegex)
	}
	return nil
}

func (verifier *Verifier) verifyLastName(lastName string) error {
	if len(lastName) > userMaxLastNameLength {
		return fmt.Errorf("Last name cannot exceed %d characters long", userMaxLastNameLength)
	} else if ok, err := regexp.MatchString(lastNameRegex, lastName); !ok || err != nil {
		return errors.New("Last name must match regex: " + lastNameRegex)
	}
	return nil
}

func (verifier *Verifier) verifyKeyAlias(userId uint64, alias string) error {
	if len(alias) > userMaxKeyAliasLength {
		return fmt.Errorf("Key alias cannot exceed %d characters long", userMaxKeyAliasLength)
	} else if ok, err := regexp.MatchString(keyAliasRegex, alias); !ok || err != nil {
		return errors.New("SSH Key alias must match regex: " + keyAliasRegex)
	} else if err := verifier.verifyKeyDoesNotExistWithAlias(userId, alias); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyPublicKey(publicKey string) error {
	if len(publicKey) > userMaxPublicKeyLength {
		return fmt.Errorf("Public key cannot exceed %d characters long", userMaxPublicKeyLength)
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
		return resources.UserAlreadyExistsError{"User already exists with email: " + email}
	}
	return nil
}

func (verifier *Verifier) verifyKeyDoesNotExistWithAlias(userId uint64, alias string) error {
	query := "SELECT id FROM ssh_keys WHERE user_id=$1 and alias=$2"
	err := verifier.database.QueryRow(query, userId, alias).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.KeyAlreadyExistsError{"SSH Public key already exists with alias: " + alias}
	}
	return nil
}

func (verifier *Verifier) verifyPublicKeyDoesNotExist(publicKey string) error {
	query := "SELECT id FROM ssh_keys WHERE public_key=$1"
	err := verifier.database.QueryRow(query, publicKey).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.KeyAlreadyExistsError{"SSH Public key already exists"}
	}
	return nil
}
