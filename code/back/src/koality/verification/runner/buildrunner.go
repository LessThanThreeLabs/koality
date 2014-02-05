package runner

import (
	"fmt"
	"koality/repositorymanager"
	"koality/resources"
	"koality/shell"
	"koality/util/log"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/verification/stagerunner"
	"koality/vm"
	"koality/vm/poolmanager"
	"runtime"
	"time"
)

type BuildRunner struct {
	resourcesConnection *resources.Connection
	poolManager         *poolmanager.PoolManager
	repositoryManager   repositorymanager.RepositoryManager
}

type BuildData struct {
	*resources.Repository
	config.VerificationConfig
	Pool vm.VirtualMachinePool
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager, repositoryManager repositorymanager.RepositoryManager) *BuildRunner {
	return &BuildRunner{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
		repositoryManager:   repositoryManager,
	}
}

func (buildRunner *BuildRunner) GetBuildData(verification *resources.Verification) (*BuildData, error) {
	if verification == nil {
		return nil, fmt.Errorf("Cannot run a nil verification")
	}
	log.Info("Running verification: %v", verification)
	repository, err := buildRunner.resourcesConnection.Repositories.Read.Get(verification.RepositoryId)
	if err != nil {
		buildRunner.failVerification(verification)
		return nil, err
	}

	verificationConfig, err := buildRunner.getVerificationConfig(verification, repository)
	if err != nil {
		buildRunner.failVerification(verification)
		return nil, err
	}

	virtualMachinePool, err := buildRunner.poolManager.GetPool(verificationConfig.Params.PoolId)
	if err != nil {
		buildRunner.failVerification(verification)
		return nil, err
	}
	return &BuildData{repository, verificationConfig, virtualMachinePool}, nil
}

func (buildRunner *BuildRunner) ProcessResults(currentVerification *resources.Verification,
	newStageRunnersChan chan *stagerunner.StageRunner, buildData *BuildData) (bool, error) {

	resultsChan := buildRunner.combineResults(currentVerification, newStageRunnersChan)
	receivedResult := make(map[string]bool)

	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			if currentVerification.Status == "running" {
				err := buildRunner.passVerification(currentVerification, buildData.Repository)
				if err != nil {
					return false, err
				}
			}
			return currentVerification.Status == "passed", nil
		}
		if !result.Final {
			if !result.Passed {
				switch result.FailSectionOn {
				case section.FailOnNever:
					// Do nothing
				case section.FailOnAny:
					if currentVerification.Status != "cancelled" && currentVerification.Status != "failed" {
						err := buildRunner.failVerification(currentVerification)
						if err != nil {
							return false, err
						}
					}
				case section.FailOnFirst:
					if !receivedResult[result.Section] {
						if currentVerification.Status != "cancelled" && currentVerification.Status != "failed" {
							err := buildRunner.failVerification(currentVerification)
							if err != nil {
								return false, err
							}
						}
					}
				}
			}
			receivedResult[result.Section] = true
			if len(receivedResult) == len(buildData.VerificationConfig.Sections) && currentVerification.Status == "running" {
				err := buildRunner.passVerification(currentVerification, buildData.Repository)
				if err != nil {
					return false, err
				}
			}
		}
	}
}

func (buildRunner *BuildRunner) failVerification(verification *resources.Verification) error {
	(*verification).Status = "failed"
	log.Infof("verification %d %sFAILED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := buildRunner.resourcesConnection.Verifications.Update.SetStatus(verification.Id, "failed")
	if err != nil {
		return err
	}

	err = buildRunner.resourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

func (buildRunner *BuildRunner) passVerification(verification *resources.Verification, repository *resources.Repository) error {
	log.Infof("verification %d %sPASSED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := buildRunner.repositoryManager.MergeChangeset(repository, repositorymanager.GitHiddenRef(verification.Changeset.HeadSha), verification.Changeset.BaseSha, verification.MergeTarget)
	if err == nil {
		(*verification).MergeStatus = "passed"
		buildRunner.resourcesConnection.Verifications.Update.SetMergeStatus(verification.Id, "passed")
	} else {
		(*verification).MergeStatus = "failed"
		buildRunner.resourcesConnection.Verifications.Update.SetMergeStatus(verification.Id, "failed")
	}

	(*verification).Status = "passed"
	err = buildRunner.resourcesConnection.Verifications.Update.SetStatus(verification.Id, "passed")
	if err != nil {
		return err
	}

	err = buildRunner.resourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

func (buildRunner *BuildRunner) getVerificationConfig(currentVerification *resources.Verification, repository *resources.Repository) (config.VerificationConfig, error) {
	var emptyConfig config.VerificationConfig

	configYaml, err := buildRunner.repositoryManager.GetYamlFile(repository, currentVerification.Changeset.HeadSha)
	if err != nil {
		return emptyConfig, err
	}

	verificationConfig, err := config.FromYaml(configYaml, repository.Name)
	if err != nil {
		return emptyConfig, err
	}

	// TODO (bbland): add retry logic
	checkoutCommand, err := buildRunner.repositoryManager.GetCheckoutCommand(repository, currentVerification.Changeset.HeadSha)
	if err != nil {
		return emptyConfig, err
	}

	setupCommands := []verification.Command{verification.NewShellCommand(repository.VcsType, checkoutCommand)}
	setupSection := section.New("setup", false, section.RunOnAll, section.FailOnFirst, false, nil, commandgroup.New(setupCommands), nil)
	verificationConfig.Sections = append([]section.Section{setupSection}, verificationConfig.Sections...)
	return verificationConfig, nil
}

func (buildRunner *BuildRunner) CreateStages(currentVerification *resources.Verification, verificationConfig *config.VerificationConfig) error {
	sections, finalSections := verificationConfig.Sections, verificationConfig.FinalSections
	for sectionNumber, section := range append(sections, finalSections...) {
		var err error

		stageNumber := 0

		factoryCommands := section.FactoryCommands(true)
		for command, err := factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
			_, err := buildRunner.resourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
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
			_, err := buildRunner.resourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
		if err != nil && err != commandgroup.NoMoreCommands {
			return err
		}

		exports := section.Exports()
		if len(exports) > 0 {
			_, err := buildRunner.resourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), section.Name()+".export", uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
	}
	return nil
}

func (buildRunner *BuildRunner) RunStagesOnNewMachines(
	numNodes uint64, buildData *BuildData, currentVerification *resources.Verification,
	newStageRunnersChan chan *stagerunner.StageRunner, finishFunc func()) {
	newMachinesChan, errorChan := buildData.Pool.GetReady(numNodes)
	go func(newMachinesChan <-chan vm.VirtualMachine) {
		for newMachine := range newMachinesChan {
			if currentVerification.Status != "passed" && currentVerification.Status != "failed" && currentVerification.Status != "cancelled" {
				go func() {
					buildRunner.RunStages(
						newMachine, buildData, currentVerification, newStageRunnersChan, finishFunc)
					buildData.Pool.Free()
					newMachine.Terminate()
				}()
			} else {
				buildData.Pool.Return(newMachine)
			}
		}
	}(newMachinesChan)

	// TODO (bbland): do something with the errorChan
	go func(errorChan <-chan error) {
		<-errorChan
	}(errorChan)
}

func (buildRunner *BuildRunner) RunStages(virtualMachine vm.VirtualMachine,
	buildData *BuildData, verification *resources.Verification,
	newStageRunnersChan chan *stagerunner.StageRunner, finishFunc func()) {
	stageRunner := stagerunner.New(
		buildRunner.resourcesConnection, virtualMachine, verification,
		new(stagerunner.S3Exporter))
	defer close(stageRunner.ResultsChan)

	newStageRunnersChan <- stageRunner

	err := stageRunner.RunStages(buildData.VerificationConfig.Sections,
		buildData.VerificationConfig.FinalSections,
		buildData.VerificationConfig.Params.Environment)
	if err != nil {
		stacktrace := make([]byte, 4096)
		stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
		log.Errorf("Failed to run stages for verification: %v\nConfig: %#v\nError: %v\n%s", verification,
			buildData.VerificationConfig, err, stacktrace)
	}
	finishFunc()
}

func (buildRunner *BuildRunner) combineResults(currentVerification *resources.Verification, newStageRunnersChan <-chan *stagerunner.StageRunner) <-chan verification.SectionResult {
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
					buildRunner.resourcesConnection.Verifications.Update.SetStartTime(currentVerification.Id, time.Now())
					buildRunner.resourcesConnection.Verifications.Update.SetStatus(currentVerification.Id, "running")
					currentVerification.Status = "running"
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
			for extraResult := range stageRunner.ResultsChan {
				log.Debug("Got an extra result: %v", extraResult)
			}
		}
	}(newStageRunnersChan)
	return resultsChan
}
