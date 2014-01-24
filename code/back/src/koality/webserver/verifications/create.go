package verifications

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

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
	fmt.Fprintf(writer, "{id:%d}", newVerification.Id)
}
