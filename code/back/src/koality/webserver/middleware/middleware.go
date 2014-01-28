package middleware

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"koality/resources"
	"net/http"
)

func IsAdminWrapper(resourcesConnection *resources.Connection, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId, ok := context.Get(request, "userId").(uint64)
		if !ok || userId == 0 {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprint(writer, "Forbidden request, must be logged in")
			return
		}

		user, err := resourcesConnection.Users.Read.Get(userId)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
		} else if !user.IsAdmin || user.IsDeleted {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprint(writer, "Forbidden request, must be an admin")
		} else {
			next(writer, request)
		}
	}
}

func IsLoggedInWrapper(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId, ok := context.Get(request, "userId").(uint64)
		if !ok || userId == 0 {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprint(writer, "Forbidden request, must be logged in")
		} else {
			next(writer, request)
		}
	}
}

func IsLoggedOutWrapper(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId, ok := context.Get(request, "userId").(uint64)
		if ok && userId != 0 {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprint(writer, "Forbidden request, must be logged out")
		} else {
			next(writer, request)
		}
	}
}

func CheckCsrfTokenWraper(resourcesConnection *resources.Connection, router *mux.Router) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		router.ServeHTTP(writer, request)
	}
}

func HasApiKeyWrapper(resourcesConnection *resources.Connection, router *mux.Router) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		apiKeyToVerify := request.FormValue("apiKey")
		if apiKeyToVerify == "" {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprint(writer, "Forbidden request, must provide api key")
			return
		}

		apiKey, err := resourcesConnection.Settings.Read.GetApiKey()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
		} else if apiKey.Key != apiKeyToVerify {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprint(writer, "Forbidden request, invalid api key")
		} else {
			router.ServeHTTP(writer, request)
		}
	}
}
