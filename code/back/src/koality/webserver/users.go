package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"koality/resources"
	"net/http"
	"time"
)

func handleUsersSubroute(subrouter *mux.Router, usersConnection *resources.UsersHandler) {
	subrouter.HandleFunc("/", SomeHandler)
	subrouter.HandleFunc("/2", SomeHandler2)
}

func SomeHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Some handler - %v", time.Now())
}

func SomeHandler2(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Some handler 2 - %v", time.Now())
}
