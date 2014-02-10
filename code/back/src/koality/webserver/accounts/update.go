package accounts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (accountsHandler *AccountsHandler) getMaxSessionAge(rememberMe bool) int {
	if rememberMe {
		return rememberMeDuration
	} else {
		return 0
	}
}

func (accountsHandler *AccountsHandler) login(writer http.ResponseWriter, request *http.Request) {
	loginRequestData := new(loginRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(loginRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	user, err := accountsHandler.resourcesConnection.Users.Read.GetByEmail(loginRequestData.Email)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	passwordHash, err := accountsHandler.passwordHasher.ComputeHash(loginRequestData.Password, user.PasswordSalt)
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
	session.Options.MaxAge = accountsHandler.getMaxSessionAge(loginRequestData.RememberMe)
	session.Save(request, writer)

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) logout(writer http.ResponseWriter, request *http.Request) {
	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	session.Options.MaxAge = -1
	session.Save(request, writer)

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) resetPassword(writer http.ResponseWriter, request *http.Request) {
	emailBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "need to reset password for "+string(emailBytes))
}
