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

	buildCount := 0
	errorChannel := make(chan error)
	for _, repository := range repositories {
		builds, err := connection.Verifications.Read.GetTail(repository.Id, 0, 1000000)
		if err != nil {
			return err
		}
		buildCount += len(builds)

		for index, build := range builds {
			go func(index int, buildId uint64) {
				containsOutput, err := doesVerificationContainOutput(connection, buildId)
				if err != nil {
					errorChannel <- err
					return
				}

				if index < int(numVerificationsToRetain) && !containsOutput {
					errorChannel <- fmt.Errorf("Expected output for build #%d with id: %d\n", index, buildId)
				} else if index >= int(numVerificationsToRetain) && containsOutput {
					errorChannel <- fmt.Errorf("Expected no output for build #%d with id: %d\n", index, buildId)
				} else {
					errorChannel <- nil
				}
			}(index, build.Id)
		}
	}

	for index := 0; index < buildCount; index++ {
		err := <-errorChannel
		if err != nil {
			return err
		}
	}
	return nil
}

func doesVerificationContainOutput(connection *resources.Connection, buildId uint64) (bool, error) {
	stages, err := connection.Stages.Read.GetAll(buildId)
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
