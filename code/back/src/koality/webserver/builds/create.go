package builds

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (buildsHandler *BuildsHandler) create(writer http.ResponseWriter, request *http.Request) {
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

	patchFile, _, _ := request.FormFile("patch")

	var patchContents []byte
	if patchFile != nil {
		patchContents, err = ioutil.ReadAll(patchFile)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
			return
		}
	}

	baseSha := headSha

	if err = buildsHandler.repositoryManager.StorePending(repository, headSha); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	// TODO (bbland): use actual shouldMerge
	shouldMerge := false
	reuseChangeset := true

	build, err := buildsHandler.resourcesConnection.Builds.Create.Create(repositoryId, headSha, headSha, headMessage, headUsername, headEmail, patchContents, headEmail, ref, reuseChangeset, shouldMerge)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedBuild := getSanitizedBuild(build)
	jsonedBuild, err := json.Marshal(sanitizedBuild)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedBuild)
}

func (buildsHandler *BuildsHandler) retrigger(writer http.ResponseWriter, request *http.Request) {
	buildIdString := mux.Vars(request)["buildId"]
	buildId, err := strconv.ParseUint(buildIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse buildId: %v", err)
		return
	}

	build, err := buildsHandler.resourcesConnection.Builds.Read.Get(buildId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	newBuild, err := buildsHandler.resourcesConnection.Builds.Create.CreateFromChangeset(build.RepositoryId,
		build.Changeset.Id, build.EmailToNotify, build.Ref, build.ShouldMerge)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedBuild := getSanitizedBuild(newBuild)
	jsonedBuild, err := json.Marshal(sanitizedBuild)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedBuild)
}
