package snapshots_test

import (
	"koality/resources"
	"koality/resources/database"
	"testing"
	"time"
)

func TestCreateInvalidSnapshot(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	_, err = connection.Snapshots.Create.Create(1337, nil)
	if _, ok := err.(resources.NoSuchPoolError); !ok {
		test.Fatal("Expected NoSuchPoolError when providing invalid pool id")
	}
}

func TestCreateAndDeleteSnapshot(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	createdEventReceived := make(chan bool, 1)
	var createdEventSnapshot *resources.Snapshot
	snapshotCreatedHandler := func(snapshot *resources.Snapshot) {
		createdEventSnapshot = snapshot
		createdEventReceived <- true
	}
	_, err = connection.Snapshots.Subscription.SubscribeToCreatedEvents(snapshotCreatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	deletedEventReceived := make(chan bool, 1)
	deletedEventId := uint64(0)
	snapshotDeletedHandler := func(snapshotId uint64) {
		deletedEventId = snapshotId
		deletedEventReceived <- true
	}
	_, err = connection.Snapshots.Subscription.SubscribeToDeletedEvents(snapshotDeletedHandler)
	if err != nil {
		test.Fatal(err)
	}

	pools, err := connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	firstPool := pools[0]
	snapshot, err := connection.Snapshots.Create.Create(firstPool.Id, nil)
	if err != nil {
		test.Fatal(err)
	}

	if snapshot.PoolId != firstPool.Id {
		test.Fatal("snapshot.PoolId mismatch")
	}

	select {
	case <-createdEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear snapshot creation event")
	}

	if createdEventSnapshot.Id != snapshot.Id {
		test.Fatal("Bad snapshot.Id in snapshot creation event")
	} else if createdEventSnapshot.PoolId != snapshot.PoolId {
		test.Fatal("Bad snapshot.PoolId in snapshot creation event")
	}

	snapshot2, err := connection.Snapshots.Read.Get(snapshot.Id)
	if err != nil {
		test.Fatal(err)
	}

	if snapshot.Id != snapshot2.Id {
		test.Fatal("snapshot.Id mismatch")
	} else if snapshot.PoolId != snapshot2.PoolId {
		test.Fatal("snapshot.PoolId mismatch")
	}

	snapshot3, err := connection.Snapshots.Read.GetByImageId(snapshot.ImageId)
	if err != nil {
		test.Fatal(err)
	}

	if snapshot.Id != snapshot3.Id {
		test.Fatal("snapshot.Id mismatch")
	} else if snapshot.PoolId != snapshot3.PoolId {
		test.Fatal("snapshot.PoolId mismatch")
	}

	snapshots, err := connection.Snapshots.Read.GetAllForPool(firstPool.Id)
	if err != nil {
		test.Fatal(err)
	} else if len(snapshots) != 1 {
		test.Fatal("Expected only 1 snapshot")
	}

	err = connection.Snapshots.Delete.Delete(snapshot.Id)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-deletedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear snapshot deletion event")
	}

	if deletedEventId != snapshot.Id {
		test.Fatal("Bad snapshot.Id in snapshot deletion event")
	}

	err = connection.Snapshots.Delete.Delete(snapshot.Id)
	if _, ok := err.(resources.NoSuchSnapshotError); !ok {
		test.Fatal("Expected NoSuchSnapshotError when trying to delete same snapshot twice")
	}

	deletedSnapshot, err := connection.Snapshots.Read.Get(snapshot.Id)
	if err != nil {
		test.Fatal(err)
	} else if !deletedSnapshot.IsDeleted {
		test.Fatal("Expected snapshot to be marked as deleted")
	}
}

func TestSnapshotStatuses(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	snapshotStatusEventReceived := make(chan bool, 1)
	snapshotStatusEventId := uint64(0)
	snapshotStatusEventStatus := ""
	snapshotStatusUpdatedHandler := func(snapshotId uint64, status string) {
		snapshotStatusEventId = snapshotId
		snapshotStatusEventStatus = status
		snapshotStatusEventReceived <- true
	}
	_, err = connection.Snapshots.Subscription.SubscribeToStatusUpdatedEvents(snapshotStatusUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	pools, err := connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	firstPool := pools[0]
	snapshot, err := connection.Snapshots.Create.Create(firstPool.Id, nil)
	if err != nil {
		test.Fatal(err)
	}

	if snapshot.Status != "declared" {
		test.Fatal("Expected initial snapshot status to be 'declared'")
	}

	err = connection.Snapshots.Update.SetStatus(snapshot.Id, "succeeded")
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-snapshotStatusEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear snapshot status updated event")
	}

	if snapshotStatusEventId != snapshot.Id {
		test.Fatal("Bad snapshot.Id in status updated event")
	} else if snapshotStatusEventStatus != "succeeded" {
		test.Fatal("Bad snapshot status in status updated event")
	}

	snapshot, err = connection.Snapshots.Read.Get(snapshot.Id)
	if err != nil {
		test.Fatal(err)
	} else if snapshot.Status != "succeeded" {
		test.Fatal("Failed to update snapshot status")
	}

	err = connection.Snapshots.Update.SetStatus(snapshot.Id, "bad-status")
	if _, ok := err.(resources.InvalidSnapshotStatusError); !ok {
		test.Fatal("Expected InvalidSnapshotStatusError when trying to set status")
	}
}

func TestSnapshotTimes(test *testing.T) {
	if err := database.PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := database.New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	snapshotStartTimeEventReceived := make(chan bool, 1)
	snapshotStartTimeEventId := uint64(0)
	snapshotStartTimeEventTime := time.Now()
	snapshotStartTimeUpdatedHandler := func(snapshotId uint64, startTime time.Time) {
		snapshotStartTimeEventId = snapshotId
		snapshotStartTimeEventTime = startTime
		snapshotStartTimeEventReceived <- true
	}
	_, err = connection.Snapshots.Subscription.SubscribeToStartTimeUpdatedEvents(snapshotStartTimeUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	snapshotEndTimeEventReceived := make(chan bool, 1)
	snapshotEndTimeEventId := uint64(0)
	snapshotEndTimeEventTime := time.Now()
	snapshotEndTimeUpdatedHandler := func(snapshotId uint64, endTime time.Time) {
		snapshotEndTimeEventId = snapshotId
		snapshotEndTimeEventTime = endTime
		snapshotEndTimeEventReceived <- true
	}
	_, err = connection.Snapshots.Subscription.SubscribeToEndTimeUpdatedEvents(snapshotEndTimeUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	pools, err := connection.Pools.Read.GetAllEc2Pools()
	if err != nil {
		test.Fatal(err)
	}

	firstPool := pools[0]
	snapshot, err := connection.Snapshots.Create.Create(firstPool.Id, nil)
	if err != nil {
		test.Fatal(err)
	}

	err = connection.Snapshots.Update.SetEndTime(snapshot.Id, time.Now())
	if err == nil {
		test.Fatal("Expected error when setting end time without start time")
	}

	err = connection.Snapshots.Update.SetStartTime(snapshot.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting start time before create time")
	}

	startTime := time.Now()
	err = connection.Snapshots.Update.SetStartTime(snapshot.Id, startTime)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-snapshotStartTimeEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear snapshot start time event")
	}

	if snapshotStartTimeEventId != snapshot.Id {
		test.Fatal("Bad snapshot.Id in start time event")
	} else if snapshotStartTimeEventTime != startTime {
		test.Fatal("Bad snapshot start time in start time event")
	}

	err = connection.Snapshots.Update.SetStartTime(0, time.Now())
	if _, ok := err.(resources.NoSuchSnapshotError); !ok {
		test.Fatal("Expected NoSuchSnapshotError when trying to set start time for nonexistent snapshot")
	}

	err = connection.Snapshots.Update.SetEndTime(snapshot.Id, time.Unix(0, 0))
	if err == nil {
		test.Fatal("Expected error when setting end time before create time")
	}

	endTime := time.Now()
	err = connection.Snapshots.Update.SetEndTime(snapshot.Id, endTime)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-snapshotEndTimeEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear snapshot end time event")
	}

	if snapshotEndTimeEventId != snapshot.Id {
		test.Fatal("Bad snapshot.Id in end time event")
	} else if snapshotEndTimeEventTime != endTime {
		test.Fatal("Bad snapshot end time in end time event")
	}

	err = connection.Snapshots.Update.SetEndTime(0, time.Now())
	if _, ok := err.(resources.NoSuchSnapshotError); !ok {
		test.Fatal("Expected NoSuchSnapshotError when trying to set end time for nonexistent snapshot")
	}
}
