package database

import (
	"database/sql"
	"fmt"
	"testing"
)

func invertAdmins(transaction *sql.Tx) error {
	_, err := transaction.Exec("UPDATE users SET is_admin = NOT is_admin")
	return err
}

func failingMigration(transaction *sql.Tx) error {
	return fmt.Errorf("This is a dummy migration failure")
}

func testMigrations(test *testing.T, database *sql.DB, migrations []Migration, expectedVersion uint64, expectError bool) {
	err := Migrate(migrations)
	if err != nil && !expectError {
		test.Fatal(err)
	} else if err == nil && expectError {
		test.Fatal("Expected migrations to fail")
	}

	databaseVersion, err := getDatabaseVersion(database)
	if databaseVersion != expectedVersion {
		test.Fatalf("Expected database version %d, was %d", expectedVersion, databaseVersion)
	}
}

func TestSingleMigration(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	database, err := getDatabaseConnection()
	if err != nil {
		test.Fatal(err)
	}

	databaseVersion, err := getDatabaseVersion(database)
	if err != nil {
		test.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	migrations := []Migration{
		Migration{databaseVersion, databaseVersion + 1, invertAdmins},
	}

	testMigrations(test, database, migrations, databaseVersion+1, false)

	for _, user := range users {
		newUser, err := connection.Users.Read.Get(user.Id)
		if err != nil {
			test.Fatal(err)
		}
		if user.IsAdmin == newUser.IsAdmin {
			test.Fatalf("Migration did not change database properties correctly for user: %#v", user)
		}
	}
}

func TestMultiplePassingMigrations(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	database, err := getDatabaseConnection()
	if err != nil {
		test.Fatal(err)
	}

	databaseVersion, err := getDatabaseVersion(database)
	if err != nil {
		test.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	migrations := []Migration{
		Migration{databaseVersion, databaseVersion + 1, invertAdmins},
		Migration{databaseVersion + 1, databaseVersion + 2, invertAdmins},
	}

	testMigrations(test, database, migrations, databaseVersion+2, false)

	for _, user := range users {
		newUser, err := connection.Users.Read.Get(user.Id)
		if err != nil {
			test.Fatal(err)
		}
		if user.IsAdmin != newUser.IsAdmin {
			test.Fatalf("Migrations should not have changed user: %v, is now %v", user, newUser)
		}
	}
}

func TestMultipleMigrationsWithFailure(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	database, err := getDatabaseConnection()
	if err != nil {
		test.Fatal(err)
	}

	databaseVersion, err := getDatabaseVersion(database)
	if err != nil {
		test.Fatal(err)
	}

	users, err := connection.Users.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	migrations := []Migration{
		Migration{databaseVersion, databaseVersion + 1, invertAdmins},
		Migration{databaseVersion + 1, databaseVersion + 2, failingMigration},
	}

	testMigrations(test, database, migrations, databaseVersion, true)

	for _, user := range users {
		newUser, err := connection.Users.Read.Get(user.Id)
		if err != nil {
			test.Fatal(err)
		}
		if user.IsAdmin != newUser.IsAdmin {
			test.Fatalf("Migrations should not have changed user: %v, is now %v", user, newUser)
		}
	}
}

func TestMigrationSanity(test *testing.T) {
	migrations := []Migration{
		Migration{0, 1, invertAdmins},
		Migration{0, 2, invertAdmins},
	}
	if err := Migrate(migrations); err == nil {
		test.Fatal("Expected migration to fail with a duplicate FromVersion")
	}
}
