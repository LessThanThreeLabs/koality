package debuginstancerunner

import (
	"fmt"
	"koality/build/runner"
	"koality/build/stagerunner"
	"koality/repositorymanager"
	"koality/resources"
	"koality/vm/poolmanager"
	"time"
)

type DebugInstanceRunner struct {
	resourcesConnection                *resources.Connection
	poolManager                        *poolmanager.PoolManager
	repositoryManager                  repositorymanager.RepositoryManager
	debugInstanceCreatedSubscriptionId resources.SubscriptionId
	buildRunner                        *runner.BuildRunner
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager,
	repositoryManager repositorymanager.RepositoryManager) *DebugInstanceRunner {
	return &DebugInstanceRunner{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
		repositoryManager:   repositoryManager,
		buildRunner:         runner.New(resourcesConnection, poolManager, repositoryManager),
	}
}

func (debugInstanceRunner *DebugInstanceRunner) SubscribeToEvents() error {
	onDebugInstanceCreated := func(debugInstance *resources.DebugInstance) {
		debugInstanceRunner.RunDebugInstance(debugInstance)
	}
	var err error
	debugInstanceRunner.debugInstanceCreatedSubscriptionId, err = debugInstanceRunner.resourcesConnection.DebugInstances.Subscription.SubscribeToCreatedEvents(onDebugInstanceCreated)
	return err
}

func (debugInstanceRunner *DebugInstanceRunner) UnsubscribeFromEvents() error {
	var err error

	if debugInstanceRunner.debugInstanceCreatedSubscriptionId == 0 {
		return fmt.Errorf("Debug instance created events not subscribed to")
	} else {
		unsubscribeError := debugInstanceRunner.resourcesConnection.DebugInstances.Subscription.UnsubscribeFromCreatedEvents(debugInstanceRunner.debugInstanceCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		} else {
			debugInstanceRunner.debugInstanceCreatedSubscriptionId = 0
		}
	}
	return err
}

func (debugInstanceRunner *DebugInstanceRunner) RunDebugInstance(debugInstance *resources.DebugInstance) (bool, error) {
	build, err := debugInstanceRunner.resourcesConnection.Builds.Read.Get(debugInstance.BuildId)
	if err != nil {
		// TODO(dhuang) handle errors here?
		return false, err
	}

	buildData, err := debugInstanceRunner.buildRunner.GetBuildData(build)
	if err != nil {
		return false, err
	}

	for i, section := range buildData.BuildConfig.Sections {
		if section.Name() == buildData.BuildConfig.Params.SnapshotUntil {
			buildData.BuildConfig.Sections = buildData.BuildConfig.Sections[:i]
			break
		}
	}
	buildData.BuildConfig.FinalSections = nil
	numNodes := uint64(1)
	err = debugInstanceRunner.buildRunner.CreateStages(build, &buildData.BuildConfig)
	if err != nil {
		return false, err
	}

	finishFunc := func() {
		time.Sleep(debugInstance.Expires.Sub(time.Now()))
	}

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)
	debugInstanceRunner.buildRunner.RunStagesOnNewMachines(
		numNodes, buildData, build, newStageRunnersChan, finishFunc)

	return debugInstanceRunner.buildRunner.ProcessResults(build, newStageRunnersChan, buildData)
}
