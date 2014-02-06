package testrunner

import (
	"fmt"
	"koality/build/runner"
	"koality/build/stagerunner"
	"koality/repositorymanager"
	"koality/resources"
	"koality/util/log"
	"koality/vm/poolmanager"
)

type TestRunner struct {
	resourcesConnection            *resources.Connection
	poolManager                    *poolmanager.PoolManager
	repositoryManager              repositorymanager.RepositoryManager
	testBuildCreatedSubscriptionId resources.SubscriptionId
	buildRunner                    *runner.BuildRunner
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager, repositoryManager repositorymanager.RepositoryManager) *TestRunner {
	return &TestRunner{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
		repositoryManager:   repositoryManager,
		buildRunner:         runner.New(resourcesConnection, poolManager, repositoryManager),
	}
}

func (testRunner *TestRunner) SubscribeToEvents() error {
	onTestCreated := func(testBuild *resources.Build) {
		testRunner.RunBuild(testBuild)
	}
	var err error
	testRunner.testBuildCreatedSubscriptionId, err = testRunner.resourcesConnection.Builds.Subscription.SubscribeToCreatedEvents(onTestCreated)
	return err
}

func (testRunner *TestRunner) UnsubscribeFromEvents() error {
	var err error

	if testRunner.testBuildCreatedSubscriptionId == 0 {
		return fmt.Errorf("Build created events not subscribed to")
	} else {
		unsubscribeError := testRunner.resourcesConnection.Builds.Subscription.UnsubscribeFromCreatedEvents(testRunner.testBuildCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		} else {
			testRunner.testBuildCreatedSubscriptionId = 0
		}
	}
	return err
}

func (testRunner *TestRunner) RunBuild(currentBuild *resources.Build) (bool, error) {
	buildData, err := testRunner.buildRunner.GetBuildData(currentBuild)
	if err != nil {
		// TODO handle errors
		return false, err
	}

	err = testRunner.buildRunner.CreateStages(currentBuild, &buildData.BuildConfig)
	if err != nil {
		return false, err
	}

	numNodes := buildData.BuildConfig.Params.Nodes
	if numNodes == 0 {
		numNodes = 1 // TODO (bbland): Do a better job guessing the number of nodes when unspecified
	}
	log.Debugf("Using %d nodes for build: %v", numNodes, currentBuild)

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)
	emptyFinishFunc := func() {}

	testRunner.buildRunner.RunStagesOnNewMachines(
		numNodes, buildData, currentBuild, newStageRunnersChan, emptyFinishFunc)
	return testRunner.buildRunner.ProcessResults(currentBuild, newStageRunnersChan, buildData)
}
