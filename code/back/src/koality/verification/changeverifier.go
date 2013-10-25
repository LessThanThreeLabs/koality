package verification

import (
	"fmt"
	"koality/vm"
	"sync"
)

type ChangeVerifier struct {
	vm.VirtualMachinePool
}

type VerificationConfig struct {
	numMachines     int
	setupCommands   []command
	compileCommands []command
	factoryCommands []command
	testCommands    []command
}

type BuildVerifier struct {
	virtualMachine vm.VirtualMachine
	resultsChan    chan result
}

type result struct {
	stageType string
	passed    bool
}

type command struct{}

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

	newBuildVerifiersChannel := make(chan *BuildVerifier, verificationConfig.numMachines)

	launchNewMachineToRunTestsAndStuff := func() {
		buildVerifier := NewBuildVerifier(changeVerifier.VirtualMachinePool.Get())
		defer close(buildVerifier.resultsChan)

		newBuildVerifiersChannel <- buildVerifier

		// Populate pre-test commands
		preTestCommandsChan := make(chan command, len(verificationConfig.setupCommands)+len(verificationConfig.compileCommands))
		for _, setupCommand := range verificationConfig.setupCommands {
			preTestCommandsChan <- setupCommand
		}
		for _, compileCommand := range verificationConfig.compileCommands {
			preTestCommandsChan <- compileCommand
		}
		close(preTestCommandsChan)

		shouldContinue, err := buildVerifier.runPreTestCommands(preTestCommandsChan)
		if err != nil {
			panic(err)
		}
		if !shouldContinue {
			return
		}

		err = buildVerifier.runFactoryCommands(factoryCommandsChan, testCommandsChan, factoriesRun)
		if err != nil {
			panic(err)
		}
		factoriesRun.Wait()

		err = buildVerifier.runTestCommands(testCommandsChan)
		if err != nil {
			panic(err)
		}
	}
	for machineNum := 0; machineNum < verificationConfig.numMachines; machineNum++ {
		go launchNewMachineToRunTestsAndStuff() // Best name
	}

	resultsChan := changeVerifier.combineResults(newBuildVerifiersChannel)

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

func (changeVerifier *ChangeVerifier) combineResults(newBuildVerifiersChan <-chan *BuildVerifier) <-chan result {
	resultsChan := make(chan result)
	go func(newBuildVerifiersChan <-chan *BuildVerifier) {
		combinedResults := make(chan result)
		buildVerifiers := make([]*BuildVerifier, 0, cap(newBuildVerifiersChan))
		buildVerifierDoneChan := make(chan error, cap(newBuildVerifiersChan))
		buildVerifiersDoneCounter := 0

		handleNewBuildVerifier := func(buildVerifier *BuildVerifier) {
			for {
				result, ok := <-buildVerifier.resultsChan
				if !ok {
					buildVerifierDoneChan <- nil
					return
				}
				combinedResults <- result
			}
		}

		for {
			select {
			case buildVerifier, ok := <-newBuildVerifiersChan:
				if !ok {
					panic("new build verifiers channel closed")
				}
				buildVerifiers = append(buildVerifiers, buildVerifier)
				go handleNewBuildVerifier(buildVerifier)

			case err, ok := <-buildVerifierDoneChan:
				if !ok {
					panic("build verifier done channel closed")
				}
				if err != nil {
					panic(err)
				}
				buildVerifiersDoneCounter++
				if buildVerifiersDoneCounter >= len(buildVerifiers) {
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
		for _, buildVerifier := range buildVerifiers {
			for {
				_, ok := <-buildVerifier.resultsChan
				if !ok {
					break
				}
			}
		}
	}(newBuildVerifiersChan)
	return resultsChan
}

func NewBuildVerifier(virtualMachine vm.VirtualMachine) *BuildVerifier {
	return &BuildVerifier{
		virtualMachine: virtualMachine,
		resultsChan:    make(chan result),
	}
}

func (buildVerifier *BuildVerifier) runPreTestCommands(commandChan <-chan command) (bool, error) {
	for {
		command, ok := <-commandChan
		if !ok {
			return true, nil
		}
		success, err := buildVerifier.runCommand(command)
		if err != nil {
			return false, err
		}
		if !success {
			return false, nil
		}
		// Otherwise continue
	}
}

func (buildVerifier *BuildVerifier) runFactoryCommands(inCommandChan <-chan command, outCommandChan chan<- command, doneTracker *sync.WaitGroup) error {
	for {
		c, ok := <-inCommandChan
		if !ok {
			return nil
		}
		_, err := buildVerifier.runCommand(c)
		defer doneTracker.Done()
		if err != nil {
			return err
		}
		outCommandChan <- command{} // Yes, this makes NO sense
	}
}

func (buildVerifier *BuildVerifier) runTestCommands(commandChan <-chan command) error {
	for {
		command, ok := <-commandChan
		if !ok {
			return nil
		}
		_, err := buildVerifier.runCommand(command)
		if err != nil {
			return err
		}
		// Otherwise continue
	}
}

func (buildVerifier *BuildVerifier) runCommand(c command) (bool, error) {
	return false, nil
}
