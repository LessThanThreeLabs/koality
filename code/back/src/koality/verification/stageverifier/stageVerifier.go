package stageverifier

import (
	"koality/verification"
	"koality/verification/config/commandgroup"
	"koality/vm"
)

type StageVerifier struct {
	virtualMachine vm.VirtualMachine
	ResultsChan    chan verification.Result
}

func New(virtualMachine vm.VirtualMachine) *StageVerifier {
	return &StageVerifier{
		virtualMachine: virtualMachine,
		ResultsChan:    make(chan verification.Result),
	}
}

func (stageVerifier *StageVerifier) RunChangeStages(setupCommands commandgroup.CommandGroup, compileCommands commandgroup.CommandGroup, factoryCommands commandgroup.CommandGroup, testCommands commandgroup.AppendableCommandGroup) error {
	shouldContinue, err := stageVerifier.runSetupCommands(setupCommands)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	shouldContinue, err = stageVerifier.runCompileCommands(setupCommands)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	shouldContinue, err = stageVerifier.runFactoryCommands(factoryCommands, testCommands)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	factoryCommands.Wait()
	// closeTests

	_, err = stageVerifier.runTestCommands(testCommands)
	return err
}

func (stageVerifier *StageVerifier) runSetupCommands(setupCommands commandgroup.CommandGroup) (bool, error) {
	var err error
	var setupCommand *verification.Command
	for setupCommand, err = setupCommands.Next(); setupCommand != nil && err == nil; setupCommand, err = setupCommands.Next() {
		success, err := stageVerifier.runCommand(setupCommand)
		setupCommands.Done()
		if err != nil {
			return false, err
		}
		if !success {
			return false, nil
		}
	}
	if err == commandgroup.NoMoreCommands {
		return true, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runCompileCommands(compileCommands commandgroup.CommandGroup) (bool, error) {
	var err error
	for compileCommand, err := compileCommands.Next(); compileCommand != nil && err == nil; compileCommand, err = compileCommands.Next() {
		success, err := stageVerifier.runCommand(compileCommand)
		compileCommands.Done()
		if err != nil {
			return false, err
		}
		if !success {
			return false, nil
		}
	}
	if err == commandgroup.NoMoreCommands {
		return true, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runFactoryCommands(factoryCommands commandgroup.CommandGroup, testCommands commandgroup.AppendableCommandGroup) (bool, error) {
	var err error
	for factoryCommand, err := factoryCommands.Next(); factoryCommand != nil && err == nil; factoryCommand, err = factoryCommands.Next() {
		success, err := stageVerifier.runCommand(factoryCommand)
		factoryCommands.Done()
		if err != nil {
			return false, err
		}
		if !success {
			return false, nil
		}
		// outCommandChan <- command{} // Yes, this makes NO sense
	}
	if err == commandgroup.NoMoreCommands {
		return true, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runTestCommands(testCommands commandgroup.CommandGroup) (bool, error) {
	testsSuccess := true
	var err error
	for testCommand, err := testCommands.Next(); testCommand != nil && err == nil; testCommand, err = testCommands.Next() {
		success, err := stageVerifier.runCommand(testCommand)
		testCommands.Done()
		testsSuccess = testsSuccess && success
		if err != nil {
			return false, err
		}
	}
	if err == commandgroup.NoMoreCommands {
		return testsSuccess, nil
	}
	return false, err
}

func (stageVerifier *StageVerifier) runCommand(command *verification.Command) (bool, error) {
	return false, nil
}
