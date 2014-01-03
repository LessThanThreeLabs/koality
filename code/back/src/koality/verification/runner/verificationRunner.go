package runner

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

type VerificationRunner struct {
	ResourcesConnection *resources.Connection
	VirtualMachinePools map[uint64]vm.VirtualMachinePool
}

func (verificationRunner *VerificationRunner) RunVerification(currentVerification *resources.Verification) (bool, error) {
	if currentVerification == nil {
		panic("Cannot run a nil verification")
	}
	verificationConfig, err := verificationRunner.getVerificationConfig(currentVerification)
	if err != nil {
		verificationRunner.failVerification(currentVerification)
		return false, err
	}

	virtualMachinePool, ok := verificationRunner.VirtualMachinePools[verificationConfig.Params.PoolId]
	if !ok {
		verificationRunner.failVerification(currentVerification)
		return false, fmt.Errorf("No virtual machine pool found for repository: %d", currentVerification.RepositoryId)
	}

	err = verificationRunner.createStages(currentVerification, verificationConfig.Sections, verificationConfig.FinalSections)
	if err != nil {
		verificationRunner.failVerification(currentVerification)
		return false, err
	}

	numNodes := verificationConfig.Params.Nodes
	if numNodes == 0 {
		numNodes = 1 // TEMPORARY
	}

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)

	runStages := func(virtualMachine vm.VirtualMachine) {
		defer virtualMachinePool.Free()
		defer virtualMachine.Terminate()

		stageRunner := stagerunner.New(verificationRunner.ResourcesConnection, virtualMachine, currentVerification)
		defer close(stageRunner.ResultsChan)

		newStageRunnersChan <- stageRunner

		err := stageRunner.RunStages(verificationConfig.Sections, verificationConfig.FinalSections)
		if err != nil {
			panic(err)
		}
	}

	newMachinesChan := virtualMachinePool.GetN(numNodes)
	go func(newMachinesChan <-chan vm.VirtualMachine) {
		for newMachine := range newMachinesChan {
			go runStages(newMachine)
		}
	}(newMachinesChan)

	resultsChan := verificationRunner.combineResults(currentVerification, newStageRunnersChan)
	receivedResult := make(map[string]bool)

	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			verificationPassed := currentVerification.Status != "cancelled" && currentVerification.Status != "failed"
			if verificationPassed {
				err := verificationRunner.passVerification(currentVerification)
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
				if currentVerification.Status != "cancelled" && currentVerification.Status != "failed" {
					err := verificationRunner.failVerification(currentVerification)
					if err != nil {
						return false, err
					}
				}
			case section.FailOnFirst:
				if !receivedResult[result.Section] {
					if currentVerification.Status != "cancelled" && currentVerification.Status != "failed" {
						err := verificationRunner.failVerification(currentVerification)
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

func (verificationRunner *VerificationRunner) failVerification(verification *resources.Verification) error {
	(*verification).Status = "failed"
	fmt.Printf("verification %d %sFAILED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := verificationRunner.ResourcesConnection.Verifications.Update.SetStatus(verification.Id, "failed")
	if err != nil {
		return err
	}

	err = verificationRunner.ResourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

func (verificationRunner *VerificationRunner) passVerification(verification *resources.Verification) error {
	(*verification).Status = "passed"
	fmt.Printf("verification %d %sPASSED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := verificationRunner.ResourcesConnection.Verifications.Update.SetStatus(verification.Id, "passed")
	if err != nil {
		return err
	}

	err = verificationRunner.ResourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

// TODO (bbland): make this not bogus
func (verificationRunner *VerificationRunner) getVerificationConfig(currentVerification *resources.Verification) (config.VerificationConfig, error) {
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

func (verificationRunner *VerificationRunner) createStages(currentVerification *resources.Verification, sections, finalSections []section.Section) error {
	for sectionNumber, section := range append(sections, finalSections...) {
		var err error

		stageNumber := 0

		factoryCommands := section.FactoryCommands(true)

		for command, err := factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
			_, err := verificationRunner.ResourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
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
			_, err := verificationRunner.ResourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
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

func (verificationRunner *VerificationRunner) combineResults(currentVerification *resources.Verification, newStageRunnersChan <-chan *stagerunner.StageRunner) <-chan verification.SectionResult {
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
					verificationRunner.ResourcesConnection.Verifications.Update.SetStartTime(currentVerification.Id, time.Now())
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
