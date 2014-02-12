package github

import (
	"fmt"
	"github.com/gorilla/context"
	"net/http"
)

func (gitHubHandler *GitHubHandler) getConnectUri(writer http.ResponseWriter, request *http.Request) {
	action := request.FormValue("action")
	if action == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "No action provided")
		return
	}

	authorizationUrl, err := gitHubHandler.gitHubConnection.GetAuthorizationUrl(action)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprint(writer, fmt.Sprintf(`{"redirectUri": "%s"}`, authorizationUrl))
}

func (gitHubHandler *GitHubHandler) handleOAuthToken(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	oAuthToken := request.FormValue("token")
	action := request.FormValue("action")

	if oAuthToken == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "No oAuth token provided")
		return
	} else if action == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "No action provided")
		return
	}

	user, err := gitHubHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = gitHubHandler.resourcesConnection.Users.Update.SetGitHubOAuth(user.Id, oAuthToken)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	var redirectUrl string
	switch action {
	case "sshKeys":
		redirectUrl = "/account?view=sshKeys&importGitHubKeys"
	case "addRepository":
		redirectUrl = "/admin?view=repositories&addGitHubRepository"
	case "editRepository":
		redirectUrl = "/admin?view=repositories&gitHubAuthenticated"
	default:
		redirectUrl = "/"
	}
	http.Redirect(writer, request, redirectUrl, http.StatusSeeOther)
}
