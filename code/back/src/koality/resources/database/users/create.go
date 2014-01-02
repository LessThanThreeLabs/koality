package users

import (
	"database/sql"
	"encoding/base64"
	"koality/resources"
)

type CreateHandler struct {
	database            *sql.DB
	verifier            *Verifier
	readHandler         resources.UsersReadHandler
	subscriptionHandler resources.InternalUsersSubscriptionHandler
}

func NewCreateHandler(database *sql.DB, verifier *Verifier, readHandler resources.UsersReadHandler,
	subscriptionHandler resources.InternalUsersSubscriptionHandler) (resources.UsersCreateHandler, error) {

	return &CreateHandler{database, verifier, readHandler, subscriptionHandler}, nil
}

func (createHandler *CreateHandler) Create(email, firstName, lastName string, passwordHash, passwordSalt []byte, admin bool) (*resources.User, error) {
	err := createHandler.getUserParamsError(email, firstName, lastName)
	if err != nil {
		return nil, err
	}

	passwordHashBase64 := base64.StdEncoding.EncodeToString(passwordHash)
	passwordSaltBase64 := base64.StdEncoding.EncodeToString(passwordSalt)

	id := uint64(0)
	query := "INSERT INTO users (email, first_name, last_name, password_hash, password_salt, is_admin)" +
		" VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err = createHandler.database.QueryRow(query, email, firstName, lastName, passwordHashBase64, passwordSaltBase64, admin).Scan(&id)
	if err != nil {
		return nil, err
	}

	user, err := createHandler.readHandler.Get(id)
	if err != nil {
		return nil, err
	}

	createHandler.subscriptionHandler.FireCreatedEvent(user)
	return user, nil
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
