package stagerunner

import (
	"fmt"
	"koality/resources/database"
	"koality/shell"
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/vm/localmachine"
	"testing"
)

func TestSimplePassingStages(test *testing.T) {
	database.PopulateDatabase()

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	if err != nil {
		test.Fatal(err)
	}

	currentVerification, err := resourcesConnection.Verifications.Create.Create(repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	stageRunner := New(resourcesConnection, virtualMachine, currentVerification)

	commands := []verification.Command{verification.NewShellCommand("pass", shell.Command("true"))}
	for index, command := range commands {
		_, err := resourcesConnection.Stages.Create.Create(currentVerification.Id, 0, command.Name(), uint64(index))
		if err != nil {
			test.Fatal(err)
		}
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range stageRunner.ResultsChan {
			if !result.Passed {
				test.Log(fmt.Sprintf("Failed section %s", result.Section))
				test.Fail()
			}
		}

		doneChan <- err
	}(doneChan)

	err = stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	if err != nil {
		test.Fatal(err)
	}

	close(stageRunner.ResultsChan)
}

func TestSimpleFailingStages(test *testing.T) {
	database.PopulateDatabase()

	resourcesConnection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}

	virtualMachine := localmachine.New()
	defer virtualMachine.Terminate()

	repository, err := resourcesConnection.Repositories.Create.Create("repositoryName", "git", "localUri", "remote@Uri")
	if err != nil {
		test.Fatal(err)
	}

	currentVerification, err := resourcesConnection.Verifications.Create.Create(repository.Id, "1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890",
		"headMessage", "headUsername", "head@Ema.il", "mergeTarget", "a@b.com")
	if err != nil {
		test.Fatal(err)
	}

	stageRunner := New(resourcesConnection, virtualMachine, currentVerification)

	commands := []verification.Command{verification.NewShellCommand("fail", shell.Command("false"))}
	for index, command := range commands {
		_, err := resourcesConnection.Stages.Create.Create(currentVerification.Id, 0, command.Name(), uint64(index))
		if err != nil {
			test.Fatal(err)
		}
	}

	commandGroup := commandgroup.New(commands)

	testSection := section.New("test", false, section.RunOnAll, section.FailOnAny, false, nil, commandGroup, nil)

	doneChan := make(chan error)

	go func(doneChan chan error) {
		for result := range stageRunner.ResultsChan {
			if result.Passed {
				test.Log(fmt.Sprintf("Passed section %s", result.Section))
				test.Fail()
			}
		}

		doneChan <- err
	}(doneChan)

	err = stageRunner.RunStages([]section.Section{testSection}, nil, nil)
	if err != nil {
		test.Fatal(err)
	}

	close(stageRunner.ResultsChan)
}
