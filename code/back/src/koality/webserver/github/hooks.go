package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"koality/resources"
	"net/http"
)

func (gitHubHandler *GitHubHandler) verifyChange(writer http.ResponseWriter, request *http.Request) {
	payload := []byte(request.PostFormValue("payload"))

	var pullRequestPayload PullRequestHookPayload
	var pushPayload PushHookPayload

	if err := json.Unmarshal(payload, &pullRequestPayload); err == nil && pullRequestPayload.Action != "" {
		gitHubHandler.handlePullRequest(pullRequestPayload, writer, request)
	} else if err := json.Unmarshal(payload, &pushPayload); err == nil && pushPayload.Ref != "" {
		gitHubHandler.handlePush(pushPayload, writer, request)
	} else {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Unexpected hook received")
	}
}

func (gitHubHandler *GitHubHandler) handlePullRequest(pullRequestPayload PullRequestHookPayload, writer http.ResponseWriter, request *http.Request) {
	if pullRequestPayload.Action == "closed" {
		fmt.Fprint(writer, "ok")
		return
	}
	gitHubRepository := pullRequestPayload.PullRequest.Base.Repository
	gitHubHandler.triggerVerification(gitHubRepository.Owner.Login, gitHubRepository.Name,
		pullRequestPayload.PullRequest.Head.Sha, pullRequestPayload.PullRequest.Base.Sha,
		writer, request)
}

func (gitHubHandler *GitHubHandler) handlePush(pushPayload PushHookPayload, writer http.ResponseWriter, request *http.Request) {
	if pushPayload.After == "" {
		// Branch deleted
		fmt.Fprint(writer, "ok")
		return
	}
	gitHubHandler.triggerVerification(pushPayload.Repository.Owner.Name, pushPayload.Repository.Name,
		pushPayload.After, pushPayload.Before,
		writer, request)
}

func (gitHubHandler *GitHubHandler) triggerVerification(repositoryOwner, repositoryName, headSha, baseSha string, writer http.ResponseWriter, request *http.Request) {
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

	changeset, err := gitHubHandler.resourcesConnection.Verifications.Read.GetChangesetFromShas(headSha, baseSha)
	if _, ok := err.(resources.NoSuchChangesetError); err != nil && !ok {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if changeset != nil {
		_, err = gitHubHandler.resourcesConnection.Verifications.Create.CreateFromChangeset(repository.Id, changeset.Id, "", headEmail)
	} else {
		_, err = gitHubHandler.resourcesConnection.Verifications.Create.Create(repository.Id, headSha, baseSha, headMessage, headUsername, headEmail, "", headEmail)
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

func (gitHubHandler *GitHubHandler) checkSignature(payload []byte, hookSecret, signature string) bool {
	hasher := hmac.New(sha1.New, []byte(hookSecret))
	checksum := hasher.Sum(payload)
	hexString := hex.EncodeToString(checksum)
	return hexString == signature
}