package runner

import (
	"bytes"
	"fmt"
	"github.com/LessThanThreeLabs/go.codereview/patch"
	"koality/build"
	"koality/build/config"
	"koality/build/config/commandgroup"
	"koality/build/config/section"
	"koality/build/stagerunner"
	"koality/repositorymanager"
	"koality/resources"
	"koality/shell"
	"koality/shell/shellutil"
	"koality/util/log"
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
	config.BuildConfig
	Pool vm.VirtualMachinePool
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager, repositoryManager repositorymanager.RepositoryManager) *BuildRunner {
	return &BuildRunner{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
		repositoryManager:   repositoryManager,
	}
}

func (buildRunner *BuildRunner) GetBuildData(build *resources.Build) (*BuildData, error) {
	if build == nil {
		return nil, fmt.Errorf("Cannot run a nil build")
	}
	log.Info("Running build: %v", build)
	repository, err := buildRunner.resourcesConnection.Repositories.Read.Get(build.RepositoryId)
	if err != nil {
		buildRunner.failBuild(build)
		return nil, err
	}

	buildConfig, err := buildRunner.getBuildConfig(build, repository)
	if err != nil {
		buildRunner.failBuild(build)
		return nil, err
	}

	virtualMachinePool, err := buildRunner.poolManager.GetPool(buildConfig.Params.PoolId)
	if err != nil {
		buildRunner.failBuild(build)
		return nil, err
	}
	return &BuildData{repository, buildConfig, virtualMachinePool}, nil
}

func (buildRunner *BuildRunner) ProcessResults(currentBuild *resources.Build, newStageRunnersChan chan *stagerunner.StageRunner, buildData *BuildData) (bool, error) {

	resultsChan := buildRunner.combineResults(currentBuild, newStageRunnersChan)
	receivedResult := make(map[string]bool)

	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			if currentBuild.Status == "running" {
				err := buildRunner.passBuild(currentBuild, buildData.Repository)
				if err != nil {
					return false, err
				}
			}
			return currentBuild.Status == "passed", nil
		}
		if !result.Final {
			if !result.Passed {
				switch result.FailSectionOn {
				case section.FailOnNever:
					// Do nothing
				case section.FailOnAny:
					if currentBuild.Status != "cancelled" && currentBuild.Status != "failed" {
						err := buildRunner.failBuild(currentBuild)
						if err != nil {
							return false, err
						}
					}
				case section.FailOnFirst:
					if !receivedResult[result.Section] {
						if currentBuild.Status != "cancelled" && currentBuild.Status != "failed" {
							err := buildRunner.failBuild(currentBuild)
							if err != nil {
								return false, err
							}
						}
					}
				}
			}
			receivedResult[result.Section] = true
			if len(receivedResult) == len(buildData.BuildConfig.Sections) && currentBuild.Status == "running" {
				err := buildRunner.passBuild(currentBuild, buildData.Repository)
				if err != nil {
					return false, err
				}
			}
		}
	}
}

func (buildRunner *BuildRunner) failBuild(build *resources.Build) error {
	(*build).Status = "failed"
	log.Infof("build %d %sFAILED!!!%s\n", build.Id, shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := buildRunner.resourcesConnection.Builds.Update.SetStatus(build.Id, "failed")
	if err != nil {
		return err
	}

	err = buildRunner.resourcesConnection.Builds.Update.SetEndTime(build.Id, time.Now())
	return err
}

func (buildRunner *BuildRunner) passBuild(build *resources.Build, repository *resources.Repository) error {
	log.Infof("build %d %sPASSED!!!%s\n", build.Id, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := buildRunner.repositoryManager.MergeChangeset(repository, repositorymanager.GitHiddenRef(build.Changeset.HeadSha), build.Changeset.BaseSha, build.MergeTarget)
	if err == nil {
		(*build).MergeStatus = "passed"
		buildRunner.resourcesConnection.Builds.Update.SetMergeStatus(build.Id, "passed")
	} else {
		(*build).MergeStatus = "failed"
		buildRunner.resourcesConnection.Builds.Update.SetMergeStatus(build.Id, "failed")
	}

	(*build).Status = "passed"
	err = buildRunner.resourcesConnection.Builds.Update.SetStatus(build.Id, "passed")
	if err != nil {
		return err
	}

	err = buildRunner.resourcesConnection.Builds.Update.SetEndTime(build.Id, time.Now())
	return err
}

func (buildRunner *BuildRunner) getBuildConfig(currentBuild *resources.Build, repository *resources.Repository) (config.BuildConfig, error) {
	var emptyConfig config.BuildConfig

	configYaml, err := buildRunner.repositoryManager.GetYamlFile(repository, currentBuild.Changeset.HeadSha)
	if err != nil {
		return emptyConfig, err
	}

	if len(currentBuild.Changeset.PatchContents) > 0 {
		patchSet, err := patch.Parse(currentBuild.Changeset.PatchContents)
		if err != nil {
			return emptyConfig, err
		}
		for _, filePatch := range patchSet.File {
			if filePatch.Src == "koality.yml" && filePatch.Dst == "koality.yml" {
				patchedYaml, err := filePatch.Apply([]byte(configYaml))
				// TODO (bbland): log this error, set Nodes=1
				if err == nil {
					configYaml = string(patchedYaml)
				}
			}
		}
	}

	buildConfig, err := config.FromYaml(configYaml, repository.Name)
	if err != nil {
		return emptyConfig, err
	}

	// TODO (bbland): add retry logic
	checkoutCommand, err := buildRunner.repositoryManager.GetCheckoutCommand(repository, currentBuild.Changeset.HeadSha)
	if err != nil {
		return emptyConfig, err
	}

	setupCommands := []build.Command{build.NewShellCommand(repository.VcsType, checkoutCommand)}
	if len(currentBuild.Changeset.PatchContents) > 0 {
		setupCommands = append(setupCommands, build.NewBasicCommand("patch", shellutil.CreatePatchExecutable(repository.Name, bytes.NewReader(currentBuild.Changeset.PatchContents))))
	}
	setupSection := section.New("setup", false, section.RunOnAll, section.FailOnFirst, false, nil, commandgroup.New(setupCommands), nil)
	buildConfig.Sections = append([]section.Section{setupSection}, buildConfig.Sections...)
	return buildConfig, nil
}

func (buildRunner *BuildRunner) CreateStages(currentBuild *resources.Build, buildConfig *config.BuildConfig) error {
	sections, finalSections := buildConfig.Sections, buildConfig.FinalSections
	for sectionNumber, section := range append(sections, finalSections...) {
		var err error

		stageNumber := 0

		factoryCommands := section.FactoryCommands(true)
		for command, err := factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
			_, err := buildRunner.resourcesConnection.Stages.Create.Create(currentBuild.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
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
			_, err := buildRunner.resourcesConnection.Stages.Create.Create(currentBuild.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
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
			_, err := buildRunner.resourcesConnection.Stages.Create.Create(currentBuild.Id, uint64(sectionNumber), section.Name()+".export", uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
	}
	return nil
}

func (buildRunner *BuildRunner) RunStagesOnNewMachines(numNodes uint64, buildData *BuildData, currentBuild *resources.Build, newStageRunnersChan chan *stagerunner.StageRunner, finishFunc func(vm.VirtualMachine)) {
	newMachinesChan, errorChan := buildData.Pool.GetReady(numNodes)
	go func(newMachinesChan <-chan vm.VirtualMachine) {
		for newMachine := range newMachinesChan {
			if currentBuild.Status != "passed" && currentBuild.Status != "failed" && currentBuild.Status != "cancelled" {
				go func(virtualMachine vm.VirtualMachine) {
					defer virtualMachine.Terminate()
					defer buildData.Pool.Free()
					buildRunner.RunStages(
						virtualMachine, buildData, currentBuild, newStageRunnersChan, finishFunc)
				}(newMachine)
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

func (buildRunner *BuildRunner) RunStages(virtualMachine vm.VirtualMachine, buildData *BuildData, build *resources.Build, newStageRunnersChan chan *stagerunner.StageRunner, finishFunc func(vm.VirtualMachine)) {
	stageRunner := stagerunner.New(
		buildRunner.resourcesConnection, virtualMachine, build,
		new(stagerunner.S3Exporter))
	defer close(stageRunner.ResultsChan)

	newStageRunnersChan <- stageRunner

	err := stageRunner.RunStages(buildData.BuildConfig.Sections,
		buildData.BuildConfig.FinalSections,
		buildData.BuildConfig.Params.Environment)
	if err != nil {
		stacktrace := make([]byte, 4096)
		stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
		log.Errorf("Failed to run stages for build: %v\nConfig: %#v\nError: %v\n%s", build,
			buildData.BuildConfig, err, stacktrace)
	}
	finishFunc(virtualMachine)
}

func (buildRunner *BuildRunner) combineResults(currentBuild *resources.Build, newStageRunnersChan <-chan *stagerunner.StageRunner) <-chan build.SectionResult {
	resultsChan := make(chan build.SectionResult)
	go func(newStageRunnersChan <-chan *stagerunner.StageRunner) {
		isStarted := false

		combinedResults := make(chan build.SectionResult)
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
					buildRunner.resourcesConnection.Builds.Update.SetStartTime(currentBuild.Id, time.Now())
					buildRunner.resourcesConnection.Builds.Update.SetStatus(currentBuild.Id, "running")
					currentBuild.Status = "running"
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
