package database

import (
	"fmt"
	"koality/resources"
	"testing"
)

func TestMaid(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	err = addVerifications(connection, 10)
	if err != nil {
		test.Fatal(err)
	}

	numVerificationsToRetain := uint32(10)
	Clean(connection, numVerificationsToRetain)
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

func checkVerificationsCleaned(connection *resources.Connection, numVerificationsToRetain uint32) error {
	repositories, err := connection.Repositories.Read.GetAll()
	if err != nil {
		return err
	}

	verificationCount := 0
	errorChannel := make(chan error)
	for _, repository := range repositories {
		verifications, err := connection.Verifications.Read.GetTail(repository.Id, 0, 1000000)
		if err != nil {
			return err
		}
		verificationCount += len(verifications)

		for index, verification := range verifications {
			go func(index int, verificationId uint64) {
				containsOutput, err := doesVerificationContainOutput(connection, verificationId)
				if err != nil {
					errorChannel <- err
					return
				}

				if index < int(numVerificationsToRetain) && !containsOutput {
					errorChannel <- fmt.Errorf("Expected output for verification #%d with id: %d\n", index, verificationId)
				} else if index >= int(numVerificationsToRetain) && containsOutput {
					errorChannel <- fmt.Errorf("Expected no output for verification #%d with id: %d\n", index, verificationId)
				} else {
					errorChannel <- nil
				}
			}(index, verification.Id)
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
	stages, err := connection.Stages.Read.GetAll(verificationId)
	if err != nil {
		return false, err
	}

	for _, stage := range stages {
		for _, stageRun := range stage.Runs {
			consoleLines, err := connection.Stages.Read.GetAllConsoleLines(stageRun.Id)
			if err != nil {
				return false, err
			}

			xunitResults, err := connection.Stages.Read.GetAllXunitResults(stageRun.Id)
			if err != nil {
				return false, err
			}

			if len(consoleLines) != 0 || len(xunitResults) != 0 {
				return true, nil
			}
		}
	}

	return false, nil
}
