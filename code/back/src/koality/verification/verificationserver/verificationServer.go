package verificationserver

import (
	"fmt"
	"io/ioutil"
	"koality/resources"
	"koality/shell"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/verification/stagerunner"
	"koality/vm"
	"os/user"
	"time"
)

type VerificationServer struct {
	ResourcesConnection *resources.Connection
	VirtualMachinePools map[uint64]vm.VirtualMachinePool
}

func (verificationServer *VerificationServer) RunVerification(currentVerification *resources.Verification) (bool, error) {
	if currentVerification == nil {
		panic("Cannot run a nil verification")
	}
	virtualMachinePool, ok := verificationServer.VirtualMachinePools[currentVerification.RepositoryId]
	if !ok {
		verificationServer.failVerification(currentVerification)
		return false, fmt.Errorf("No virtual machine pool found for repository: %d", currentVerification.RepositoryId)
	}
	verificationConfig, err := verificationServer.getVerificationConfig(currentVerification)
	if err != nil {
		verificationServer.failVerification(currentVerification)
		return false, err
	}

	err = verificationServer.createStages(currentVerification, verificationConfig.Sections)
	if err != nil {
		verificationServer.failVerification(currentVerification)
		return false, err
	}

	numNodes := verificationConfig.Params.Nodes
	if numNodes <= 0 {
		numNodes = 4 // TEMPORARY
	}

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)

	runStages := func(virtualMachine vm.VirtualMachine) {
		defer virtualMachinePool.Free()
		defer virtualMachine.Terminate()

		stageRunner := stagerunner.New(verificationServer.ResourcesConnection, virtualMachine, currentVerification)
		defer close(stageRunner.ResultsChan)

		newStageRunnersChan <- stageRunner

		err := stageRunner.RunStages(verificationConfig.Sections)
		if err != nil {
			panic(err)
		}
	}

	newMachinesChan := virtualMachinePool.GetN(int(numNodes))
	go func(newMachinesChan <-chan vm.VirtualMachine) {
		for newMachine := range newMachinesChan {
			go runStages(newMachine)
		}
	}(newMachinesChan)

	resultsChan := verificationServer.combineResults(currentVerification, newStageRunnersChan)
	receivedResult := make(map[string]bool)

	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			verificationPassed := currentVerification.VerificationStatus != "cancelled" && currentVerification.VerificationStatus != "failed"
			if verificationPassed {
				err := verificationServer.passVerification(currentVerification)
				if err != nil {
					return false, err
				}
			}
			return verificationPassed, nil
		}
		if result.Passed == false {
			switch result.FailSectionOn {
			case section.FailOnNever:
				// Do nothing
			case section.FailOnAny:
				if currentVerification.VerificationStatus != "cancelled" && currentVerification.VerificationStatus != "failed" {
					err := verificationServer.failVerification(currentVerification)
					if err != nil {
						return false, err
					}
				}
			case section.FailOnFirst:
				if !receivedResult[result.Section] {
					if currentVerification.VerificationStatus != "cancelled" && currentVerification.VerificationStatus != "failed" {
						err := verificationServer.failVerification(currentVerification)
						if err != nil {
							return false, err
						}
					}
				}
			}
		}
		receivedResult[result.Section] = true
	}
}

func (verificationServer *VerificationServer) failVerification(verification *resources.Verification) error {
	(*verification).VerificationStatus = "failed"
	fmt.Printf("verification %d %sFAILED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := verificationServer.ResourcesConnection.Verifications.Update.SetStatus(verification.Id, "failed")
	if err != nil {
		return err
	}

	err = verificationServer.ResourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

func (verificationServer *VerificationServer) passVerification(verification *resources.Verification) error {
	(*verification).VerificationStatus = "passed"
	fmt.Printf("verification %d %sPASSED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := verificationServer.ResourcesConnection.Verifications.Update.SetStatus(verification.Id, "passed")
	if err != nil {
		return err
	}

	err = verificationServer.ResourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

// TODO (bbland): make this not bogus
func (verificationServer *VerificationServer) getVerificationConfig(currentVerification *resources.Verification) (config.VerificationConfig, error) {
	var emptyConfig config.VerificationConfig

	usr, err := user.Current()
	if err != nil {
		return emptyConfig, err
	}

	example_yaml, err := ioutil.ReadFile(fmt.Sprintf("%s/code/back/src/koality/verification/config/example_koality.yml", usr.HomeDir))
	if err != nil {
		return emptyConfig, err
	}

	return config.FromYaml(string(example_yaml))
}

func (verificationServer *VerificationServer) createStages(currentVerification *resources.Verification, sections []section.Section) error {
	for sectionNumber, section := range sections {
		var err error

		stageNumber := 0

		factoryCommands := section.FactoryCommands(true)

		for command, err := factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
			_, err := verificationServer.ResourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
		if err != nil && err != commandgroup.NoMoreCommands {
			return err
		}

		commands := section.Commands(true)
		for command, err := commands.Next(); err == nil; command, err = commands.Next() {
			_, err := verificationServer.ResourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
		if err != nil && err != commandgroup.NoMoreCommands {
			return err
		}
	}
	return nil
}

func (verificationServer *VerificationServer) combineResults(currentVerification *resources.Verification, newStageRunnersChan <-chan *stagerunner.StageRunner) <-chan verification.SectionResult {
	resultsChan := make(chan verification.SectionResult)
	go func(newStageRunnersChan <-chan *stagerunner.StageRunner) {
		isStarted := false

		combinedResults := make(chan verification.SectionResult)
		stageRunners := make([]*stagerunner.StageRunner, 0, cap(newStageRunnersChan))
		stageRunnerDoneChan := make(chan error, cap(newStageRunnersChan))
		stageRunnersDoneCounter := 0

		handleNewStageRunner := func(stageRunner *stagerunner.StageRunner) {
			for {
				result, ok := <-stageRunner.ResultsChan
				if !ok {
					stageRunnerDoneChan <- nil
					return
				}
				combinedResults <- result
			}
		}

		// TODO (bbland): make this not so ugly
	gatherResults:
		for {
			select {
			case stageRunner, ok := <-newStageRunnersChan:
				if !ok {
					panic("new stage Runners channel closed")
				}
				if !isStarted {
					isStarted = true
					verificationServer.ResourcesConnection.Verifications.Update.SetStartTime(currentVerification.Id, time.Now())
				}
				stageRunners = append(stageRunners, stageRunner)
				go handleNewStageRunner(stageRunner)

			case err, ok := <-stageRunnerDoneChan:
				if !ok {
					panic("stage runner done channel closed")
				}
				if err != nil {
					panic(err)
				}
				stageRunnersDoneCounter++
				if stageRunnersDoneCounter >= len(stageRunners) {
					break gatherResults
				}

			case result, ok := <-combinedResults:
				if !ok {
					panic("combined results channel closed")
				}
				resultsChan <- result
			}
		}
		// Drain the remaining results that matter
		drainRemaining := func() {
			for {
				select {
				case result := <-combinedResults:
					resultsChan <- result
				default:
					close(resultsChan)
					return
				}
			}
		}
		drainRemaining()
		// Drain any extra possible results that we don't care about
		for _, stageRunner := range stageRunners {
			for _ = range stageRunner.ResultsChan {
				fmt.Println("Got an extra result")
			}
		}
	}(newStageRunnersChan)
	return resultsChan
}
