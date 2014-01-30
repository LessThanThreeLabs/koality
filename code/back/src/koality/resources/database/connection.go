package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Adds the Postgres driver
	"io/ioutil"
	"koality/resources"
	"koality/resources/database/pools"
	"koality/resources/database/repositories"
	"koality/resources/database/settings"
	"koality/resources/database/snapshots"
	"koality/resources/database/stages"
	"koality/resources/database/users"
	"koality/resources/database/verifications"
	"koality/util/pathtranslator"
	"os"
	"os/exec"
	"path"
	"time"
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

	poolsHandler, err := pools.New(database)
	if err != nil {
		return nil, err
	}

	settingsHandler, err := settings.New(database)
	if err != nil {
		return nil, err
	}

	snapshotsHandler, err := snapshots.New(database, verificationsHandler)
	if err != nil {
		return nil, err
	}

	connection := &resources.Connection{
		Users:         usersHandler,
		Repositories:  repositoriesHandler,
		Verifications: verificationsHandler,
		Stages:        stagesHandler,
		Pools:         poolsHandler,
		Settings:      settingsHandler,
		Snapshots:     snapshotsHandler,
		Closer:        database,
	}

	if err = checkSettingsInitialized(connection); err != nil {
		return nil, err
	}

	return connection, nil
}

func getSchemaLocation() (string, error) {
	relativePath := path.Join("postgres", "schema.sql")
	return pathtranslator.TranslatePathAndCheckExists(relativePath)
}

func getDumpLocation() (string, error) {
	relativePath := path.Join("postgres", "backup.tar")
	return pathtranslator.TranslatePath(relativePath)
}

func checkSettingsInitialized(connection *resources.Connection) error {
	_, err := connection.Settings.Read.GetRepositoryKeyPair()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		if _, err = connection.Settings.Update.ResetRepositoryKeyPair(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = connection.Settings.Read.GetCookieStoreKeys()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		if _, err = connection.Settings.Update.ResetCookieStoreKeys(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = connection.Settings.Read.GetApiKey()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		if _, err = connection.Settings.Update.ResetApiKey(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func Reseed() error {
	database, err := getDatabaseConnection()
	if err != nil {
		return err
	}
	defer database.Close()

	_, err = database.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	if err != nil {
		return err
	}

	err = loadSchema(database)
	return err
}

func DumpExistsAndNotStale(staleTime time.Time) (bool, error) {
	dumpLocation, err := getDumpLocation()
	if err != nil {
		return false, err
	}

	fileInfo, err := os.Stat(dumpLocation)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else if fileInfo.ModTime().Before(staleTime) {
		return false, nil
	} else {
		return true, nil
	}
}

func CreateDump() error {
	setEnvironment()

	dumpLocation, err := getDumpLocation()
	if err != nil {
		return err
	}

	command := exec.Command("pg_dump", "--file", dumpLocation, "--format", "custom")
	output, err := command.CombinedOutput()
	if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%v: %s", err, output)
	}
	return err
}

func LoadDump() error {
	database, err := getDatabaseConnection()
	if err != nil {
		return err
	}
	defer database.Close()

	_, err = database.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	if err != nil {
		return err
	}

	_, err = database.Exec("CREATE SCHEMA IF NOT EXISTS public")
	if err != nil {
		return err
	}

	dumpLocation, err := getDumpLocation()
	if err != nil {
		return err
	}

	command := exec.Command("pg_restore", "--dbname", databaseName, "--jobs", "4", dumpLocation)
	output, err := command.CombinedOutput()
	if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%v: %s", err, output)
	}
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
	schemaLocation, err := getSchemaLocation()
	if err != nil {
		return err
	}

	file, err := ioutil.ReadFile(schemaLocation)
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
