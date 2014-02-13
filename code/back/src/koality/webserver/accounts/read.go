package accounts

import (
	"fmt"
	"koality/resources"
	"net/http"
	"net/url"
	"strconv"
)

func (accountsHandler *AccountsHandler) getGoogleLoginRedirect(writer http.ResponseWriter, request *http.Request) {
	accountsHandler.getGoogleRedirect(writer, request, "login")
}

func (accountsHandler *AccountsHandler) getGoogleCreateAccountRedirect(writer http.ResponseWriter, request *http.Request) {
	accountsHandler.getGoogleRedirect(writer, request, "createAccount")
}

func (accountsHandler *AccountsHandler) getGoogleRedirect(writer http.ResponseWriter, request *http.Request, action string) {
	rememberMeString := request.URL.Query().Get("rememberMe")
	rememberMe, _ := strconv.ParseBool(rememberMeString)

	domainName, err := accountsHandler.resourcesConnection.Settings.Read.GetDomainName()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		domainName = "127.0.0.1:10443"
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	session.Values["rememberMe"] = rememberMe
	session.Save(request, writer)

	queryValues := url.Values{}
	queryValues.Set("redirectUri", fmt.Sprintf("https://%s/oAuth/google/token", domainName))
	queryValues.Set("action", action)
	loginRedirectUri := fmt.Sprintf("https://koalitycode.com/google/authenticate?%s", queryValues.Encode())
	fmt.Fprint(writer, loginRedirectUri)
}
