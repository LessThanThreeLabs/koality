package stages

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"time"
)

type sanitizedStage struct {
	Id             uint64              `json:"id"`
	VerificationId uint64              `json:"verificationId"`
	SectionNumber  uint64              `json:"sectionNumber"`
	Name           string              `json:"name"`
	OrderNumber    uint64              `json:"orderNumber"`
	Runs           []sanitizedStageRun `json:"runs"`
}

type sanitizedStageRun struct {
	Id         uint64     `json:"id"`
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

func (verificationsHandler *StagesHandler) WireSubroutes(subrouter *mux.Router) {
	// subrouter.HandleFunc("/{verificationId:[0-9]+}", verificationsHandler.Get).Methods("GET")
	// subrouter.HandleFunc("/tail", verificationsHandler.GetTail).Methods("GET")

	// subrouter.HandleFunc("/{verificationId:[0-9]+}/retrigger", verificationsHandler.Retrigger).Methods("POST")
}

func getSanitizedStage(stage *resources.Stage) *sanitizedStage {
	return &sanitizedStage{
		Id:             stage.Id,
		VerificationId: stage.VerificationId,
		SectionNumber:  stage.SectionNumber,
		Name:           stage.Name,
		OrderNumber:    stage.OrderNumber,
		Runs:           getSanitizedStageRuns(stage.Runs),
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
