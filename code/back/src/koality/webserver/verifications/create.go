package verifications

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"koality/resources"
	"net/http"
	"strconv"
)

func (verificationsHandler *VerificationsHandler) Create(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := request.PostFormValue("repositoryId")
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse repositoryId: %v", err)
		return
	}

	ref := request.PostFormValue("ref")

	repository, err := verificationsHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	headSha, headMessage, headUsername, headEmail, err := verificationsHandler.repositoryManager.GetCommitAttributes(repository, ref)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if err = verificationsHandler.repositoryManager.StorePending(repository, headSha); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return

	}

	baseSha := headSha

	changeset, err := verificationsHandler.resourcesConnection.Verifications.Read.GetChangesetFromShas(headSha, baseSha)
	if _, ok := err.(resources.NoSuchChangesetError); err != nil && !ok {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	var verification *resources.Verification
	if changeset != nil {
		verification, err = verificationsHandler.resourcesConnection.Verifications.Create.CreateFromChangeset(repositoryId, changeset.Id, "", headEmail)
	} else {
		verification, err = verificationsHandler.resourcesConnection.Verifications.Create.Create(repositoryId, headSha, headSha, headMessage, headUsername, headEmail, "", headEmail)
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerification := getSanitizedVerification(verification)
	jsonedVerification, err := json.Marshal(sanitizedVerification)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedVerification)
}

func (verificationsHandler *VerificationsHandler) Retrigger(writer http.ResponseWriter, request *http.Request) {
	verificationIdString := mux.Vars(request)["verificationId"]
	verificationId, err := strconv.ParseUint(verificationIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse verificationId: %v", err)
		return
	}

	verification, err := verificationsHandler.resourcesConnection.Verifications.Read.Get(verificationId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	newVerification, err := verificationsHandler.resourcesConnection.Verifications.Create.CreateFromChangeset(verification.RepositoryId,
		verification.Changeset.Id, verification.MergeTarget, verification.EmailToNotify)
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
