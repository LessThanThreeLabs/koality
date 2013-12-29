package database

import (
	"fmt"
	"testing"
)

func TestMaid(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		test.Fatal(err)
	}

	createAdditionalVerifications := func(numVerificationsToCreatePerRepository int) {
		errorChannel := make(chan error)
		for _, repository := range repositories {
			go func(repositoryId uint64) {
				err := createVerifications(connection, repositoryId, numVerificationsToCreatePerRepository)
				errorChannel <- err
			}(repository.Id)
		}

		for index := 0; index < len(repositories); index++ {
			err := <-errorChannel
			if err != nil {
				test.Fatal(err)
			}
		}
	}

	createAdditionalVerifications(10)

	numVerificationsToRetain := uint64(10)
	Clean(connection, numVerificationsToRetain)

	fmt.Println("...need to confirm that only first ", numVerificationsToRetain, " have console lines")
}
