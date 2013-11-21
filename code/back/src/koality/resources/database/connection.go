package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Adds the Postgres driver
	"io/ioutil"
	"koality/resources"
	"koality/resources/database/users"
)

const (
	host         = "localhost"
	userName     = "koality"
	password     = "lt3"
	databaseName = "koality"
	sslMode      = "disable"
)

func New() (*resources.Connection, error) {
	paramsString := fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=%s", host, userName, password, databaseName, sslMode)
	database, err := sql.Open("postgres", paramsString)
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(10)

	err = setSchema(database)
	if err != nil {
		return nil, err
	}

	usersHandler, err := users.New(database)
	if err != nil {
		return nil, err
	}

	connection := resources.Connection{usersHandler}
	return &connection, nil
}

func setSchema(database *sql.DB) error {
	file, err := ioutil.ReadFile("schema.sql")
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
