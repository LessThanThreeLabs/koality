package accounts

import (
	"fmt"
	"koality/resources"
	"net/http"
	"net/url"
)

func (accountsHandler *AccountsHandler) getGoogleLoginRedirect(writer http.ResponseWriter, request *http.Request) {
	domainName, err := accountsHandler.resourcesConnection.Settings.Read.GetDomainName()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		domainName = "127.0.0.1:10443"
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	queryValues := url.Values{}
	queryValues.Set("redirectUri", fmt.Sprintf("https://%s/oAuth/google/token", domainName))
	queryValues.Set("action", "login")
	loginRedirectUri := fmt.Sprintf("https://koalitycode.com/google/authenticate?%s", queryValues.Encode())
	fmt.Fprint(writer, loginRedirectUri)
}
