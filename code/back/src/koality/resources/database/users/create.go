package users

import (
	"database/sql"
	"encoding/base64"
	"koality/resources"
	"math"
)

type CreateHandler struct {
	database *sql.DB
	verifier *Verifier
}

func NewCreateHandler(database *sql.DB) (resources.UsersCreateHandler, error) {
	verifier, err := NewVerifier(database)
	if err != nil {
		return nil, err
	}
	return &CreateHandler{database, verifier}, nil
}

func (createHandler *CreateHandler) Create(email, firstName, lastName string, passwordHash, passwordSalt []byte, admin bool) (uint64, error) {
	err := createHandler.getUserParamsError(email, firstName, lastName)
	if err != nil {
		return math.MaxUint64, err
	}

	passwordHashBase64 := base64.StdEncoding.EncodeToString(passwordHash)
	passwordSaltBase64 := base64.StdEncoding.EncodeToString(passwordSalt)

	id := uint64(0)
	query := "INSERT INTO users (email, first_name, last_name, password_hash, password_salt, is_admin)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err = createHandler.database.QueryRow(query, email, firstName, lastName, passwordHashBase64, passwordSaltBase64, admin).Scan(&id)
	if err != nil {
		return math.MaxUint64, err
	}
	return id, nil
}

func (createHandler *CreateHandler) getUserParamsError(email, firstName, lastName string) error {
	if err := createHandler.verifier.verifyEmail(email); err != nil {
		return err
	} else if err := createHandler.verifier.verifyFirstName(firstName); err != nil {
		return err
	} else if err := createHandler.verifier.verifyLastName(lastName); err != nil {
		return err
	}
	return nil
}