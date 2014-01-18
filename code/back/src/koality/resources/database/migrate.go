package database

import (
	"database/sql"
	"fmt"
	"koality/resources/database/migrate"
)

func Migrate() error {
	database, err := getDatabaseConnection()
	if err != nil {
		return err
	}

	version, err := getDatabaseVersion(database)
	if err != nil {
		return err
	}

	originalVersion := version

	transaction, err := database.Begin()
	if err != nil {
		return err
	}

	for {
		migration, ok := migrate.Migrations[version]
		if !ok {
			if version != originalVersion {
				_, err = transaction.Exec("UPDATE version SET version=$1", version)
				if err == nil {
					err = transaction.Commit()
				}
				if err != nil {
					fmt.Printf("Failed to migrate database from version %d to %d\n", originalVersion, version)
					txErr := transaction.Rollback()
					if txErr != nil {
						return txErr
					}
					return err
				}
				fmt.Printf("Migrated database from version %d to %d\n", originalVersion, version)
				return nil
			}
			fmt.Printf("Already at latest database version: %d\n", version)
			return nil
		}

		err = migration.Migrate(transaction)
		if err != nil {
			fmt.Printf("Failed to migrate database from version %d to %d, rolling back to %d\n", version, migration.ToVersion, originalVersion)
			txErr := transaction.Rollback()
			if txErr != nil {
				return txErr
			}
			return err
		}
		version = migration.ToVersion
	}
}

func getDatabaseVersion(database *sql.DB) (uint64, error) {
	var version uint64
	query := "SELECT version FROM version"
	row := database.QueryRow(query)
	err := row.Scan(&version)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("Unable to determine database version")
	} else if err != nil {
		return 0, err
	}

	return version, nil
}
