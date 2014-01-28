package verifications

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (verificationsHandler *VerificationsHandler) Create(writer http.ResponseWriter, request *http.Request) {
	repositoryIdString := request.PostFormValue("repositoryId")
	repositoryId, err := strconv.ParseUint(repositoryIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	ref := request.PostFormValue("ref")

	repository, err := verificationsHandler.resourcesConnection.Repositories.Read.Get(repositoryId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	verificationsHandler.repositoryManager.StorePending(repository, ref)
	headMessage, headUsername, headEmail, err := verificationsHandler.repositoryManager.GetCommitAttributes(repository, ref)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	headSha := "this-is-a-bad-sha"
	verification, err := verificationsHandler.resourcesConnection.Verifications.Create.Create(repositoryId, headSha, headSha, headMessage, headUsername, headEmail, "", headEmail)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedVerification := getSanitizedVerification(verification)
	jsonedVerification, err := json.Marshal(sanitizedVerification)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedVerification)
}

func (verificationsHandler *VerificationsHandler) Retrigger(writer http.ResponseWriter, request *http.Request) {
	verificationIdString := mux.Vars(request)["verificationId"]
	verificationId, err := strconv.ParseUint(verificationIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
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
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedVerification)
}
