package events

import (
	"strings"
	"testing"
)

const (
	route = "test-route"
)

var (
	shortString = makeTestString("short", 10)    // =60B
	longString  = makeTestString("long", 200000) // =1MB
)

func makeTestString(text string, numRepeats int) string {
	stringsArray := make([]string, numRepeats)
	for index := 0; index < numRepeats; index++ {
		stringsArray[index] = text
	}
	return strings.Join(stringsArray, " ")
}

func TestSubscibeAndUnsubscribe(testing *testing.T) {
	publisher := NewPublisher(route)
	subscriber := NewSubscriber(route)
	event := NewEvent("resource", "test event name", shortString, shortString, shortString)

	firstSubscription := subscriber.Subscribe()
	err := publisher.Publish(event)
	if err != nil {
		testing.Error(err)
	}
	<-firstSubscription

	secondSubscription := subscriber.Subscribe()
	publisher.Publish(event)
	if err != nil {
		testing.Error(err)
	}
	<-firstSubscription
	<-secondSubscription

	err = subscriber.Unsubscribe(firstSubscription)
	if err != nil {
		testing.Error(err)
	}
	publisher.Publish(event)
	if err != nil {
		testing.Error(err)
	}
	<-secondSubscription
	_, ok := <-firstSubscription
	if ok {
		testing.Error("Event passed to unsubscribed subscription")
	}
}

func TestSmallEvents(testing *testing.T) {
	publisher := NewPublisher(route)
	subscriber := NewSubscriber(route)
	shortEvent := NewEvent("resource", "test event name", shortString, shortString, shortString)
	sendEvents(testing, publisher, subscriber, shortEvent, 10000)
}

func TestLargeEvents(testing *testing.T) {
	publisher := NewPublisher(route)
	subscriber := NewSubscriber(route)
	longEvent := NewEvent("resource", "test event name", longString, longString, longString)
	sendEvents(testing, publisher, subscriber, longEvent, 50)
}

func sendEvents(testing *testing.T, eventPublisher *Publisher, eventSubscriber *Subscriber, event *Event, numEvents int) {
	events := eventSubscriber.Subscribe()

	go func() {
		for index := 0; index < numEvents; index++ {
			go func() {
				err := eventPublisher.Publish(event)
				if err != nil {
					testing.Error(err)
				}
			}()
		}
	}()

	for index := 0; index < numEvents; index++ {
		<-events
	}
}
