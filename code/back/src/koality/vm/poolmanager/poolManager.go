package poolmanager

import (
	"fmt"
	"koality/resources"
	"koality/util/log"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"koality/vm/ec2/ec2vm"
	"runtime"
	"sync"
)

func New(virtualMachinePools []vm.VirtualMachinePool) *PoolManager {
	virtualMachinePoolMap := make(map[uint64]vm.VirtualMachinePool, len(virtualMachinePools))
	for _, virtualMachinePool := range virtualMachinePools {
		virtualMachinePoolMap[virtualMachinePool.Id()] = virtualMachinePool
	}
	return &PoolManager{
		poolMap: virtualMachinePoolMap,
		locker:  new(sync.Mutex),
	}
}

type PoolManager struct {
	poolMap          map[uint64]vm.VirtualMachinePool
	locker           sync.Locker
	subscriptionInfo *subscriptionInfo
}

type subscriptionInfo struct {
	resourcesConnection                  *resources.Connection
	ec2Broker                            *ec2broker.Ec2Broker
	ec2PoolCreatedSubscriptionId         resources.SubscriptionId
	ec2PoolDeletedSubscriptionId         resources.SubscriptionId
	ec2PoolSettingsUpdatedSubscriptionId resources.SubscriptionId
}

func (poolManager *PoolManager) GetPool(poolId uint64) (vm.VirtualMachinePool, error) {
	poolManager.locker.Lock()
	defer poolManager.locker.Unlock()

	pool, ok := poolManager.poolMap[poolId]
	if !ok {
		return nil, fmt.Errorf("No pool found with id %d", poolId)
	}

	return pool, nil
}

func (poolManager *PoolManager) SubscribeToEvents(resourcesConnection *resources.Connection, ec2Broker *ec2broker.Ec2Broker) error {
	if poolManager.subscriptionInfo != nil {
		return fmt.Errorf("Pool manager already subscribed to events")
	}

	onEc2PoolCreated := func(ec2Pool *resources.Ec2Pool) {
		poolManager.locker.Lock()
		defer poolManager.locker.Unlock()

		ec2Manager, err := ec2vm.NewManager(ec2Broker, ec2Pool, resourcesConnection)
		if err != nil {
			stacktrace := make([]byte, 4096)
			stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
			log.Errorf("Failed to construct new ec2 launcher with pool parameters: %v\n%v\n%s", ec2Pool, err, stacktrace)
		}

		ec2VirtualMachinePool := ec2vm.NewPool(ec2Manager)
		poolManager.poolMap[ec2Pool.Id] = ec2VirtualMachinePool
	}
	onEc2PoolDeleted := func(ec2PoolId uint64) {
		poolManager.locker.Lock()
		defer poolManager.locker.Unlock()

		delete(poolManager.poolMap, ec2PoolId)
	}
	onEc2PoolSettingsUpdated := func(ec2PoolId uint64, accessKey, secretKey, username, baseAmiId, securityGroupId,
		vpcSubnetId, instanceType string, numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) {
		poolManager.locker.Lock()
		vmPool, ok := poolManager.poolMap[ec2PoolId]
		poolManager.locker.Unlock()
		if !ok {
			stacktrace := make([]byte, 4096)
			stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
			log.Errorf("Tried to update nonexistent pool with id: %d\n%s", ec2PoolId, stacktrace)
		}

		ec2VmPool, ok := vmPool.(ec2vm.Ec2VirtualMachinePool)
		if !ok {
			stacktrace := make([]byte, 4096)
			stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
			log.Errorf("Pool with id: %d is not an EC2 pool\n%s", ec2PoolId, stacktrace)
		}

		ec2Pool := ec2VmPool.Ec2VirtualMachineManager.Ec2Pool

		ec2VmPool.UpdateSettings(resources.Ec2Pool{ec2PoolId, ec2Pool.Name, accessKey, secretKey, username, baseAmiId,
			securityGroupId, vpcSubnetId, instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData, ec2Pool.Created, ec2Pool.IsDeleted})
	}
	var err error

	poolManager.subscriptionInfo = &subscriptionInfo{
		resourcesConnection: resourcesConnection,
		ec2Broker:           ec2Broker,
	}

	poolManager.subscriptionInfo.ec2PoolCreatedSubscriptionId, err = resourcesConnection.Pools.Subscription.SubscribeToEc2CreatedEvents(onEc2PoolCreated)
	if err != nil {
		poolManager.UnsubscribeFromEvents()
		return err
	}

	poolManager.subscriptionInfo.ec2PoolDeletedSubscriptionId, err = resourcesConnection.Pools.Subscription.SubscribeToEc2DeletedEvents(onEc2PoolDeleted)
	if err != nil {
		poolManager.UnsubscribeFromEvents()
		return err
	}

	poolManager.subscriptionInfo.ec2PoolSettingsUpdatedSubscriptionId, err = resourcesConnection.Pools.Subscription.SubscribeToEc2SettingsUpdatedEvents(onEc2PoolSettingsUpdated)
	if err != nil {
		poolManager.UnsubscribeFromEvents()
		return err
	}
	return nil
}

func (poolManager *PoolManager) UnsubscribeFromEvents() error {
	var err error

	subscriptionInfo := poolManager.subscriptionInfo
	if subscriptionInfo == nil {
		return fmt.Errorf("Pool manager not subscribed to events")
	}

	if subscriptionInfo.ec2PoolCreatedSubscriptionId != 0 {
		unsubscribeError := subscriptionInfo.resourcesConnection.Pools.Subscription.UnsubscribeFromEc2CreatedEvents(subscriptionInfo.ec2PoolCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	if subscriptionInfo.ec2PoolDeletedSubscriptionId != 0 {
		unsubscribeError := subscriptionInfo.resourcesConnection.Pools.Subscription.UnsubscribeFromEc2DeletedEvents(subscriptionInfo.ec2PoolDeletedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	if subscriptionInfo.ec2PoolSettingsUpdatedSubscriptionId != 0 {
		unsubscribeError := subscriptionInfo.resourcesConnection.Pools.Subscription.UnsubscribeFromEc2SettingsUpdatedEvents(subscriptionInfo.ec2PoolSettingsUpdatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	if err != nil {
		return err
	}

	poolManager.subscriptionInfo = nil
	return nil
}
