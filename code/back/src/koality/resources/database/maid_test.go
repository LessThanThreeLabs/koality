package database

import (
	"fmt"
	"koality/resources"
	"testing"
)

func TestMaid(test *testing.T) {
	PopulateDatabase()

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	err = addVerifications(connection, 10)
	if err != nil {
		test.Fatal(err)
	}

	numVerificationsToRetain := uint64(10)
	Clean(connection, numVerificationsToRetain)

	fmt.Println("...need to confirm that only first ", numVerificationsToRetain, " have console lines")
	err = checkVerificationsCleaned(connection, numVerificationsToRetain)
	if err != nil {
		test.Fatal(err)
	}
}

func addVerifications(connection *resources.Connection, numVerificationsToCreatePerRepository int) error {
	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		return err
	}

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
			return err
		}
	}
	return nil
}

func checkVerificationsCleaned(connection *resources.Connection, numVerificationsToRetain uint64) error {
	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		return err
	}

	verificationCount := 0
	errorChannel := make(chan error)
	for _, repository := range repositories {
		verifications, err := connection.Verifications.Read.GetTail(repository.Id, 0, 100000)
		if err != nil {
			return err
		}
		verificationCount += len(verifications)

		for index, verification := range verifications {
			go func(verificationId uint64) {
				containsOutput, err := doesVerificationContainOutput(connection, verificationId)
				if err != nil {
					errorChannel <- err
					return
				}

				if index < int(numVerificationsToRetain) && !containsOutput {
					errorChannel <- fmt.Errorf("Expected output for verification #%d with id: %d\n", index, verificationId)
				} else if index >= int(numVerificationsToRetain) && containsOutput {
					errorChannel <- fmt.Errorf("Expected no output for verification #%d with id: %d\n", index, verificationId)
				}
			}(verification.Id)
		}
	}

	for index := 0; index < verificationCount; index++ {
		err := <-errorChannel
		if err != nil {
			return err
		}
	}
	return nil
}

func doesVerificationContainOutput(connection *resources.Connection, verificationId uint64) (bool, error) {
	return true, nil
}
