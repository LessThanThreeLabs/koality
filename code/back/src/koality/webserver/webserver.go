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
	// fmt.Println("TODO: the cookie secret should be pulled from the database")
	cookieStoreAuthenticationKey := []byte("7e0ac5f56f9412d524eba19902b2345d")
	cookieStoreEncryptionKey := []byte("7e0ac5f56f9412d524eba19902b2345d")
	sessionStore := sessions.NewCookieStore(cookieStoreAuthenticationKey, cookieStoreEncryptionKey)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   2592000,
		HttpOnly: true,
		Secure:   true,
	}

	// TODO: can make it so that if not logged in, doesn't get to certain paths!!

	router := mux.NewRouter()
	http.Handle("/", wrapWithMiddleware(sessionStore, router))

	subrouter := router.PathPrefix("/app").Subrouter()
	subrouter.HandleFunc("/", SomeHandler)
	subrouter.HandleFunc("/2", SomeHandler2)

	address := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(address, nil)
}

func wrapWithMiddleware(sessionStore sessions.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, sessionName)
		session.Values["userId"] = uint64(17)

		context.Set(request, "session", session)

		err := session.Save(request, writer)
		if err != nil {
			fmt.Println(err) // TODO: should use the logger here
		}

		next.ServeHTTP(writer, request)

	})
}

func SomeHandler(writer http.ResponseWriter, request *http.Request) {
	session := context.Get(request, "session").(*sessions.Session)
	fmt.Fprintf(writer, "Hello there! - %v - %v", time.Now(), session.Values)
}

func SomeHandler2(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Hello again!! - %v", time.Now())
}
