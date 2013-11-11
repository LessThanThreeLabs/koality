package users

import (
	"database/sql"
	"koality/resources"
)

type ReadHandler struct {
	database *sql.DB
}

func NewReadHandler(database *sql.DB) (resources.UsersReadHandler, error) {
	return &ReadHandler{database}, nil
}

func (readHandler *ReadHandler) Get(id int) (*resources.User, error) {
	user := new(resources.User)

	row := readHandler.database.QueryRow(
		"SELECT id, email, first_name, last_name, password_hash, password_salt,"+
			" github_oauth, admin, created, deleted FROM users WHERE id=$1",
		id)

	var gitHubOAuth sql.NullString
	err := row.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.PasswordSalt,
		&gitHubOAuth, &user.Admin, &user.Created, &user.Deleted)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchUserError(err)
	} else if err != nil {
		return nil, err
	}

	if gitHubOAuth.Valid {
		user.GitHubOauth = gitHubOAuth.String
	}

	return user, nil
}
