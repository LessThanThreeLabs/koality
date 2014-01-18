package migrate

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	FromVersion uint64
	ToVersion   uint64
	Migrate     func(transaction *sql.Tx) error
}

var Migrations map[uint64]Migration
var migrations = []Migration{migrateV0ToV1}

func init() {
	Migrations = make(map[uint64]Migration, len(migrations))
	for _, migration := range migrations {
		_, ok := Migrations[migration.FromVersion]
		if ok {
			panic(fmt.Sprintf("Found multiple migrations which have FromVersion: %d", migration.FromVersion))
		}

		Migrations[migration.FromVersion] = migration
	}
}

var migrateV0ToV1 = Migration{0, 1,
	func(transaction *sql.Tx) error {
		_, err := transaction.Exec("SELECT 1 FROM users") // This is an example migration that does nothing
		return err
	},
}
