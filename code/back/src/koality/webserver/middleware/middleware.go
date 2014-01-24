package middleware

import (
	"fmt"
	"github.com/gorilla/context"
	"koality/resources"
	"net/http"
)

func IsAdminWrapper(resourcesConnection *resources.Connection, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId := context.Get(request, "userId").(uint64)
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
