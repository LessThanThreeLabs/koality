package webserver

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/resources"
	"net/http"
	"time"
)

type Webserver struct {
	resourcesConnection *resources.Connection
	sessionName         string
	address             string
}

func New(resourcesConnection *resources.Connection, port int) (*Webserver, error) {
	address := fmt.Sprintf(":%d", port)
	return &Webserver{resourcesConnection, "koality", address}, nil
}

func (webserver *Webserver) Start() error {
	sessionStore := webserver.createSessionStore()
	router := webserver.createRouter()
	loadUserIdRouter := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, webserver.sessionName)
		context.Set(request, "userId", session.Values["userId"])
		router.ServeHTTP(writer, request)
	})

	http.Handle("/", loadUserIdRouter)
	return http.ListenAndServe(webserver.address, nil)
}

func (webserver *Webserver) createSessionStore() sessions.Store {
	cookieStoreKeys, err := webserver.resourcesConnection.Settings.Read.GetCookieStoreKeys()
	if err != nil {
		panic(err)
	}

	sessionStore := sessions.NewCookieStore(cookieStoreKeys.Authentication, cookieStoreKeys.Encryption)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   2592000,
		HttpOnly: true,
		Secure:   true,
	}
	return sessionStore
}

func (webserver *Webserver) createRouter() *mux.Router {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/app").Subrouter()

	userSubrouter := subrouter.PathPrefix("/users").MatcherFunc(isLoggedIn).Subrouter()
	userSubrouter.HandleFunc("/blah", SomeUserHandler)

	subrouter.HandleFunc("/", SomeHandler)
	subrouter.HandleFunc("/2", SomeHandler2)

	return router
}

func isLoggedIn(request *http.Request, match *mux.RouteMatch) bool {
	userId, ok := context.Get(request, "userId").(uint64)
	if !ok {
		return false
	}
	return userId != 0
}

func SomeUserHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Some user handler - %v", time.Now())
}

func SomeHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Some handler - %v", time.Now())
}

func SomeHandler2(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Some handler 2 - %v", time.Now())
}
