package accounts

import (
	"fmt"
	"net/http"
	"net/url"
)

func (accountsHandler *AccountsHandler) getGoogleLoginRedirect(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Need to actually read off domain name")
	domainName := "127.0.0.1:10443"

	queryValues := url.Values{}
	queryValues.Set("redirectUri", fmt.Sprintf("https://%s/google/token", domainName))
	queryValues.Set("action", "login")
	loginRedirectUri := fmt.Sprintf("https://koalitycode.com/google/authenticate?%s", queryValues.Encode())
	fmt.Fprint(writer, loginRedirectUri)
}
