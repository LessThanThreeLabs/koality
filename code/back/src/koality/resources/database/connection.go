package database

import (
	"database/sql"
	_ "github.com/lib/pq" // Adds the Postgres driver

	"io/ioutil"
	"koality/resources"
	"koality/resources/database/repositories"
	"koality/resources/database/stages"
	"koality/resources/database/users"
	"koality/resources/database/verifications"
	"os"
	"os/exec"
	"os/user"
)

const (
	host         = "localhost"
	userName     = "koality"
	password     = "lt3"
	databaseName = "koality"
	sslMode      = "disable"
)

const (
	dumpLocation   = "/postgres/backup.tar"
	schemaLocation = "/postgres/schema.sql"
)

func New() (*resources.Connection, error) {
	database, err := getDatabaseConnection()
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(10)
	database.SetMaxOpenConns(100)

	err = loadSchema(database)
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

	_, err = database.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	if err != nil {
		return err
	}

	err = loadSchema(database)
	return err
}

func SaveDump() error {
	setEnvironment()

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	command := exec.Command("pg_dump", "--file", currentUser.HomeDir+dumpLocation, "--format", "custom")
	err = command.Run()
	return err
}

func LoadDump() error {
	database, err := getDatabaseConnection()
	if err != nil {
		return err
	}

	_, err = database.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE SCHEMA IF NOT EXISTS public")
	if err != nil {
		return err
	}

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	command := exec.Command("pg_restore", "--dbname", databaseName, "--jobs", "4", currentUser.HomeDir+dumpLocation)
	err = command.Run()
	return err
}

func setEnvironment() {
	os.Setenv("PGHOST", host)
	os.Setenv("PGUSER", userName)
	os.Setenv("PGPASSWORD", password)
	os.Setenv("PGDATABASE", databaseName)
	os.Setenv("PGSSLMODE", sslMode)
}

func getDatabaseConnection() (*sql.DB, error) {
	setEnvironment()
	return sql.Open("postgres", "")
}

func loadSchema(database *sql.DB) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	file, err := ioutil.ReadFile(currentUser.HomeDir + schemaLocation)
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
	return err
}
