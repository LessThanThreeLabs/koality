package resources

import (
	"errors"
	"reflect"
	"sync"
)

type SubscriptionId uint64

type SubscriptionManager struct {
	idCounter          SubscriptionId
	idCounterMutex     sync.Mutex
	subscriptions      []EventSubscription
	subscriptionsMutex sync.Mutex
}

type EventSubscription struct {
	id                SubscriptionId
	reflectedFunction reflect.Value
}

func NewSubscriptionManager() (*SubscriptionManager, error) {
	return &SubscriptionManager{}, nil
}

func (subscriptionManager *SubscriptionManager) getNextId() SubscriptionId {
	subscriptionManager.idCounterMutex.Lock()
	defer subscriptionManager.idCounterMutex.Unlock()

	subscriptionManager.idCounter++
	return subscriptionManager.idCounter
}

func (subscriptionManager *SubscriptionManager) Add(function interface{}) (SubscriptionId, error) {
	reflectedFunction := reflect.ValueOf(function)
	subscription := EventSubscription{subscriptionManager.getNextId(), reflectedFunction}

	subscriptionManager.subscriptionsMutex.Lock()
	subscriptionManager.subscriptions = append(subscriptionManager.subscriptions, subscription)
	subscriptionManager.subscriptionsMutex.Unlock()

	return subscription.id, nil
}

func (subscriptionManager *SubscriptionManager) Remove(subscriptionId SubscriptionId) error {
	for index, subscription := range subscriptionManager.subscriptions {
		if subscription.id == subscriptionId {
			subscriptionManager.subscriptionsMutex.Lock()
			subscriptionManager.subscriptions = append(subscriptionManager.subscriptions[:index], subscriptionManager.subscriptions[index+1:]...)
			subscriptionManager.subscriptionsMutex.Unlock()
			return nil
		}
	}
	return errors.New("Unable to find subscription")
}

func (subscriptionManager *SubscriptionManager) Fire(params ...interface{}) {
	convertToReflectValues := func(params ...interface{}) []reflect.Value {
		values := make([]reflect.Value, 0, 1)
		for _, param := range params {
			values = append(values, reflect.ValueOf(param))
		}
		return values
	}

	reflectedParams := convertToReflectValues(params...)

	subscriptionManager.subscriptionsMutex.Lock()
	for _, subscription := range subscriptionManager.subscriptions {
		subscription.reflectedFunction.Call(reflectedParams)
	}
	subscriptionManager.subscriptionsMutex.Unlock()
}
