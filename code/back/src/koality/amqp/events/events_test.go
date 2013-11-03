package events

import (
	"fmt"
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

func TestSmallEvents(testing *testing.T) {
	publisher := NewPublisher(route)
	subscriber := NewSubscriber(route, new(EventHandler))
	shortEvent := NewEvent("resource", "test event name", shortString, shortString, shortString)
	sendEvents(publisher, subscriber, shortEvent, 10000, 1000)
}

func sendEvents(eventPublisher *Publisher, eventSubscriber *Subscriber, event *Event, numEvents, maxConcurrentEvents int) {
	semaphore := make(chan bool, maxConcurrentEvents)
	completed := make(chan bool, maxConcurrentEvents)

	go func() {
		for index := 0; index < numEvents; index++ {
			semaphore <- true
			go func() {
				err := eventPublisher.SendEvent(event)
				if err != nil {
					panic(err)
				}

				completed <- true
				<-semaphore
			}()
		}
	}()

	for index := 0; index < numEvents; index++ {
		<-completed
	}
}

type EventHandler int

func (eventHandler *EventHandler) HandleEvent(event *Event) {
	fmt.Println(event)
}
