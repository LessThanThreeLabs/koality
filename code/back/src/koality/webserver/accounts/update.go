package accounts

import (
	"bytes"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"net/http"
	"strconv"
)

func (accountsHandler *AccountsHandler) Login(writer http.ResponseWriter, request *http.Request) {
	_, ok := context.Get(request, "userId").(uint64)
	if ok {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Already logged in")
		return
	}

	email := request.PostFormValue("email")
	password := request.PostFormValue("password")
	rememberMeString := request.PostFormValue("rememberMe")
	rememberMe, err := strconv.ParseBool(rememberMeString)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
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

	session, err := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	session.Values["userId"] = user.Id
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   accountsHandler.getMaxSessionAge(rememberMe),
		HttpOnly: true,
		Secure:   true,
	}
	session.Save(request, writer)

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) getMaxSessionAge(rememberMe bool) int {
	if rememberMe {
		return rememberMeDuration
	} else {
		return 0
	}
}
