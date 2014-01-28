package accounts

import (
	"bytes"
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
	"strconv"
)

func (accountsHandler *AccountsHandler) getMaxSessionAge(rememberMe bool) int {
	if rememberMe {
		return rememberMeDuration
	} else {
		return 0
	}
}

func (accountsHandler *AccountsHandler) Login(writer http.ResponseWriter, request *http.Request) {
	email := request.PostFormValue("email")
	password := request.PostFormValue("password")
	rememberMeString := request.PostFormValue("rememberMe")
	rememberMe, err := strconv.ParseBool(rememberMeString)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse rememberMe: %v", err)
		return
	}

	user, err := accountsHandler.resourcesConnection.Users.Read.GetByEmail(email)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	passwordHash, err := accountsHandler.passwordHasher.ComputeHash(password, user.PasswordSalt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if bytes.Compare(passwordHash, user.PasswordHash) != 0 {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Invalid password")
		return
	}

	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	session.Values["userId"] = user.Id
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   accountsHandler.getMaxSessionAge(rememberMe),
		HttpOnly: true,
		Secure:   true,
	}
	session.Save(request, writer)

	http.Redirect(writer, request, "/", http.StatusFound)
}

func (accountsHandler *AccountsHandler) Logout(writer http.ResponseWriter, request *http.Request) {
	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	session.Options = &sessions.Options{
		MaxAge: -1,
	}
	session.Save(request, writer)

	http.Redirect(writer, request, "/", http.StatusFound)
}
