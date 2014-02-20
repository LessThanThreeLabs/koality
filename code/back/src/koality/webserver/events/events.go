package events

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"koality/resources"
	"koality/webserver/websockets"
	"net/http"
	"strconv"
	"sync"
)

type subscriptionRequestData struct {
	ConnectionId uint64 `json:"connectionId"`
	AllResources bool   `json:"allResources"`
	ResourceId   uint64 `json:"resourceId"`
}

type eventData struct {
	SubscriptionId uint64      `json:"subscriptionId"`
	Data           interface{} `json:"data"`
}

type EventsHandler struct {
	resourcesConnection   *resources.Connection
	websocketsManager     *websockets.WebsocketsManager
	subscriptionIdCounter uint64
	subscriptionIdMutex   sync.Mutex
	subscriptions         map[subscriptionType][]subscription
	subscriptionsRWMutex  sync.RWMutex
}

func New(resourcesConnection *resources.Connection, websocketsManager *websockets.WebsocketsManager) (*EventsHandler, error) {
	eventsHandler := EventsHandler{
		resourcesConnection: resourcesConnection,
		websocketsManager:   websocketsManager,
		subscriptions:       make(map[subscriptionType][]subscription),
	}

	eventsHandler.listenForUserEvents()
	eventsHandler.listenForSettingsEvents()

	return &eventsHandler, nil
}

func (eventsHandler *EventsHandler) WireAppSubroutes(subrouter *mux.Router) {
	usersSubrouter := subrouter.PathPrefix("/users").Subrouter()
	eventsHandler.wireUserAppSubroutes(usersSubrouter)

	settingsSubrouter := subrouter.PathPrefix("/settings").Subrouter()
	eventsHandler.wireSettingsAppSubroutes(settingsSubrouter)
}

func (eventsHandler *EventsHandler) getNextSubscriptionId() uint64 {
	eventsHandler.subscriptionIdMutex.Lock()
	defer eventsHandler.subscriptionIdMutex.Unlock()
	eventsHandler.subscriptionIdCounter++
	return eventsHandler.subscriptionIdCounter
}

func (eventsHandler *EventsHandler) createSubscription(subscriptionType subscriptionType, mustBeAll, mustBeSelf bool) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId := context.Get(request, "userId").(uint64)

		subscriptionRequestData := new(subscriptionRequestData)
		defer request.Body.Close()
		if err := json.NewDecoder(request.Body).Decode(subscriptionRequestData); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if subscriptionRequestData.AllResources {
			subscriptionRequestData.ResourceId = 0
		}

		if subscriptionRequestData.ResourceId < 0 {
			http.Error(writer, "Forbidden request, must specify valid resource id", http.StatusForbidden)
			return
		}

		if mustBeAll && !subscriptionRequestData.AllResources {
			http.Error(writer, "Forbidden request, must subscribe to all events for resource", http.StatusForbidden)
			return
		}

		if mustBeSelf && (subscriptionRequestData.AllResources || userId != subscriptionRequestData.ResourceId) {
			http.Error(writer, "Forbidden request, can only subscribe to events for self", http.StatusForbidden)
			return
		}

		websocketId := eventsHandler.websocketsManager.GetId(userId, subscriptionRequestData.ConnectionId)
		subscription := subscription{eventsHandler.getNextSubscriptionId(), userId, websocketId,
			subscriptionRequestData.AllResources, subscriptionRequestData.ResourceId}

		err := eventsHandler.addToSubscriptions(subscriptionType, subscription)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(writer, subscription.id)
	}
}

func (eventsHandler *EventsHandler) deleteSubscription(subscriptionType subscriptionType) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId := context.Get(request, "userId").(uint64)

		subscriptionIdString := mux.Vars(request)["subscriptionId"]
		subscriptionId, err := strconv.ParseUint(subscriptionIdString, 10, 64)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Unable to parse subscriptionId: %v", err), http.StatusInternalServerError)
			return
		}

		err = eventsHandler.removeFromSubscriptions(subscriptionType, userId, subscriptionId)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(writer, "ok")
	}
}
