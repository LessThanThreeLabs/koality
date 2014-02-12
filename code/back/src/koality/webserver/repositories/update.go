package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"io/ioutil"
	"koality/github"
	"koality/resources"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func (repositoriesHandler *RepositoriesHandler) setRemoteUri(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	remoteUriBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	remoteUri := string(remoteUriBytes)
	err = repositoriesHandler.resourcesConnection.Repositories.Update.SetRemoteUri(repositoryId, remoteUri)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (repositoriesHandler *RepositoriesHandler) setGitHubHookTypes(writer http.ResponseWriter, request *http.Request) {
	hookTypes := make([]string, 0, 2)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&hookTypes); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprint(writer, err)
		return
	}

	if repository.GitHub == nil {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, repository was not added from GitHub")
		return
	}

	hookSecret := generateSecret()
	hookId := new(int64)
	addRepositoryHookAction := func(repository *resources.Repository) error {
		newHookId, err := repositoriesHandler.gitHubConnection.AddRepositoryHook(repository, hookTypes, hookSecret)
		if err != nil {
			return err
		}
		*hookId = newHookId
		return nil
	}
	userId := context.Get(request, "userId").(uint64)
	redirectUri, err := repositoriesHandler.authenticateGitHubRepositoryAction(userId, repository, addRepositoryHookAction)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	} else if redirectUri != "" {
		writer.WriteHeader(http.StatusPreconditionFailed)
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, fmt.Sprintf(`{"redirectUri": "%s"}`, redirectUri))
		return
	}

	err = repositoriesHandler.resourcesConnection.Repositories.Update.SetGitHubHook(repositoryId, *hookId, hookSecret, hookTypes)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if repository.GitHub.HookId != 0 {
		err = repositoriesHandler.gitHubConnection.RemoveRepositoryHook(repository)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
			return
		}
	}

	fmt.Fprint(writer, "ok")
}

func (repositoriesHandler *RepositoriesHandler) clearGitHubHook(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := mux.Vars(request)["repositoryId"]
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	repository, err := repositoriesHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if repository.GitHub == nil {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, repository was not added from GitHub")
		return
	}

	if repository.GitHub.HookId == 0 {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, repository does not have a GitHub hook")
		return
	}

	userId := context.Get(request, "userId").(uint64)
	redirectUri, err := repositoriesHandler.authenticateGitHubRepositoryAction(userId, repository, repositoriesHandler.gitHubConnection.RemoveRepositoryHook)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	} else if redirectUri != "" {
		writer.WriteHeader(http.StatusPreconditionFailed)
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, fmt.Sprintf(`{"redirectUri": "%s"}`, redirectUri))
		return
	}

	err = repositoriesHandler.resourcesConnection.Repositories.Update.ClearGitHubHook(repository.Id)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

func (repositoriesHandler *RepositoriesHandler) authenticateGitHubRepositoryAction(userId uint64, repository *resources.Repository, action func(repository *resources.Repository) error) (string, error) {
	err := action(repository)
	if _, ok := err.(github.InvalidOAuthTokenError); ok {
		user, err := repositoriesHandler.resourcesConnection.Users.Read.Get(userId)
		if err != nil {
			return "", err
		}
		ok, err := repositoriesHandler.gitHubConnection.CheckValidOAuthToken(user.GitHubOAuth)
		if err != nil {
			return "", err
		}
		if !ok {
			authorizationUrl, err := repositoriesHandler.gitHubConnection.GetAuthorizationUrl("editRepository")
			if err != nil {
				return "", err
			}
			return authorizationUrl, nil
		} else {
			err = repositoriesHandler.resourcesConnection.Repositories.Update.SetGitHubOAuthToken(repository.Id, user.GitHubOAuth)
			if err != nil {
				return "", err
			}
			repository.GitHub.OAuthToken = user.GitHubOAuth
			err = action(repository)
			if err != nil {
				return "", err
			}
		}
	}
	return "", err
}

func generateSecret() string {
	// 36^12 < 2^63; does not overflow
	maxRandValue := int64(1)
	for i := 0; i < 12; i++ {
		maxRandValue *= 36
	}
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.FormatInt(random.Int63n(maxRandValue), 36)
}
