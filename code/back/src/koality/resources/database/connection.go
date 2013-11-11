package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Adds the Postgres driver
	"time"
)

const (
	host         = "localhost"
	userName     = "koality"
	password     = "lt3"
	databaseName = "koality"
	sslMode      = "disable"
)

func New() error {
	fmt.Println("connecting...")
	paramsString := fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=%s", host, userName, password, databaseName, sslMode)
	database, err := sql.Open("postgres", paramsString)
	if err != nil {
		fmt.Println("1")
		fmt.Println(err)
		return err
	}

	fmt.Println(database)

	err = database.Ping()
	if err != nil {
		fmt.Println("1.5")
		fmt.Println(err)
		return err
	}

	_, err = database.Exec("CREATE TABLE users (name varchar[])")
	if err != nil {
		fmt.Println("2")
		fmt.Println(err)
		return err
	}

	time.Sleep(1 * time.Second)
	return nil
}
