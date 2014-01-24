package stages

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (stagesHandler *StagesHandler) Get(writer http.ResponseWriter, request *http.Request) {
	stageIdString := mux.Vars(request)["stageId"]
	stageId, err := strconv.ParseUint(stageIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	stage, err := stagesHandler.resourcesConnection.Stages.Read.Get(stageId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedStage := getSanitizedStage(stage)
	jsonedStage, err := json.Marshal(sanitizedStage)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedStage)
}

func (stagesHandler *StagesHandler) GetAll(writer http.ResponseWriter, request *http.Request) {
	queryValues := request.URL.Query()
	verificationIdString := queryValues.Get("verificationId")
	verificationId, err := strconv.ParseUint(verificationIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	stages, err := stagesHandler.resourcesConnection.Stages.Read.GetAll(verificationId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedStages := make([]sanitizedStage, 0, len(stages))
	for _, stage := range stages {
		sanitizedStages = append(sanitizedStages, *getSanitizedStage(&stage))
	}

	jsonedStages, err := json.Marshal(sanitizedStages)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedStages)
}

func (stagesHandler *StagesHandler) GetRun(writer http.ResponseWriter, request *http.Request) {
	stageRunIdString := mux.Vars(request)["stageRunId"]
	stageRunId, err := strconv.ParseUint(stageRunIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	stageRun, err := stagesHandler.resourcesConnection.Stages.Read.GetRun(stageRunId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedStageRun := getSanitizedStageRun(stageRun)
	jsonedStageRun, err := json.Marshal(sanitizedStageRun)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedStageRun)
}

func (stagesHandler *StagesHandler) GetAllRuns(writer http.ResponseWriter, request *http.Request) {
	queryValues := request.URL.Query()
	stageIdString := queryValues.Get("stageId")
	stageId, err := strconv.ParseUint(stageIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	stageRuns, err := stagesHandler.resourcesConnection.Stages.Read.GetAllRuns(stageId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedStageRuns := make([]sanitizedStageRun, 0, len(stageRuns))
	for _, stageRun := range stageRuns {
		sanitizedStageRuns = append(sanitizedStageRuns, *getSanitizedStageRun(&stageRun))
	}

	jsonedStageRuns, err := json.Marshal(sanitizedStageRuns)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedStageRuns)
}

func (stagesHandler *StagesHandler) GetAllConsoleLines(writer http.ResponseWriter, request *http.Request) {
	stageRunIdString := mux.Vars(request)["stageRunId"]
	stageRunId, err := strconv.ParseUint(stageRunIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	consoleLines, err := stagesHandler.resourcesConnection.Stages.Read.GetAllConsoleLines(stageRunId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedConsoleLines := getSanitizedConsoleLines(consoleLines)
	jsonedConsoleLines, err := json.Marshal(sanitizedConsoleLines)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedConsoleLines)
}
