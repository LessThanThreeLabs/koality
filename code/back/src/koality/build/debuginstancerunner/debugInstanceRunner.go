package debuginstancerunner

import (
	"fmt"
	"koality/build/runner"
	"koality/build/stagerunner"
	"koality/notify"
	"koality/repositorymanager"
	"koality/resources"
	"koality/vm"
	"koality/vm/poolmanager"
	"time"
)

type DebugInstanceRunner struct {
	resourcesConnection                *resources.Connection
	poolManager                        *poolmanager.PoolManager
	repositoryManager                  repositorymanager.RepositoryManager
	debugInstanceCreatedSubscriptionId resources.SubscriptionId
	notifier                           notify.DebugInstanceNotifier
	buildRunner                        *runner.BuildRunner
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager, repositoryManager repositorymanager.RepositoryManager, notifier notify.Notifier) *DebugInstanceRunner {
	return &DebugInstanceRunner{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
		repositoryManager:   repositoryManager,
		notifier:            notifier,
		buildRunner:         runner.New(resourcesConnection, poolManager, repositoryManager),
	}
}

func (debugInstanceRunner *DebugInstanceRunner) SubscribeToEvents() (err error) {
	onDebugInstanceCreated := func(debugInstance *resources.DebugInstance) {
		debugInstanceRunner.RunDebugInstance(debugInstance)
	}
	debugInstanceRunner.debugInstanceCreatedSubscriptionId, err = debugInstanceRunner.resourcesConnection.DebugInstances.Subscription.SubscribeToCreatedEvents(onDebugInstanceCreated)
	return
}

func (debugInstanceRunner *DebugInstanceRunner) UnsubscribeFromEvents() (err error) {
	if debugInstanceRunner.debugInstanceCreatedSubscriptionId == 0 {
		return fmt.Errorf("Debug instance created events not subscribed to")
	} else {
		err = debugInstanceRunner.resourcesConnection.DebugInstances.Subscription.UnsubscribeFromCreatedEvents(debugInstanceRunner.debugInstanceCreatedSubscriptionId)
		if err == nil {
			debugInstanceRunner.debugInstanceCreatedSubscriptionId = 0
		}
	}
	return
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

	finishFunc := func(vm vm.VirtualMachine) {
		err := debugInstanceRunner.notifier.NotifyDebugInstance(vm, build, debugInstance)
		if err != nil {
			// TODO(dhuang) what do...
		}
		time.Sleep(debugInstance.Expires.Sub(time.Now()))
	}

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)
	debugInstanceRunner.buildRunner.RunStagesOnNewMachines(
		numNodes, buildData, build, newStageRunnersChan, finishFunc)

	return debugInstanceRunner.buildRunner.ProcessResults(build, newStageRunnersChan, buildData)
}
