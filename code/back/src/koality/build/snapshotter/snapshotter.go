package snapshotter

import (
	"fmt"
	"koality/build/runner"
	"koality/build/stagerunner"
	"koality/repositorymanager"
	"koality/resources"
	"koality/vm"
	"koality/vm/poolmanager"
	"time"
)

type Snapshotter struct {
	resourcesConnection           *resources.Connection
	poolManager                   *poolmanager.PoolManager
	repositoryManager             repositorymanager.RepositoryManager
	snapshotCreatedSubscriptionId resources.SubscriptionId
	buildRunner                   *runner.BuildRunner
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager) *Snapshotter {
	return &Snapshotter{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
	}
}

func (snapshotter *Snapshotter) SubscribeToEvents() error {
	onSnapshotCreated := func(snapshot *resources.Snapshot) {
		snapshotter.MakeSnapshot(snapshot)
	}

	var err error
	snapshotter.snapshotCreatedSubscriptionId, err = snapshotter.resourcesConnection.Snapshots.Subscription.SubscribeToCreatedEvents(onSnapshotCreated)
	return err
}

func (snapshotter *Snapshotter) UnsubscribeFromEvents() error {
	var err error

	if snapshotter.snapshotCreatedSubscriptionId == 0 {
		return fmt.Errorf("Snapshot created events not subscribed to")
	} else {
		unsubscribeError := snapshotter.resourcesConnection.Snapshots.Subscription.UnsubscribeFromCreatedEvents(snapshotter.snapshotCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		} else {
			snapshotter.snapshotCreatedSubscriptionId = 0
		}
	}
	return err
}

func (snapshotter *Snapshotter) MakeSnapshot(snapshot *resources.Snapshot) (bool, error) {
	builds, err := snapshotter.resourcesConnection.Builds.Read.GetForSnapshot(snapshot.Id)
	if err != nil {
		return false, err
	}

	virtualMachinePool, err := snapshotter.poolManager.GetPool(snapshot.PoolId)
	if err != nil {
		snapshotter.failSnapshot(snapshot)
		return false, err
	}

	defer virtualMachinePool.Free()

	newMachinesChan, _ := virtualMachinePool.GetReady(1)
	virtualMachine := <-newMachinesChan
	defer virtualMachine.Terminate()

	for _, build := range builds {
		buildData, err := snapshotter.buildRunner.GetBuildData(&build)

		if err != nil {
			snapshotter.failSnapshot(snapshot)
			return false, err
		}

		for i, section := range buildData.BuildConfig.Sections {
			if section.Name() == buildData.BuildConfig.Params.SnapshotUntil {
				buildData.BuildConfig.Sections = buildData.BuildConfig.Sections[:i]
				break
			}
		}
		buildData.BuildConfig.FinalSections = nil
		err = snapshotter.buildRunner.CreateStages(&build, &buildData.BuildConfig)
		if err != nil {
			snapshotter.failSnapshot(snapshot)
			return false, err
		}

		emptyFinishFunc := func(vm vm.VirtualMachine) {}

		newStageRunnersChan := make(chan *stagerunner.StageRunner, 1)

		snapshotter.buildRunner.RunStages(virtualMachine, buildData, &build, newStageRunnersChan, emptyFinishFunc)

		success, err := snapshotter.buildRunner.ProcessResults(&build, newStageRunnersChan, buildData)
		if !success {
			snapshotter.failSnapshot(snapshot)
			break
		}
	}

	//TODO(akostov) Decide on snapshot naming
	imageId, err := virtualMachine.SaveState("SomeCoolSnapshotName")
	if err != nil {
		snapshotter.failSnapshot(snapshot)
		return false, err
	}

	if err = snapshotter.resourcesConnection.Snapshots.Update.SetImageId(snapshot.Id, imageId); err != nil {
		return false, err
	}

	if err = snapshotter.passSnapshot(snapshot); err != nil {
		return false, err
	}
	return true, nil
}

func (snapshotter *Snapshotter) failSnapshot(snapshot *resources.Snapshot) error {
	(*snapshot).Status = "failed"
	err := snapshotter.resourcesConnection.Snapshots.Update.SetStatus(snapshot.Id, "failed")
	if err != nil {
		return err
	}

	err = snapshotter.resourcesConnection.Snapshots.Update.SetEndTime(snapshot.Id, time.Now())
	return err
}

func (snapshotter *Snapshotter) passSnapshot(snapshot *resources.Snapshot) error {
	(*snapshot).Status = "passed"
	err := snapshotter.resourcesConnection.Snapshots.Update.SetStatus(snapshot.Id, "passed")
	if err != nil {
		return err
	}

	err = snapshotter.resourcesConnection.Snapshots.Update.SetEndTime(snapshot.Id, time.Now())
	return err
}
