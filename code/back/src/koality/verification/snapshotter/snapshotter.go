package snapshotter

type Snapshotter struct {
	resourcesConnection           *resources.Connection
	poolManager                   *poolmanager.PoolManager
	SnapshotCreatedSubscriptionId resources.SubscriptionId
}

func New(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager) *Snapshotter {
	return &Snapshotter{
		resourcesConnection: resourcesConnection,
		poolManager:         poolManager,
	}
}

func (snapshotter *Snapshotter) SubscribeToEvents() error {
	onVerificationCreated := func(snapshot *resources.Verification) {
		snapshotter.MakeSnapshot(snapshot)
	}

	var err error

	snapshotter.snapshotCreatedSubscriptionId, err = snapshotter.resourcesConnection.Snapshots.Subscription.SubscribeToCreatedEvents(onVerificationCreated)

	return err
}

func (snapshotter *Snapshotter) UnsubscribeFromEvents() error {
	var err error

	if snapshotter.snapshotCreatedSubscriptionId == 0 {
		return fmt.Errorf("Verification created events not subscribed to")
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

	// clone shit

	// for each verification

	// 		get config

	// 		config -> get vm

	// 		figure out which shit to run (by reading the config)

	//		run said shit (controlling snapshot + verification states)

	// 	if all passed -> make snapshot success story\

	// 	vm.CaptureState -> save instanceId

}
