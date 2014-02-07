package google

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"koality/resources"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type BadAuthenticationError struct {
	message string
}

func (err BadAuthenticationError) Error() string {
	return err.message
}

func (googleHandler *GoogleHandler) handleOAuthToken(writer http.ResponseWriter, request *http.Request) {
	oAuthToken := request.FormValue("token")
	action := request.FormValue("action")

	if oAuthToken == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "No oAuth token provided")
		return
	}

	if action == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "No action provided")
	} else if action == "login" {
		googleHandler.processLoginAction(oAuthToken, writer, request)
	} else if action == "createAccount" {
		googleHandler.processCreateAccountAction(oAuthToken, writer, request)
	} else {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(writer, "Unknown action provided: %s", action)
	}
}

func (googleHandler *GoogleHandler) processLoginAction(oAuthToken string, writer http.ResponseWriter, request *http.Request) {
	user, err := googleHandler.handleLogin(oAuthToken)
	if _, ok := err.(BadAuthenticationError); ok {
		queryValues := url.Values{}
		queryValues.Set("googleLoginError", err.Error())
		http.Redirect(writer, request, "/login?"+queryValues.Encode(), http.StatusSeeOther)
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
	} else {
		googleHandler.login(user, writer, request)
		http.Redirect(writer, request, "/", http.StatusSeeOther)
	}
}

func (googleHandler *GoogleHandler) processCreateAccountAction(oAuthToken string, writer http.ResponseWriter, request *http.Request) {
	user, err := googleHandler.handleCreateAccount(oAuthToken)
	if _, ok := err.(BadAuthenticationError); ok {
		queryValues := url.Values{}
		queryValues.Set("googleCreateAccountError", err.Error())
		http.Redirect(writer, request, "/login?"+queryValues.Encode(), http.StatusSeeOther)
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	} else {
		googleHandler.login(user, writer, request)
		http.Redirect(writer, request, "/", http.StatusSeeOther)
	}
}

func (googleHandler *GoogleHandler) handleLogin(oAuthToken string) (*resources.User, error) {
	authenticationSettings, err := googleHandler.resourcesConnection.Settings.Read.GetAuthenticationSettings()
	if err != nil {
		return nil, err
	}

	if !authenticationSettings.GoogleLoginAllowed {
		return nil, BadAuthenticationError{"The administrator has disabled Google Account login"}
	}

	userInformation, err := googleHandler.getGoogleUserInformation(oAuthToken)
	if err != nil {
		return nil, err
	}

	if userInformation.Email == "" {
		return nil, BadAuthenticationError{"No email address provided"}
	} else if verifiedEmail, err := strconv.ParseBool(fmt.Sprint(userInformation.VerifiedEmail)); verifiedEmail && err != nil {
		return nil, BadAuthenticationError{"Email address not verified"}
	}

	user, err := googleHandler.resourcesConnection.Users.Read.GetByEmail(userInformation.Email)
	if _, ok := err.(resources.NoSuchUserError); ok {
		return nil, BadAuthenticationError{fmt.Sprintf("No user found with email address %s", userInformation.Email)}
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func (googleHandler *GoogleHandler) handleCreateAccount(oAuthToken string) (*resources.User, error) {
	authenticationSettings, err := googleHandler.resourcesConnection.Settings.Read.GetAuthenticationSettings()
	if err != nil {
		return nil, err
	}

	if !authenticationSettings.GoogleLoginAllowed {
		return nil, BadAuthenticationError{"The administrator has disabled Google Account login"}
	}

	userInformation, err := googleHandler.getGoogleUserInformation(oAuthToken)
	if err != nil {
		return nil, err
	}

	if userInformation.Email == "" {
		return nil, BadAuthenticationError{"No email address provided"}
	} else if verifiedEmail, err := strconv.ParseBool(fmt.Sprint(userInformation.VerifiedEmail)); verifiedEmail && err != nil {
		return nil, BadAuthenticationError{"Email address not verified"}
	} else if userInformation.GivenName == "" {
		return nil, BadAuthenticationError{"No first name provided"}
	} else if userInformation.FamilyName == "" {
		return nil, BadAuthenticationError{"No last name provided"}
	}

	emailAllowed := len(authenticationSettings.AllowedDomains) == 0

	for _, allowedDomain := range authenticationSettings.AllowedDomains {
		if strings.HasSuffix(userInformation.Email, "@"+allowedDomain) {
			emailAllowed = true
			break
		}
	}

	if !emailAllowed {
		return nil, BadAuthenticationError{"The email address you provided has not been authorized by the administrator"}
	}

	passwordHash := make([]byte, 32)
	passwordSalt := make([]byte, 16)

	if _, err = rand.Read(passwordHash); err != nil {
		return nil, err
	}

	if _, err = rand.Read(passwordSalt); err != nil {
		return nil, err
	}

	user, err := googleHandler.resourcesConnection.Users.Create.Create(userInformation.Email, userInformation.GivenName, userInformation.FamilyName, passwordHash, passwordSalt, false)
	if _, ok := err.(resources.UserAlreadyExistsError); ok {
		return nil, BadAuthenticationError{fmt.Sprintf("User already exists with email address %s", userInformation.Email)}
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func (googleHandler *GoogleHandler) getGoogleUserInformation(oAuthToken string) (*UserInformation, error) {
	httpClient := new(http.Client)

	requestQueryValues := url.Values{}
	requestQueryValues.Set("access_token", oAuthToken)
	requestUrl := "https://www.googleapis.com/oauth2/v2/userinfo?" + requestQueryValues.Encode()
	response, err := httpClient.Get(requestUrl)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		return nil, BadAuthenticationError{fmt.Sprintf("Received http status %d (%s) while trying to check oAuth token", response.StatusCode, response.Status)}
	}

	userInformation := new(UserInformation)

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(userInformation)
	if err != nil {
		return nil, err
	}
	return userInformation, nil
}

func (googleHandler *GoogleHandler) login(user *resources.User, writer http.ResponseWriter, request *http.Request) {
	session, _ := googleHandler.sessionStore.Get(request, googleHandler.sessionName)
	session.Values["userId"] = user.Id
	session.Save(request, writer)
}
