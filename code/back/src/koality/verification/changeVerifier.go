package verification

import (
	"fmt"
	"koality/vm"
	"sync"
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

	factoryCommandsChan := make(chan command, len(verificationConfig.factoryCommands))
	for _, factoryCommand := range verificationConfig.factoryCommands {
		factoryCommandsChan <- factoryCommand
	}
	close(factoryCommandsChan)

	factoriesRun := new(sync.WaitGroup)
	factoriesRun.Add(len(verificationConfig.factoryCommands))

	testCommandsChan := make(chan command, len(verificationConfig.testCommands))
	for _, testCommand := range verificationConfig.testCommands {
		testCommandsChan <- testCommand
	}

	newStageVerifiersChan := make(chan *StageVerifier, verificationConfig.numMachines)

	launchNewMachineToRunTestsAndStuff := func() {
		stageVerifier := NewStageVerifier(changeVerifier.VirtualMachinePool.Get())
		defer close(stageVerifier.resultsChan)

		newStageVerifiersChan <- stageVerifier

		// Populate pre-test commands
		preTestCommandsChan := make(chan command, len(verificationConfig.setupCommands)+len(verificationConfig.compileCommands))
		for _, setupCommand := range verificationConfig.setupCommands {
			preTestCommandsChan <- setupCommand
		}
		for _, compileCommand := range verificationConfig.compileCommands {
			preTestCommandsChan <- compileCommand
		}
		close(preTestCommandsChan)

		err := stageVerifier.RunChangeStages(preTestCommandsChan, factoryCommandsChan, testCommandsChan, factoriesRun)
		if err != nil {
			panic(err)
		}
	}
	for machineNum := 0; machineNum < verificationConfig.numMachines; machineNum++ {
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
		if result.passed == false {
			if result.stageType == "setup" || result.stageType == "compile" {
				if !testsStarted {
					failed = true
					err := changeVerifier.failChange(changeId)
					if err != nil {
						return false, err
					}
				}
				// changeDone <- nil
			} else if result.stageType == "test" {
				failed = true
				err := changeVerifier.failChange(changeId)
				if err != nil {
					return false, err
				}
			} else {
				panic(fmt.Sprintf("Unexpected result %#v", result))
			}
		}
		if result.stageType == "test" {
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

func (changeVerifier *ChangeVerifier) getVerificationConfig(changeId int) (VerificationConfig, error) {
	panic(fmt.Sprintf("not implemented"))
}

func (changeVerifier *ChangeVerifier) combineResults(newStageVerifiersChan <-chan *StageVerifier) <-chan result {
	resultsChan := make(chan result)
	go func(newStageVerifiersChan <-chan *StageVerifier) {
		combinedResults := make(chan result)
		stageVerifiers := make([]*StageVerifier, 0, cap(newStageVerifiersChan))
		stageVerifierDoneChan := make(chan error, cap(newStageVerifiersChan))
		stageVerifiersDoneCounter := 0

		handleNewStageVerifier := func(stageVerifier *StageVerifier) {
			for {
				result, ok := <-stageVerifier.resultsChan
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
				_, ok := <-stageVerifier.resultsChan
				if !ok {
					break
				}
			}
		}
	}(newStageVerifiersChan)
	return resultsChan
}
