package events

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"koality/webserver/middleware"
	"time"
)

type userCreatedEventData struct {
	Id             uint64     `json:"id"`
	Email          string     `json:"email"`
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	IsAdmin        bool       `json:"isAdmin"`
	Created        *time.Time `json:"created,omitempty"`
	IsDeleted      bool       `json:"isDeleted"`
	HasGitHubOAuth bool       `json:"hasGitHubOAuth"`
}

type userDeletedEventData struct {
	Id uint64 `json:"id"`
}

type userNameUpdatedEventData struct {
	Id        uint64 `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type userAdminUpdatedEventData struct {
	Id      uint64 `json:"id"`
	IsAdmin bool   `json:"isAdmin"`
}

type userSshKeyAddedEventData struct {
	Id       uint64 `json:"id"`
	SshKeyId uint64 `json:"keyId"`
}

type userSshKeyRemovedEventData struct {
	Id       uint64 `json:"id"`
	SshKeyId uint64 `json:"keyId"`
}

func (eventsHandler *EventsHandler) wireUserAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/created/subscribe",
		middleware.IsLoggedInWrapper(
			eventsHandler.createSubscription(userCreatedSubscriptions, false, false))).
		Methods("POST")
	subrouter.HandleFunc("/deleted/subscribe",
		middleware.IsLoggedInWrapper(
			eventsHandler.createSubscription(userDeletedSubscriptions, false, false))).
		Methods("POST")
	subrouter.HandleFunc("/name/subscribe",
		middleware.IsLoggedInWrapper(
			eventsHandler.createSubscription(userNameUpdatedSubscriptions, false, false))).
		Methods("POST")
	subrouter.HandleFunc("/admin/subscribe",
		middleware.IsLoggedInWrapper(
			eventsHandler.createSubscription(userAdminUpdatedSubscriptions, false, false))).
		Methods("POST")
	subrouter.HandleFunc("/sshKeyAdded/subscribe",
		middleware.IsLoggedInWrapper(
			eventsHandler.createSubscription(userSshKeyAddedSubscriptions, false, false))).
		Methods("POST")
	subrouter.HandleFunc("/sshKeyRemoved/subscribe",
		middleware.IsLoggedInWrapper(
			eventsHandler.createSubscription(userSshKeyRemovedSubscriptions, false, false))).
		Methods("POST")

	subrouter.HandleFunc("/created/{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userCreatedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/deleted/{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userDeletedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/name/{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userNameUpdatedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/admin/{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userAdminUpdatedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/sshKeyAdded/{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userSshKeyAddedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/sshKeyRemoved/{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userSshKeyRemovedSubscriptions))).
		Methods("DELETE")
}

func (eventsHandler *EventsHandler) listenForUserEvents() error {
	_, err := eventsHandler.resourcesConnection.Users.Subscription.SubscribeToCreatedEvents(eventsHandler.handleUserCreatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Users.Subscription.SubscribeToDeletedEvents(eventsHandler.handleUserDeletedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Users.Subscription.SubscribeToNameUpdatedEvents(eventsHandler.handleNameUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Users.Subscription.SubscribeToAdminUpdatedEvents(eventsHandler.handleAdminUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Users.Subscription.SubscribeToSshKeyAddedEvents(eventsHandler.handleSshKeyAddedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Users.Subscription.SubscribeToSshKeyRemovedEvents(eventsHandler.handleSshKeyRemovedEvent)
	if err != nil {
		return err
	}

	return nil
}

func (eventsHandler *EventsHandler) handleUserCreatedEvent(user *resources.User) {
	data := userCreatedEventData{
		Id:             user.Id,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		IsAdmin:        user.IsAdmin,
		Created:        user.Created,
		IsDeleted:      user.IsDeleted,
		HasGitHubOAuth: user.GitHubOAuth != "",
	}
	eventsHandler.handleEvent(userCreatedSubscriptions, user.Id, data)
}

func (eventsHandler *EventsHandler) handleUserDeletedEvent(userId uint64) {
	data := userDeletedEventData{userId}
	eventsHandler.handleEvent(userDeletedSubscriptions, userId, data)
}

func (eventsHandler *EventsHandler) handleNameUpdatedEvent(userId uint64, firstName, lastName string) {
	data := userNameUpdatedEventData{userId, firstName, lastName}
	eventsHandler.handleEvent(userNameUpdatedSubscriptions, userId, data)
}

func (eventsHandler *EventsHandler) handleAdminUpdatedEvent(userId uint64, admin bool) {
	data := userAdminUpdatedEventData{userId, admin}
	eventsHandler.handleEvent(userAdminUpdatedSubscriptions, userId, data)
}

func (eventsHandler *EventsHandler) handleSshKeyAddedEvent(userId, sshKeyId uint64) {
	data := userSshKeyAddedEventData{userId, sshKeyId}
	eventsHandler.handleEvent(userSshKeyAddedSubscriptions, userId, data)
}

func (eventsHandler *EventsHandler) handleSshKeyRemovedEvent(userId, sshKeyId uint64) {
	data := userSshKeyRemovedEventData{userId, sshKeyId}
	eventsHandler.handleEvent(userSshKeyRemovedSubscriptions, userId, data)
}
