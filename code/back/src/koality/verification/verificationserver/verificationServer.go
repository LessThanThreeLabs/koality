package verificationserver

import (
	"fmt"
	"koality/resources"
	"koality/shell"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/verification/stagerunner"
	"koality/vm"
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

	resultsChan := verificationServer.combineResults(newStageRunnersChan)
	receivedResult := make(map[string]bool)

	// TODO (bbland): do this when the first stageRunner begins
	verificationServer.ResourcesConnection.Verifications.Update.SetStartTime(currentVerification.Id, time.Now())

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
	verificationServer.ResourcesConnection.Verifications.Update.SetStatus(verification.Id, "failed")
	return nil
}

func (verificationServer *VerificationServer) passVerification(verification *resources.Verification) error {
	(*verification).VerificationStatus = "passed"
	fmt.Printf("verification %d %sPASSED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	verificationServer.ResourcesConnection.Verifications.Update.SetStatus(verification.Id, "passed")
	return nil
}

// TODO (bbland): make this not bogus
func (verificationServer *VerificationServer) getVerificationConfig(currentVerification *resources.Verification) (config.VerificationConfig, error) {
	config, err := config.FromYaml("parameters:\n  languages:\n    python: asdf")
	config.Sections = append(config.Sections, section.New(
		"test",
		section.RunOnSplit,
		section.FailOnFirst,
		true,
		commandgroup.New([]verification.Command{verification.NewShellCommand("hi", "echo hi"), verification.NewShellCommand("bye", "echo bye")}),
		commandgroup.New([]verification.Command{}),
		[]string{},
	))
	return config, err
}

func (verificationServer *VerificationServer) combineResults(newStageRunnersChan <-chan *stagerunner.StageRunner) <-chan verification.SectionResult {
	resultsChan := make(chan verification.SectionResult)
	go func(newStageRunnersChan <-chan *stagerunner.StageRunner) {
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
