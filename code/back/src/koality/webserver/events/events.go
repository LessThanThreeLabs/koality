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

	return &eventsHandler, nil
}

func (eventsHandler *EventsHandler) WireAppSubroutes(subrouter *mux.Router) {
	usersSubrouter := subrouter.PathPrefix("/users").Subrouter()
	eventsHandler.wireUserAppSubroutes(usersSubrouter)
}

func (eventsHandler *EventsHandler) getNextSubscriptionId() uint64 {
	eventsHandler.subscriptionIdMutex.Lock()
	defer eventsHandler.subscriptionIdMutex.Unlock()
	eventsHandler.subscriptionIdCounter++
	return eventsHandler.subscriptionIdCounter
}

func (eventsHandler *EventsHandler) createSubscription(subscriptionType subscriptionType, mustBeAdmin bool, mustBeSelf bool) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId := context.Get(request, "userId").(uint64)

		subscriptionRequestData := new(subscriptionRequestData)
		defer request.Body.Close()
		if err := json.NewDecoder(request.Body).Decode(subscriptionRequestData); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		websocketId := eventsHandler.websocketsManager.GetId(userId, subscriptionRequestData.ConnectionId)
		subscription := subscription{eventsHandler.getNextSubscriptionId(), userId, websocketId,
			subscriptionRequestData.AllResources, subscriptionRequestData.ResourceId}

		if mustBeAdmin {
			user, err := eventsHandler.resourcesConnection.Users.Read.Get(userId)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			} else if !user.IsAdmin {
				http.Error(writer, "Forbidden request, must be an admin", http.StatusForbidden)
				return
			}
		}

		if mustBeSelf {
			if userId != subscriptionRequestData.ResourceId || subscriptionRequestData.AllResources {
				http.Error(writer, "Forbidden request, can only subscribe to events for self", http.StatusForbidden)
				return
			}
		}

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
