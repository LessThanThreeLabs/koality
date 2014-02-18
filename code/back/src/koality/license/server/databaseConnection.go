package licenseserver

import (
	"database/sql"
	_ "github.com/lib/pq" // Adds the Postgres driver
	"io/ioutil"
	"koality/util/pathtranslator"
	"os"
	"path"
)

func GetDatabaseConnection(databaseConfiguration DatabaseConfiguration) (*sql.DB, error) {
	setEnvironment(databaseConfiguration)
	database, err := sql.Open("postgres", "")
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(10)
	database.SetMaxOpenConns(100)

	err = loadSchema(database)
	if err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func setEnvironment(databaseConfiguration DatabaseConfiguration) {
	os.Setenv("PGHOST", databaseConfiguration.Host)
	os.Setenv("PGUSER", databaseConfiguration.Username)
	os.Setenv("PGPASSWORD", databaseConfiguration.Password)
	os.Setenv("PGDATABASE", databaseConfiguration.DatabaseName)
	os.Setenv("PGSSLMODE", databaseConfiguration.SslMode)
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

func getSchemaLocation() (string, error) {
	return pathtranslator.TranslatePathAndCheckExists(path.Join("postgres", "license-server-schema.sql"))
}
