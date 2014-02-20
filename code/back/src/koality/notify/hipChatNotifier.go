package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"koality/resources"
	"net/http"
	"strings"
)

type HipChatNotifier struct {
	resourcesConnection *resources.Connection
}

func NewHipChatNotifier(resourcesConnection *resources.Connection) BuildStatusNotifier {
	return &HipChatNotifier{resourcesConnection}
}

func (hipChatNotifier *HipChatNotifier) NotifyBuildStatus(build *resources.Build) error {
	hipChatSettings, err := hipChatNotifier.resourcesConnection.Settings.Read.GetHipChatSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		return nil
	} else if err != nil {
		return err
	}

	if build.Status != "failed" && !(build.Status == "passed" && hipChatSettings.NotifyOn == "all") {
		return nil
	}

	domainName, err := hipChatNotifier.resourcesConnection.Settings.Read.GetDomainName()
	if err != nil {
		return err
	}

	branchName := build.Ref
	if strings.HasPrefix(branchName, "refs/heads/") {
		branchName = branchName[len("refs/heads/"):]
	}
	viewBuildUrl := getBuildUri(domainName, build.RepositoryId, build.Id)

	repository, err := hipChatNotifier.resourcesConnection.Repositories.Read.Get(build.RepositoryId)
	if err != nil {
		return err
	}

	commitMessageHtml := "<code>" + html.EscapeString(build.Changeset.HeadMessage) + "</code>"
	message := htmlSanitizer.Replace(
		fmt.Sprintf("%s/%s - %s's change <a href=\"%s\">%s</a>\n%s",
			html.EscapeString(repository.Name), html.EscapeString(branchName),
			html.EscapeString(build.Changeset.HeadUsername), viewBuildUrl, build.Status,
			commitMessageHtml,
		),
	)

	messageColor := "gray"
	if build.Status == "passed" {
		messageColor = "green"
	} else if build.Status == "failed" {
		messageColor = "red"
	}

	requestBody := map[string]interface{}{
		"from":    "Koality",
		"format":  "html",
		"color":   messageColor,
		"message": message,
	}

	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	var notifyErr error

	for _, room := range hipChatSettings.Rooms {
		requestUrl := fmt.Sprintf("https://api.hipchat.com/v2/room/%s/notification?auth_token=%s", room, hipChatSettings.AuthenticationToken)
		response, err := http.Post(requestUrl, "application/json", bytes.NewReader(requestBodyJson))
		if err != nil {
			notifyErr = err
		} else if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
			// For some reason HipChat docs suggests that a 204 (NoContent) should be expected
			notifyErr = fmt.Errorf("Failed to post HipChat notification: %s", response.Status)
		}
	}

	return notifyErr
}
