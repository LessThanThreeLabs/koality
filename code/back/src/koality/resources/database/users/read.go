package users

import (
	"database/sql"
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
	err := scannable.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.PasswordSalt,
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

func (readHandler *ReadHandler) Get(id int) (*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, admin, created, deleted FROM users WHERE id=$1"
	row := readHandler.database.QueryRow(query, id)
	return readHandler.scanUser(row)
}

func (readHandler *ReadHandler) GetByEmail(email string) (*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, admin, created, deleted FROM users WHERE email=$1"
	row := readHandler.database.QueryRow(query, email)
	return readHandler.scanUser(row)
}

func (readHandler *ReadHandler) GetAll() ([]*resources.User, error) {
	query := "SELECT id, email, first_name, last_name, password_hash, password_salt," +
		" github_oauth, admin, created, deleted FROM users WHERE id=$1"
	rows, err := readHandler.database.Query(query)
	if err != nil {
		return nil, err
	}

	users := make([]*resources.User, 100)
	for rows.Next() {
		user, err := readHandler.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
