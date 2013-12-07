package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Adds the Postgres driver
	"io/ioutil"
	"koality/resources"
	"koality/resources/database/repositories"
	"koality/resources/database/stages"
	"koality/resources/database/users"
	"koality/resources/database/verifications"
	"os/user"
)

const (
	host         = "localhost"
	userName     = "koality"
	password     = "lt3"
	databaseName = "koality"
	sslMode      = "disable"
)

func New() (*resources.Connection, error) {
	database, err := getDatabaseConnection()
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(100)
	database.SetMaxOpenConns(100)

	err = setSchema(database)
	if err != nil {
		return nil, err
	}

	usersHandler, err := users.New(database)
	if err != nil {
		return nil, err
	}

	repositoriesHandler, err := repositories.New(database)
	if err != nil {
		return nil, err
	}

	verificationsHandler, err := verifications.New(database)
	if err != nil {
		return nil, err
	}

	stagesHandler, err := stages.New(database)
	if err != nil {
		return nil, err
	}

	connection := resources.Connection{usersHandler, repositoriesHandler, verificationsHandler, stagesHandler}
	return &connection, nil
}

func Reseed() error {
	database, err := getDatabaseConnection()
	if err != nil {
		return err
	}

	_, err = database.Exec("DROP SCHEMA public CASCADE")
	if err != nil {
		return err
	}
	return nil
}

func getDatabaseConnection() (*sql.DB, error) {
	paramsString := fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=%s", host, userName, password, databaseName, sslMode)
	return sql.Open("postgres", paramsString)
}

func setSchema(database *sql.DB) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	file, err := ioutil.ReadFile(currentUser.HomeDir + "/postgres/schema.sql")
	if err != nil {
		return err
	}
	schemaQuery := string(file)

	// Just in case the public schema was deleted, which
	// is oftentimes done as a shortcut to drop all tables
	_, err = database.Exec("CREATE SCHEMA IF NOT EXISTS public")
	if err != nil {
		return err
	}

	_, err = database.Exec(schemaQuery)
	if err != nil {
		return err
	}

	return nil
}
