package users

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"koality/resources"
)

type CreateHandler struct {
	database *sql.DB
}

func NewCreateHandler(database *sql.DB) (resources.UsersCreateHandler, error) {
	return &CreateHandler{database}, nil
}

func (createHandler *CreateHandler) Create(email, firstName, lastName string, passwordHash, passwordSalt []byte, admin bool) (int64, error) {
	if createHandler.doesUserExistWithEmail(email) {
		return -1, resources.UserAlreadyExistsError(errors.New("User already exists with email: " + email))
	}

	passwordHashBase64 := base64.StdEncoding.EncodeToString(passwordHash)
	passwordSaltBase64 := base64.StdEncoding.EncodeToString(passwordSalt)

	id := int64(0)
	query := "INSERT INTO users (email, first_name, last_name, password_hash, password_salt, is_admin)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err := createHandler.database.QueryRow(query, email, firstName, lastName, passwordHashBase64, passwordSaltBase64, admin).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (createHandler *CreateHandler) doesUserExistWithEmail(email string) bool {
	query := "SELECT id FROM users WHERE email=$1 AND deleted=0"
	err := createHandler.database.QueryRow(query, email).Scan()
	return err != sql.ErrNoRows
}
