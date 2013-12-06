package main

import (
	"fmt"
	"koality/resources"
	"koality/resources/database"
	"time"
)

const (
	numRepositories = 3
)

func main() {
	fmt.Printf("Reseeding database (%s)\n", getTimeString())
	database.Reseed()

	connection, err := database.New()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Beginning to populate database (%s)\n", getTimeString())
	err = createUsers(connection)
	if err != nil {
		panic(err)
	}

	repositoryIds, err := createRepositories(connection)
	if err != nil {
		panic(err)
	}

	var _ = repositoryIds
}

func getTimeString() string {
	currentTime := time.Now()
	return fmt.Sprintf("%d:%02d:%d.%d", currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond())
}

func createUsers(connection *resources.Connection) error {
	userPasswordHash := []byte("password-hash")
	userPasswordSalt := []byte("password-salt")

	_, err := connection.Users.Create.Create("admin@koalitycode.com", "Mister", "Admin", userPasswordHash, userPasswordSalt, true)
	if err != nil {
		return err
	}

	_, err = connection.Users.Create.Create("jchu@koalitycode.com", "Jonathan", "Chu", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		return err
	}

	_, err = connection.Users.Create.Create("jpotter@koalitycode.com", "Jordan", "Potter", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		return err
	}

	_, err = connection.Users.Create.Create("bbland@koalitycode.com", "Brian", "Bland", userPasswordHash, userPasswordSalt, false)
	if err != nil {
		return err
	}

	return nil
}

func createRepositories(connection *resources.Connection) ([]uint64, error) {
	repositoryIds := make([]uint64, numRepositories)

	for index := 0; index < numRepositories; index++ {
		repositoryName := fmt.Sprintf("Repository %d", index)
		repositoryLocalUri := fmt.Sprintf("git@test-data.koalitycode.com:koality-%d.git", index)
		repositoryRemoteUri := fmt.Sprintf("git@github.com:KoalityCode/koality-%d.git", index)
		repositoryId, err := connection.Repositories.Create.Create(repositoryName, "git", repositoryLocalUri, repositoryRemoteUri)
		if err != nil {
			return nil, err
		}
		repositoryIds[index] = repositoryId
	}

	return repositoryIds, nil
}
