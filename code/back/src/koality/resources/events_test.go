package resources

import (
	"fmt"
	"testing"
)

func TestEvents(test *testing.T) {
	eventHandler := new(UsersEventHandler)

	eventHandler.SubscribeToNameUpdateEvents(handleNameUpdateEvent)

	eventHandler.FireNameUpdateEvent(17, "Jordan", "Potter")
	eventHandler.FireNameUpdateEvent(19, "Jordan-2", "Potter-2")
}

func handleNameUpdateEvent(userId uint64, firstName, lastName string) {
	fmt.Println("Received event", userId, firstName, lastName)
}
