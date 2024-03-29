package users

import (
	"database/sql"
	"koality/resources"
)

type Scannable interface {
	Scan(dest ...interface{}) error
}

type ReadHandler struct {
	database            *sql.DB
	verifier            *Verifier
	subscriptionHandler resources.InternalUsersSubscriptionHandler
}

func NewReadHandler(database *sql.DB, verifier *Verifier, subscriptionHandler resources.InternalUsersSubscriptionHandler) (resources.UsersReadHandler, error) {
	return &ReadHandler{database, verifier, subscriptionHandler}, nil
}

func (readHandler *ReadHandler) scanUser(scannable Scannable) (*resources.User, error) {
	user := new(resources.User)

	var gitHubOAuth sql.NullString
	var deletedId uint64
	err := scannable.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.PasswordSalt,
		&gitHubOAuth, &user.IsAdmin, &user.Created, &deletedId)
	if err == sql.ErrNoRows {
		return nil, resources.NoSuchUserError{"Unable to find user"}
	} else if err != nil {
		return nil, err
	}

	user.IsDeleted = user.Id == deletedId

	if gitHubOAuth.Valid {
		user.GitHubOAuth = gitHubOAuth.String
	}
	return user, nil
}

func (readHandler *ReadHandler) Get(userId uint64) (*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, is_admin, created, deleted FROM users WHERE id=$1"
	row := readHandler.database.QueryRow(query, userId)
	return readHandler.scanUser(row)
}

func (readHandler *ReadHandler) GetByEmail(email string) (*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, is_admin, created, deleted FROM users WHERE email=$1 AND id != deleted"
	row := readHandler.database.QueryRow(query, email)
	return readHandler.scanUser(row)
}

func (readHandler *ReadHandler) GetIdByPublicKey(publicKey string) (uint64, error) {
	query := "SELECT user_id FROM ssh_keys WHERE public_key=$1"
	row := readHandler.database.QueryRow(query, publicKey)
	var userId uint64
	err := row.Scan(&userId)
	if err == sql.ErrNoRows {
		return 0, resources.NoSuchUserError{"Unable to find user for key"}
	} else if err != nil {
		return 0, err
	}
	return userId, nil
}

func (readHandler *ReadHandler) GetAll() ([]resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, is_admin, created, deleted FROM users WHERE id >= 1000 AND id != deleted"
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
	query := "SELECT id, name, public_key, created FROM ssh_keys WHERE user_id=$1"
	rows, err := readHandler.database.Query(query, userId)
	if err != nil {
		return nil, err
	}

	sshKeys := make([]resources.SshKey, 0, 5)
	for rows.Next() {
		sshKey := resources.SshKey{}
		err = rows.Scan(&sshKey.Id, &sshKey.Name, &sshKey.PublicKey, &sshKey.Created)
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
