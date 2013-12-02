package users

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"koality/resources"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

type ReadHandler struct {
	database *sql.DB
}

func NewReadHandler(database *sql.DB) (resources.UsersReadHandler, error) {
	return &ReadHandler{database}, nil
}

func (readHandler *ReadHandler) scanUser(scannable Scannable) (*resources.User, error) {
	user := new(resources.User)

	var gitHubOAuth sql.NullString
	var passwordHashBase64, passwordSaltBase64 string
	err := scannable.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &passwordHashBase64, &passwordSaltBase64,
		&gitHubOAuth, &user.IsAdmin, &user.Created)
	if err == sql.ErrNoRows {
		errorText := "Unable to find user"
		return nil, resources.NoSuchUserError{errors.New(errorText)}
	} else if err != nil {
		return nil, err
	}

	if gitHubOAuth.Valid {
		user.GitHubOauth = gitHubOAuth.String
	}

	passwordHash, err := base64.StdEncoding.DecodeString(passwordHashBase64)
	if err != nil {
		return nil, err
	} else {
		user.PasswordHash = passwordHash
	}

	passwordSalt, err := base64.StdEncoding.DecodeString(passwordSaltBase64)
	if err != nil {
		return nil, err
	} else {
		user.PasswordSalt = passwordSalt
	}

	return user, nil
}

func (readHandler *ReadHandler) Get(userId uint64) (*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, is_admin, created FROM users WHERE id=$1"
	row := readHandler.database.QueryRow(query, userId)
	return readHandler.scanUser(row)
}

func (readHandler *ReadHandler) GetByEmail(email string) (*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, is_admin, created FROM users WHERE email=$1 AND id != deleted"
	row := readHandler.database.QueryRow(query, email)
	return readHandler.scanUser(row)
}

func (readHandler *ReadHandler) GetAll() ([]resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, is_admin, created FROM users WHERE id >= 1000 AND id != deleted"
	rows, err := readHandler.database.Query(query)
	if err != nil {
		return nil, err
	}

	users := make([]resources.User, 0, 10)
	for rows.Next() {
		user, err := readHandler.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (readHandler *ReadHandler) GetKeys(userId uint64) ([]resources.SshKey, error) {
	query := "SELECT id, alias, public_key, created FROM ssh_keys WHERE user_id=$1"
	rows, err := readHandler.database.Query(query, userId)
	if err != nil {
		return nil, err
	}

	sshKeys := make([]resources.SshKey, 0, 5)
	for rows.Next() {
		sshKey := resources.SshKey{}
		err = rows.Scan(&sshKey.Id, &sshKey.Alias, &sshKey.PublicKey, &sshKey.Created)
		if err != nil {
			return nil, err
		}
		sshKeys = append(sshKeys, sshKey)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sshKeys, nil
}
