package pools

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	ec2CreatedSubscriptionManager         resources.SubscriptionManager
	ec2DeletedSubscriptionManager         resources.SubscriptionManager
	ec2SettingsUpdatedSubscriptionManager resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalPoolsSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToEc2CreatedEvents(updateHandler resources.PoolEc2CreatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.ec2CreatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromEc2CreatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.ec2CreatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireEc2CreatedEvent(ec2PoolId uint64) {
	subscriptionHandler.ec2CreatedSubscriptionManager.Fire(ec2PoolId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToEc2DeletedEvents(updateHandler resources.PoolEc2DeletedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.ec2DeletedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromEc2DeletedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.ec2DeletedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireEc2DeletedEvent(ec2PoolId uint64) {
	subscriptionHandler.ec2DeletedSubscriptionManager.Fire(ec2PoolId)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToEc2SettingsUpdatedEvents(updateHandler resources.PoolEc2SettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.ec2SettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromEc2SettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.ec2SettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireEc2SettingsUpdatedEvent(ec2PoolId uint64, accessKey, secretKey,
	username, baseAmiId, securityGroupId, vpcSubnetId, instanceType string,
	numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) {
	subscriptionHandler.ec2SettingsUpdatedSubscriptionManager.Fire(ec2PoolId, accessKey, secretKey,
		username, baseAmiId, securityGroupId, vpcSubnetId, instanceType,
		numReadyInstances, numMaxInstances, rootDriveSize, userData)
}
