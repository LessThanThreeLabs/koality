package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Adds the Postgres driver
	"io/ioutil"
)

const (
	host         = "localhost"
	userName     = "koality"
	password     = "lt3"
	databaseName = "koality"
	sslMode      = "disable"
)

func New() error {
	file, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	schemaQuery := string(file)

	paramsString := fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=%s", host, userName, password, databaseName, sslMode)
	database, err := sql.Open("postgres", paramsString)
	if err != nil {
		return err
	}

	_, err = database.Exec(schemaQuery)
	if err != nil {
		return err
	}

	return nil
}
