package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

func (gitHubHandler *GitHubHandler) verifyChange(writer http.ResponseWriter, request *http.Request) {
	eventType := request.Header.Get("X-GitHub-Event")
	payload := []byte(request.PostFormValue("payload"))

	switch eventType {
	case "push":
		var pushPayload PushHookPayload
		if err := json.Unmarshal(payload, &pushPayload); err == nil && pushPayload.Ref != "" {
			gitHubHandler.handlePush(pushPayload, writer, request)
			return
		}
	case "pull_request":
		var pullRequestPayload PullRequestHookPayload
		if err := json.Unmarshal(payload, &pullRequestPayload); err == nil && pullRequestPayload.Action != "" {
			gitHubHandler.handlePullRequest(pullRequestPayload, writer, request)
			return
		}
	case "ping":
		fmt.Fprint(writer, "pong")
		return
	default:
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(writer, "Unexpected hook type received: %s", eventType)
		return
	}

	writer.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(writer, "Malformed %s hook payload received", eventType)
}

func (gitHubHandler *GitHubHandler) handlePullRequest(pullRequestPayload PullRequestHookPayload, writer http.ResponseWriter, request *http.Request) {
	if pullRequestPayload.Action == "closed" {
		fmt.Fprint(writer, "ok")
		return
	}

	gitHubRepository := pullRequestPayload.PullRequest.Base.Repository
	gitHubHandler.triggerBuild(gitHubRepository.Owner.Login, gitHubRepository.Name,
		pullRequestPayload.PullRequest.Head.Sha, pullRequestPayload.PullRequest.Base.Sha,
		pullRequestPayload.PullRequest.Head.Ref, writer, request)
}

func (gitHubHandler *GitHubHandler) handlePush(pushPayload PushHookPayload, writer http.ResponseWriter, request *http.Request) {
	if pushPayload.After == "" {
		// Branch deleted
		fmt.Fprint(writer, "ok")
		return
	}
	gitHubHandler.triggerBuild(pushPayload.Repository.Owner.Name, pushPayload.Repository.Name,
		pushPayload.After, pushPayload.Before, pushPayload.Ref,
		writer, request)
}

func (gitHubHandler *GitHubHandler) triggerBuild(repositoryOwner, repositoryName, headSha, baseSha, ref string, writer http.ResponseWriter, request *http.Request) {
	repository, err := gitHubHandler.resourcesConnection.Repositories.Read.GetByGitHubInfo(repositoryOwner, repositoryName)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(writer, "Repository %s/%s not found", repositoryOwner, repositoryName)
		return
	}

	if ok := gitHubHandler.checkSignature([]byte(request.PostFormValue("payload")), repository.GitHub.HookSecret, request.Header.Get("x-hub-signature")); !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(writer, "Hook secret does not match")
		return
	}

	headSha, headMessage, headUsername, headEmail, err := gitHubHandler.repositoryManager.GetCommitAttributes(repository, headSha)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if err = gitHubHandler.repositoryManager.StorePending(repository, headSha); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	_, err = gitHubHandler.resourcesConnection.Builds.Create.Create(repository.Id, headSha, baseSha, headMessage, headUsername, headEmail, nil, headEmail, ref, true, false)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

func (gitHubHandler *GitHubHandler) checkSignature(payload []byte, hookSecret, signature string) bool {
	hasher := hmac.New(sha1.New, []byte(hookSecret))
	hasher.Write(payload)
	checksum := hasher.Sum(nil)
	hexString := hex.EncodeToString(checksum)
	return hexString == signature
}
