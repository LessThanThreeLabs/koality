package accounts

import (
	"fmt"
	"net/http"
)

func (accountsHandler *AccountsHandler) create(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "hello")
}

func (accountsHandler *AccountsHandler) createWithGitHub(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "hello")
}
