package changeverifier

import (
	"fmt"
	"koality/shell"
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
	changeStatus := new(verification.ChangeStatus)

	verificationConfig, err := changeVerifier.getVerificationConfig(changeId)
	if err != nil {
		changeVerifier.failChange(changeId)
		return false, err
	}

	factoryCommands := commandgroup.New(verificationConfig.FactoryCommands)
	testCommands := commandgroup.New(verificationConfig.TestCommands)

	newStageVerifiersChan := make(chan *stageverifier.StageVerifier, verificationConfig.NumMachines)

	verifyStages := func(virtualMachine vm.VirtualMachine) {
		defer changeVerifier.VirtualMachinePool.Free()
		defer virtualMachine.Terminate()

		stageVerifier := stageverifier.New(virtualMachine, changeStatus)
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

	newMachinesChan := changeVerifier.VirtualMachinePool.GetN(verificationConfig.NumMachines)
	go func(newMachinesChan <-chan vm.VirtualMachine) {
		for newMachine := range newMachinesChan {
			go verifyStages(newMachine)
		}
	}(newMachinesChan)

	resultsChan := changeVerifier.combineResults(newStageVerifiersChan)

	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			if !changeStatus.Failed && !changeStatus.Cancelled {
				err := changeVerifier.passChange(changeId)
				if err != nil {
					return false, err
				}
			}
			return !changeStatus.Failed, nil
		}
		if result.Passed == false {
			if result.StageType == "setup" || result.StageType == "compile" || result.StageType == "factory" {
				if !factoryCommands.HasStarted() && !changeStatus.Failed {
					changeStatus.Failed = true
					err := changeVerifier.failChange(changeId)
					if err != nil {
						return false, err
					}
				}
			} else if result.StageType == "test" {
				if !changeStatus.Failed {
					changeStatus.Failed = true
					err := changeVerifier.failChange(changeId)
					if err != nil {
						return false, err
					}
				}
			} else {
				panic(fmt.Sprintf("Unexpected result %#v", result))
			}
		}
	}
}

func (changeVerifier *ChangeVerifier) failChange(changeId int) error {
	fmt.Printf("change %d %sFAILED!!!%s\n", changeId, shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	return nil
}

func (changeVerifier *ChangeVerifier) passChange(changeId int) error {
	fmt.Printf("change %d %sPASSED!!!%s\n", changeId, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	return nil
}

// TODO (bbland): make this not bogus
func (changeVerifier *ChangeVerifier) getVerificationConfig(changeId int) (config.VerificationConfig, error) {
	// panic(fmt.Sprintf("not implemented"))
	return config.VerificationConfig{
		NumMachines: 10,
		SetupCommands: []verification.Command{
			verification.ShellCommand{shell.Command("echo hello there")},
			// verification.ShellCommand{shell.Command("exit 1")},
		},
		CompileCommands: []verification.Command{},
		FactoryCommands: []verification.Command{},
		TestCommands: []verification.Command{
			verification.ShellCommand{shell.And(shell.Command(fmt.Sprintf("echo -e %s\n", shell.Quote(fmt.Sprintf("%sthis will fail!%s", shell.AnsiFormat(shell.AnsiFgRed), shell.AnsiFormat(shell.AnsiReset))))), "exit 1")},
			verification.ShellCommand{shell.And(shell.Command(fmt.Sprintf("echo -e %s\n", shell.Quote(fmt.Sprintf("%sthis will pass!%s", shell.AnsiFormat(shell.AnsiFgGreen), shell.AnsiFormat(shell.AnsiReset))))), "echo more echoing lol")},
			verification.ShellCommand{shell.And(shell.Command(fmt.Sprintf("echo -e %s\n", shell.Quote(fmt.Sprintf("%sthis will also fail!%s", shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))))), "exit 1")},
		},
	}, nil
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

		// TODO (bbland): make this not so ugly
	gatherResults:
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
		for _, stageVerifier := range stageVerifiers {
			for _ = range stageVerifier.ResultsChan {
				fmt.Println("Got an extra result")
			}
		}
	}(newStageVerifiersChan)
	return resultsChan
}
