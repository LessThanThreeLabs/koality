package changeverifier

import (
	"fmt"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/stageverifier"
	"koality/vm"
)

type ChangeVerifier struct {
	vm.VirtualMachinePool
}

func (changeVerifier *ChangeVerifier) VerifyChange(changeId int) (bool, error) {
	verificationConfig, err := changeVerifier.getVerificationConfig(changeId)
	if err != nil {
		changeVerifier.failChange(changeId)
		return false, err
	}

	factoryCommands := commandgroup.New(verificationConfig.FactoryCommands)
	testCommands := commandgroup.New(verificationConfig.TestCommands)

	newStageVerifiersChan := make(chan *stageverifier.StageVerifier, verificationConfig.NumMachines)

	launchNewMachineToRunTestsAndStuff := func() {
		stageVerifier := stageverifier.New(changeVerifier.VirtualMachinePool.Get())
		defer close(stageVerifier.ResultsChan)

		newStageVerifiersChan <- stageVerifier

		// Populate pre-test commands
		setupCommands := commandgroup.New(verificationConfig.SetupCommands)
		compileCommands := commandgroup.New(verificationConfig.CompileCommands)

		err := stageVerifier.RunChangeStages(setupCommands, compileCommands, factoryCommands, testCommands)
		if err != nil {
			panic(err)
		}
	}
	for machineNum := 0; machineNum < verificationConfig.NumMachines; machineNum++ {
		go launchNewMachineToRunTestsAndStuff() // Best name
	}

	resultsChan := changeVerifier.combineResults(newStageVerifiersChan)

	testsStarted := false
	failed := false
	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			if !failed {
				err := changeVerifier.passChange(changeId)
				if err != nil {
					return false, err
				}
			}
			return !failed, nil
		}
		if result.Passed == false {
			if result.StageType == "setup" || result.StageType == "compile" {
				if !testsStarted {
					failed = true
					err := changeVerifier.failChange(changeId)
					if err != nil {
						return false, err
					}
				}
				// changeDone <- nil
			} else if result.StageType == "test" {
				failed = true
				err := changeVerifier.failChange(changeId)
				if err != nil {
					return false, err
				}
			} else {
				panic(fmt.Sprintf("Unexpected result %#v", result))
			}
		}
		if result.StageType == "test" {
			testsStarted = true // This is the WRONG place for this
		}
	}
}

func (changeVerifier *ChangeVerifier) failChange(changeId int) error {
	panic(fmt.Sprintf("change %d failed", changeId))
}

func (changeVerifier *ChangeVerifier) passChange(changeId int) error {
	panic(fmt.Sprintf("change %d passed", changeId))
}

func (changeVerifier *ChangeVerifier) getVerificationConfig(changeId int) (config.VerificationConfig, error) {
	panic(fmt.Sprintf("not implemented"))
}

func (changeVerifier *ChangeVerifier) combineResults(newStageVerifiersChan <-chan *stageverifier.StageVerifier) <-chan verification.Result {
	resultsChan := make(chan verification.Result)
	go func(newStageVerifiersChan <-chan *stageverifier.StageVerifier) {
		combinedResults := make(chan verification.Result)
		stageVerifiers := make([]*stageverifier.StageVerifier, 0, cap(newStageVerifiersChan))
		stageVerifierDoneChan := make(chan error, cap(newStageVerifiersChan))
		stageVerifiersDoneCounter := 0

		handleNewStageVerifier := func(stageVerifier *stageverifier.StageVerifier) {
			for {
				result, ok := <-stageVerifier.ResultsChan
				if !ok {
					stageVerifierDoneChan <- nil
					return
				}
				combinedResults <- result
			}
		}

		for {
			select {
			case stageVerifier, ok := <-newStageVerifiersChan:
				if !ok {
					panic("new stage verifiers channel closed")
				}
				stageVerifiers = append(stageVerifiers, stageVerifier)
				go handleNewStageVerifier(stageVerifier)

			case err, ok := <-stageVerifierDoneChan:
				if !ok {
					panic("stage verifier done channel closed")
				}
				if err != nil {
					panic(err)
				}
				stageVerifiersDoneCounter++
				if stageVerifiersDoneCounter >= len(stageVerifiers) {
					fmt.Println("shit's done")
					break
				}

			case result, ok := <-combinedResults:
				if !ok {
					panic("combined results channel closed")
				}
				resultsChan <- result
			}
		}
		// Drain the remaining results that matter
		for {
			select {
			case result := <-combinedResults:
				resultsChan <- result
			default:
				fmt.Println("results drained")
				close(resultsChan)
				break
			}
		}
		// Drain any extra possible results that we don't care about
		for _, stageVerifier := range stageVerifiers {
			for {
				_, ok := <-stageVerifier.ResultsChan
				if !ok {
					break
				}
			}
		}
	}(newStageVerifiersChan)
	return resultsChan
}