package database

import (
	"database/sql"
	"fmt"
	"koality/util/log"
)

type Migration struct {
	FromVersion uint64
	ToVersion   uint64
	Migrate     func(transaction *sql.Tx) error
}

func Migrate(migrations []Migration) error {
	migrationsMap, err := toMigrationsMap(migrations)
	if err != nil {
		return err
	}

	database, err := getDatabaseConnection()
	if err != nil {
		return err
	}

	originalVersion, err := getDatabaseVersion(database)
	if err != nil {
		return err
	}

	transaction, err := database.Begin()
	if err != nil {
		return err
	}

	finalVersion, err := runMigrations(originalVersion, migrationsMap, transaction)
	if err != nil {
		rollbackErr := transaction.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v\n%v", err, rollbackErr)
		}
		return err
	}

	if finalVersion == originalVersion {
		log.Infof("Already at latest database version: %d\n", finalVersion)
	} else {
		_, err = transaction.Exec("UPDATE version SET version=$1", finalVersion)
		if err == nil {
			err = transaction.Commit()
		}
		if err != nil {
			fmt.Printf("Failed to migrate database from version %d to %d\n", originalVersion, finalVersion)
			rollbackErr := transaction.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("Failed to migrate database from version %d to %d\n%v\n%v", originalVersion, finalVersion, err, rollbackErr)
			}
			return fmt.Errorf("Failed to migrate database from version %d to %d\n%v", originalVersion, finalVersion, err)
		}
		log.Infof("Migrated database from version %d to %d\n", originalVersion, finalVersion)
	}
	return nil
}

func runMigrations(originalVersion uint64, migrationsMap map[uint64]Migration, transaction *sql.Tx) (uint64, error) {
	version := originalVersion
	for {
		migration, ok := migrationsMap[version]
		if !ok {
			return version, nil
		}
		if err := migration.Migrate(transaction); err != nil {
			return version, fmt.Errorf("Failed to migrate database from version %d to %d, rolling back to %d\n%v", version, migration.ToVersion, originalVersion, err)
		}
		version = migration.ToVersion
	}
}

func toMigrationsMap(migrations []Migration) (map[uint64]Migration, error) {
	migrationsMap := make(map[uint64]Migration, len(migrations))
	for _, migration := range migrations {
		_, ok := migrationsMap[migration.FromVersion]
		if ok {
			return nil, fmt.Errorf("Found multiple migrations which have FromVersion: %d", migration.FromVersion)
		}

		migrationsMap[migration.FromVersion] = migration
	}
	return migrationsMap, nil
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
