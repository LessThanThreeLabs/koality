package util

import (
	"github.com/gorilla/sessions"
	"net/http"
)

const (
	rememberMeDuration = 2592000
)

func Login(userId uint64, rememberMe bool, session *sessions.Session, writer http.ResponseWriter, request *http.Request) {
	session.Values["userId"] = userId
	session.Options.MaxAge = getMaxSessionAge(rememberMe)
	session.Save(request, writer)
}

func Logout(session *sessions.Session, writer http.ResponseWriter, request *http.Request) {
	session.Options.MaxAge = -1
	session.Save(request, writer)
}

func getMaxSessionAge(rememberMe bool) int {
	if rememberMe {
		return rememberMeDuration
	} else {
		return 0
	}
}
