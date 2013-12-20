package users

import (
	"database/sql"
	"errors"
	"fmt"
	"koality/resources"
	"regexp"
)

const (
	maxEmailLength     = 256
	maxFirstNameLength = 64
	maxLastNameLength  = 64
	maxKeyNameLength   = 256
	maxPublicKeyLength = 1024
	emailRegex         = "^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$"
	firstNameRegex     = "^[-'a-zA-Z ]+$"
	lastNameRegex      = "^[-'a-zA-Z ]+$"
	keyNameRegex       = "^[-'a-zA-Z ]+$"
	publicKeyRegex     = "^ssh-(?:dss|rsa) [A-Za-z0-9+/]+={0,2}"
)

type Verifier struct {
	database *sql.DB
}

func NewVerifier(database *sql.DB) (*Verifier, error) {
	return &Verifier{database}, nil
}

func (verifier *Verifier) verifyEmail(email string) error {
	if len(email) > maxEmailLength {
		return fmt.Errorf("Email cannot exceed %d characters long", maxEmailLength)
	} else if ok, err := regexp.MatchString(emailRegex, email); !ok || err != nil {
		return errors.New("Email must match regex: " + emailRegex)
	} else if err := verifier.verifyUserDoesNotExistWithEmail(email); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyFirstName(firstName string) error {
	if len(firstName) > maxFirstNameLength {
		return fmt.Errorf("First name exceed %d characters long", maxFirstNameLength)
	} else if ok, err := regexp.MatchString(firstNameRegex, firstName); !ok || err != nil {
		return errors.New("First name must match regex: " + firstNameRegex)
	}
	return nil
}

func (verifier *Verifier) verifyLastName(lastName string) error {
	if len(lastName) > maxLastNameLength {
		return fmt.Errorf("Last name cannot exceed %d characters long", maxLastNameLength)
	} else if ok, err := regexp.MatchString(lastNameRegex, lastName); !ok || err != nil {
		return errors.New("Last name must match regex: " + lastNameRegex)
	}
	return nil
}

func (verifier *Verifier) verifyKeyName(userId uint64, name string) error {
	if len(name) > maxKeyNameLength {
		return fmt.Errorf("SSH Key name cannot exceed %d characters long", maxKeyNameLength)
	} else if ok, err := regexp.MatchString(keyNameRegex, name); !ok || err != nil {
		return errors.New("SSH Key name must match regex: " + keyNameRegex)
	} else if err := verifier.verifyKeyDoesNotExistWithName(userId, name); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyPublicKey(publicKey string) error {
	if len(publicKey) > maxPublicKeyLength {
		return fmt.Errorf("Public key cannot exceed %d characters long", maxPublicKeyLength)
	} else if ok, err := regexp.MatchString(publicKeyRegex, publicKey); !ok || err != nil {
		return errors.New("SSH Public Key must match regex: " + publicKeyRegex)
	} else if err := verifier.verifyPublicKeyDoesNotExist(publicKey); err != nil {
		return err
	}
	return nil
}

func (verifier *Verifier) verifyUserExists(userId uint64) error {
	query := "SELECT id FROM users WHERE id=$1"
	err := verifier.database.QueryRow(query, userId).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		errorText := fmt.Sprintf("Unable to find user with id: %d", userId)
		return resources.NoSuchUserError{errorText}
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

func (verifier *Verifier) verifyKeyDoesNotExistWithName(userId uint64, name string) error {
	query := "SELECT id FROM ssh_keys WHERE user_id=$1 and name=$2"
	err := verifier.database.QueryRow(query, userId, name).Scan(new(uint64))
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err != sql.ErrNoRows {
		return resources.KeyAlreadyExistsError{"SSH Public key already exists with name: " + name}
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
