package stages

import (
	"encoding/json"
	"errors"
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
		fmt.Fprintf(writer, "Unable to parse stageId: %v", err)
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
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
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
		fmt.Fprintf(writer, "Unable to parse verificationId: %v", err)
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
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedStages)
}

func (stagesHandler *StagesHandler) GetRun(writer http.ResponseWriter, request *http.Request) {
	stageRunIdString := mux.Vars(request)["stageRunId"]
	stageRunId, err := strconv.ParseUint(stageRunIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse stageRunId: %v", err)
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
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
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
		fmt.Fprintf(writer, "Unable to parse stageId: %v", err)
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
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedStageRuns)
}

func (stagesHandler *StagesHandler) GetConsoleLines(writer http.ResponseWriter, request *http.Request) {
	stageRunIdString := mux.Vars(request)["stageRunId"]
	stageRunId, err := strconv.ParseUint(stageRunIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse stageRunId: %v", err)
		return
	}

	queryValues := request.URL.Query()
	from := queryValues.Get("from")

	offsetString := queryValues.Get("offset")
	offset, err := strconv.ParseUint(offsetString, 10, 32)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse offset: %v", err)
		return
	}

	resultsString := queryValues.Get("results")
	results, err := strconv.ParseUint(resultsString, 10, 32)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse results: %v", err)
		return
	}

	var consoleLines map[uint64]string
	if from == "" {
		consoleLines, err = stagesHandler.resourcesConnection.Stages.Read.GetAllConsoleLines(stageRunId)
	} else {
		consoleLines, err = stagesHandler.getConsoleLinesForDirection(stageRunId, from, uint32(offset), uint32(results))
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	cleanedConsoleLines := cleanConsoleLines(consoleLines)
	jsonedConsoleLines, err := json.Marshal(cleanedConsoleLines)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedConsoleLines)
}

// The json marshaller (and json standard) only supports strings for keys
func cleanConsoleLines(consoleLines map[uint64]string) map[string]string {
	cleanedConsoleLines := make(map[string]string, len(consoleLines))
	for number, line := range consoleLines {
		numberAsString := strconv.FormatUint(number, 10)
		cleanedConsoleLines[numberAsString] = line
	}
	return cleanedConsoleLines
}

func (stagesHandler *StagesHandler) getConsoleLinesForDirection(stageRunId uint64, from string, offset, results uint32) (map[uint64]string, error) {
	if from == "head" {
		return stagesHandler.resourcesConnection.Stages.Read.GetConsoleLinesHead(stageRunId, offset, results)
	} else if from == "tail" {
		return stagesHandler.resourcesConnection.Stages.Read.GetConsoleLinesTail(stageRunId, offset, results)
	} else {
		return map[uint64]string{}, errors.New("From must be \"head\" or \"tail\"")
	}
}

func (stagesHandler *StagesHandler) GetXunitResults(writer http.ResponseWriter, request *http.Request) {
	stageRunIdString := mux.Vars(request)["stageRunId"]
	stageRunId, err := strconv.ParseUint(stageRunIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse stageRunId: %v", err)
		return
	}

	xunitResults, err := stagesHandler.resourcesConnection.Stages.Read.GetAllXunitResults(stageRunId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedXunitResults := make([]sanitizedXunitResult, 0, len(xunitResults))
	for _, xunitResult := range xunitResults {
		sanitizedXunitResults = append(sanitizedXunitResults, *getSanitizedXunitResult(&xunitResult))
	}

	jsonedXunitResults, err := json.Marshal(sanitizedXunitResults)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedXunitResults)
}

func (stagesHandler *StagesHandler) GetExports(writer http.ResponseWriter, request *http.Request) {
	stageRunIdString := mux.Vars(request)["stageRunId"]
	stageRunId, err := strconv.ParseUint(stageRunIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse stageRunId: %v", err)
		return
	}

	exports, err := stagesHandler.resourcesConnection.Stages.Read.GetAllExports(stageRunId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedExports := make([]sanitizedExport, 0, len(exports))
	for _, export := range exports {
		sanitizedExports = append(sanitizedExports, *getSanitizedExport(&export))
	}

	jsonedExports, err := json.Marshal(sanitizedExports)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedExports)
}
