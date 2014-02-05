package stages

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"koality/webserver/middleware"
	"time"
)

type sanitizedStage struct {
	Id            uint64              `json:"id"`
	BuildId       uint64              `json:"buildId"`
	SectionNumber uint64              `json:"sectionNumber"`
	Name          string              `json:"name"`
	OrderNumber   uint64              `json:"orderNumber"`
	Runs          []sanitizedStageRun `json:"runs"`
}

type sanitizedStageRun struct {
	Id         uint64     `json:"id"`
	StageId    uint64     `json:"stageId"`
	ReturnCode int        `json:"returnCode"`
	Created    *time.Time `json:"created"`
	Started    *time.Time `json:"started"`
	Ended      *time.Time `json:"ended"`
}

type sanitizedXunitResult struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Sysout      string  `json:"sysout"`
	Syserr      string  `json:"syserr"`
	FailureText string  `json:"failureText"`
	ErrorText   string  `json:"errorText"`
	Seconds     float64 `json:"seconds"`
}

type sanitizedExport struct {
	BucketName string `json:"bucketName"`
	Path       string `json:"path"`
	Key        string `json:"key"`
}

type StagesHandler struct {
	resourcesConnection *resources.Connection
}

func New(resourcesConnection *resources.Connection) (*StagesHandler, error) {
	return &StagesHandler{resourcesConnection}, nil
}

func (stagesHandler *StagesHandler) WireStagesAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{stageId:[0-9]+}",
		middleware.IsLoggedInWrapper(stagesHandler.get)).
		Methods("GET")
	subrouter.HandleFunc("/",
		middleware.IsLoggedInWrapper(stagesHandler.getAll)).
		Methods("GET")
}

func (stagesHandler *StagesHandler) WireStagesApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{stageId:[0-9]+}", stagesHandler.get).Methods("GET")
	subrouter.HandleFunc("/", stagesHandler.getAll).Methods("GET")
}

func (stagesHandler *StagesHandler) WireStageRunsAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{stageRunId:[0-9]+}",
		middleware.IsLoggedInWrapper(stagesHandler.getRun)).
		Methods("GET")
	subrouter.HandleFunc("/",
		middleware.IsLoggedInWrapper(stagesHandler.getAllRuns)).
		Methods("GET")
	subrouter.HandleFunc("/{stageRunId:[0-9]+}/lines",
		middleware.IsLoggedInWrapper(stagesHandler.getConsoleLines)).
		Methods("GET")
	subrouter.HandleFunc("/{stageRunId:[0-9]+}/xunits",
		middleware.IsLoggedInWrapper(stagesHandler.getXunitResults)).
		Methods("GET")
	subrouter.HandleFunc("/{stageRunId:[0-9]+}/exports",
		middleware.IsLoggedInWrapper(stagesHandler.getExports)).
		Methods("GET")
}

func (stagesHandler *StagesHandler) WireStageRunsApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/{stageRunId:[0-9]+}", stagesHandler.getRun).Methods("GET")
	subrouter.HandleFunc("/", stagesHandler.getAllRuns).Methods("GET")
	subrouter.HandleFunc("/{stageRunId:[0-9]+}/lines", stagesHandler.getConsoleLines).Methods("GET")
	subrouter.HandleFunc("/{stageRunId:[0-9]+}/xunits", stagesHandler.getXunitResults).Methods("GET")
	subrouter.HandleFunc("/{stageRunId:[0-9]+}/exports", stagesHandler.getExports).Methods("GET")
}

func getSanitizedStage(stage *resources.Stage) *sanitizedStage {
	return &sanitizedStage{
		Id:            stage.Id,
		BuildId:       stage.BuildId,
		SectionNumber: stage.SectionNumber,
		Name:          stage.Name,
		OrderNumber:   stage.OrderNumber,
		Runs:          getSanitizedStageRuns(stage.Runs),
	}
}

func getSanitizedStageRuns(stageRuns []resources.StageRun) []sanitizedStageRun {
	sanitizedStageRuns := make([]sanitizedStageRun, 0, len(stageRuns))
	for _, stageRun := range stageRuns {
		sanitizedStageRuns = append(sanitizedStageRuns, *getSanitizedStageRun(&stageRun))
	}
	return sanitizedStageRuns
}

func getSanitizedStageRun(stageRun *resources.StageRun) *sanitizedStageRun {
	return &sanitizedStageRun{
		Id:         stageRun.Id,
		StageId:    stageRun.StageId,
		ReturnCode: stageRun.ReturnCode,
		Created:    stageRun.Created,
		Started:    stageRun.Started,
		Ended:      stageRun.Ended,
	}
}

func getSanitizedXunitResult(xunitResult *resources.XunitResult) *sanitizedXunitResult {
	return &sanitizedXunitResult{
		Name:        xunitResult.Name,
		Path:        xunitResult.Path,
		Sysout:      xunitResult.Sysout,
		Syserr:      xunitResult.Syserr,
		FailureText: xunitResult.FailureText,
		ErrorText:   xunitResult.ErrorText,
		Seconds:     xunitResult.Seconds,
	}
}

func getSanitizedExport(export *resources.Export) *sanitizedExport {
	return &sanitizedExport{
		BucketName: export.BucketName,
		Path:       export.Path,
		Key:        export.Key,
	}
}
