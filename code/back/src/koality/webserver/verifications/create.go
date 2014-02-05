package builds

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"koality/resources"
	"net/http"
	"strconv"
)

func (buildsHandler *VerificationsHandler) create(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := request.PostFormValue("repositoryId")
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	ref := request.PostFormValue("ref")

	repository, err := buildsHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	headSha, headMessage, headUsername, headEmail, err := buildsHandler.repositoryManager.GetCommitAttributes(repository, ref)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	baseSha := headSha

	if err = buildsHandler.repositoryManager.StorePending(repository, headSha); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	changeset, err := buildsHandler.resourcesConnection.Verifications.Read.GetChangesetFromShas(headSha, baseSha)
	if _, ok := err.(resources.NoSuchChangesetError); err != nil && !ok {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	var build *resources.Verification
	if changeset != nil {
		build, err = buildsHandler.resourcesConnection.Verifications.Create.CreateFromChangeset(repositoryId, changeset.Id, "", headEmail)
	} else {
		build, err = buildsHandler.resourcesConnection.Verifications.Create.Create(repositoryId, headSha, headSha, headMessage, headUsername, headEmail, "", headEmail)
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerification := getSanitizedVerification(build)
	jsonedVerification, err := json.Marshal(sanitizedVerification)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedVerification)
}

func (buildsHandler *VerificationsHandler) retrigger(writer http.ResponseWriter, request *http.Request) {
	buildIdString := mux.Vars(request)["buildId"]
	buildId, err := strconv.ParseUint(buildIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse buildId: %v", err)
		return
	}

	build, err := buildsHandler.resourcesConnection.Verifications.Read.Get(buildId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	newVerification, err := buildsHandler.resourcesConnection.Verifications.Create.CreateFromChangeset(build.RepositoryId,
		build.Changeset.Id, build.MergeTarget, build.EmailToNotify)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerification := getSanitizedVerification(newVerification)
	jsonedVerification, err := json.Marshal(sanitizedVerification)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedVerification)
}
