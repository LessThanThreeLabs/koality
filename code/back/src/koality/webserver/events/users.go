package events

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

	subrouter.HandleFunc("/created/#{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userCreatedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/deleted/#{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userDeletedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/name/#{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userNameUpdatedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/admin/#{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userAdminUpdatedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/sshKeyAdded/#{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userSshKeyAddedSubscriptions))).
		Methods("DELETE")
	subrouter.HandleFunc("/sshKeyRemoved/#{subscriptionId:[0-9]+}",
		middleware.IsLoggedInWrapper(
			eventsHandler.deleteSubscription(userSshKeyRemovedSubscriptions))).
		Methods("DELETE")
}
