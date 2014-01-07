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

const (
	sessionName = "koality"
)

func Start(resourcesConnection *resources.Connection, port int) error {
	http.Handle("/", loadUserId(createSessionStore(), createRouter()))
	address := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(address, nil)
}

func createSessionStore() sessions.Store {
	fmt.Println("TODO: the cookie keys should be pulled from the database")
	cookieStoreAuthenticationKey := []byte("7e0ac5f56f9412d524eba19902b2345d")
	cookieStoreEncryptionKey := []byte("7e0ac5f56f9412d524eba19902b2345d")
	sessionStore := sessions.NewCookieStore(cookieStoreAuthenticationKey, cookieStoreEncryptionKey)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   2592000,
		HttpOnly: true,
		Secure:   true,
	}
	return sessionStore
}

func createRouter() *mux.Router {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/app").Subrouter()

	userSubrouter := subrouter.PathPrefix("/users").MatcherFunc(isLoggedIn).Subrouter()
	userSubrouter.HandleFunc("/blah", SomeUserHandler)

	subrouter.HandleFunc("/", SomeHandler)
	subrouter.HandleFunc("/2", SomeHandler2)

	return router
}

func loadUserId(sessionStore sessions.Store, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, sessionName)
		// session.Values["userId"] = uint64(17)
		context.Set(request, "userId", session.Values["userId"])

		// err := session.Save(request, writer)
		// if err != nil {
		// 	fmt.Println(err) // TODO: should use the logger here
		// }

		handler.ServeHTTP(writer, request)
	})
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
	fmt.Fprintf(writer, "Some handler - %v - %v", time.Now())
}

func SomeHandler2(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Some handler 2 - %v", time.Now())
}
