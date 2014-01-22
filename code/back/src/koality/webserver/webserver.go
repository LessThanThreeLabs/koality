package webserver

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/repositorymanager"
	"koality/resources"
	"koality/webserver/users"
	"net/http"
)

type Webserver struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
	sessionName         string
	address             string
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager, port int) (*Webserver, error) {
	address := fmt.Sprintf(":%d", port)
	return &Webserver{resourcesConnection, repositoryManager, "koality", address}, nil
}

func (webserver *Webserver) Start() error {
	sessionStore, err := webserver.createSessionStore()
	if err != nil {
		return err
	}

	router, err := webserver.createRouter()
	if err != nil {
		return err
	}

	loadUserIdRouter := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, webserver.sessionName)
		// context.Set(request, "userId", session.Values["userId"])
		var _ = session
		context.Set(request, "userId", uint64(1001))
		router.ServeHTTP(writer, request)
	})

	http.Handle("/", loadUserIdRouter)
	return http.ListenAndServe(webserver.address, nil)
}

func (webserver *Webserver) createSessionStore() (sessions.Store, error) {
	cookieStoreKeys, err := webserver.resourcesConnection.Settings.Read.GetCookieStoreKeys()
	if err != nil {
		return nil, err
	}

	sessionStore := sessions.NewCookieStore(cookieStoreKeys.Authentication, cookieStoreKeys.Encryption)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   2592000,
		HttpOnly: true,
		Secure:   true,
	}
	return sessionStore, nil
}

func (webserver *Webserver) createRouter() (*mux.Router, error) {
	router := mux.NewRouter()
	apiSubrouter := router.PathPrefix("/api").Subrouter()
	handleApiSubroute(apiSubrouter, webserver.resourcesConnection, webserver.repositoryManager)

	appSubrouter := router.PathPrefix("/app").Subrouter()

	usersHandler, err := users.New(webserver.resourcesConnection, webserver.repositoryManager)
	if err != nil {
		return nil, err
	}

	usersSubrouter := appSubrouter.PathPrefix("/users").MatcherFunc(isLoggedIn).Subrouter()
	usersHandler.WireSubroutes(usersSubrouter)

	// TODO: create and handle more subroutes

	return router, nil
}

func isLoggedIn(request *http.Request, match *mux.RouteMatch) bool {
	fmt.Println("Temporarily assuming logged in")
	return true

	// userId, ok := context.Get(request, "userId").(uint64)
	// if !ok {
	// 	return false
	// }
	// return userId != 0
}
