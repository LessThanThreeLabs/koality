package testrunner

import (
	"fmt"
	"koality/repositorymanager"
	"koality/resources"
	"koality/util/log"
	"koality/verification/runner"
	"koality/verification/stagerunner"
	"koality/vm/poolmanager"
)

type TestRunner struct {
	resourcesConnection            *resources.Connection
	poolManager                    *poolmanager.PoolManager
	repositoryManager              repositorymanager.RepositoryManager
	testBuildCreatedSubscriptionId resources.SubscriptionId
	buildRunner                    *runner.BuildRunner
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager,
	repositoryManager repositorymanager.RepositoryManager) *TestRunner {
	return &TestRunner{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
		repositoryManager:   repositoryManager,
		buildRunner:         runner.New(resourcesConnection, poolManager, repositoryManager),
	}
}

func (testRunner *TestRunner) SubscribeToEvents() error {
	onTestCreated := func(testBuild *resources.Verification) {
		testRunner.RunVerification(testBuild)
	}
	var err error
	testRunner.testBuildCreatedSubscriptionId, err = testRunner.resourcesConnection.Verifications.Subscription.SubscribeToCreatedEvents(onTestCreated)
	return err
}

func (testRunner *TestRunner) UnsubscribeFromEvents() error {
	var err error

	if testRunner.testBuildCreatedSubscriptionId == 0 {
		return fmt.Errorf("Verification created events not subscribed to")
	} else {
		unsubscribeError := testRunner.resourcesConnection.Verifications.Subscription.UnsubscribeFromCreatedEvents(testRunner.testBuildCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		} else {
			testRunner.testBuildCreatedSubscriptionId = 0
		}
	}
	return err
}

func (testRunner *TestRunner) RunVerification(currentVerification *resources.Verification) (bool, error) {
	buildData, err := testRunner.buildRunner.GetBuildData(currentVerification)
	if err != nil {
		// TODO handle errors
		return false, err
	}

	err = testRunner.buildRunner.CreateStages(currentVerification, &buildData.VerificationConfig)
	if err != nil {
		return false, err
	}

	numNodes := buildData.VerificationConfig.Params.Nodes
	if numNodes == 0 {
		numNodes = 1 // TODO (bbland): Do a better job guessing the number of nodes when unspecified
	}
	log.Debugf("Using %d nodes for verification: %v", numNodes, currentVerification)

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)
	emptyFinishFunc := func() {}

	testRunner.buildRunner.RunStagesOnNewMachines(
		numNodes, buildData, currentVerification, newStageRunnersChan, emptyFinishFunc)
	return testRunner.buildRunner.ProcessResults(currentVerification, newStageRunnersChan, buildData)
}
