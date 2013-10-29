package verification

import (
	"koality/vm"
	"sync"
)

type StageVerifier struct {
	virtualMachine vm.VirtualMachine
	resultsChan    chan result
}

func NewStageVerifier(virtualMachine vm.VirtualMachine) *StageVerifier {
	return &StageVerifier{
		virtualMachine: virtualMachine,
		resultsChan:    make(chan result),
	}
}

func (stageVerifier *StageVerifier) RunChangeStages(preTestCommandsChan, factoryCommandsChan <-chan command, testCommandsChan chan command, factoryTracker *sync.WaitGroup) error {
	shouldContinue, err := stageVerifier.runPreTestCommands(preTestCommandsChan)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	shouldContinue, err = stageVerifier.runFactoryCommands(factoryCommandsChan, testCommandsChan, factoryTracker)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	factoryTracker.Wait()
	// closeTests

	_, err = stageVerifier.runTestCommands(testCommandsChan)
	return err
}

func (stageVerifier *StageVerifier) runPreTestCommands(commandChan <-chan command) (bool, error) {
	for {
		c, ok := <-commandChan
		if !ok {
			return true, nil
		}
		success, err := stageVerifier.runCommand(c)
		if err != nil {
			return false, err
		}
		if !success {
			return false, nil
		}
		// Otherwise continue
	}
}

func (stageVerifier *StageVerifier) runFactoryCommands(inCommandChan <-chan command, outCommandChan chan<- command, doneTracker *sync.WaitGroup) (bool, error) {
	for {
		c, ok := <-inCommandChan
		if !ok {
			return true, nil
		}
		success, err := stageVerifier.runCommand(c)
		defer doneTracker.Done()
		if err != nil {
			return false, err
		}
		if !success {
			return false, nil
		}
		// outCommandChan <- command{} // Yes, this makes NO sense
	}
}

func (stageVerifier *StageVerifier) runTestCommands(commandChan <-chan command) (bool, error) {
	testsSuccess := true
	for {
		c, ok := <-commandChan
		if !ok {
			return testsSuccess, nil
		}
		success, err := stageVerifier.runCommand(c)
		testsSuccess = testsSuccess && success
		if err != nil {
			return false, err
		}
		// Otherwise continue
	}
}

func (stageVerifier *StageVerifier) runCommand(c command) (bool, error) {
	return false, nil
}
