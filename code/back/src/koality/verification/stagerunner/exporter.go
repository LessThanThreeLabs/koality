package stagerunner

import (
	"encoding/json"
	"errors"
	"fmt"
	"koality/resources"
	"koality/shell"
	"koality/util/export"
)

type Exporter interface {
	ExportAndGetResults(stageId, stageRunId uint64, stageRunner *StageRunner, exportPaths []string, environment map[string]string) ([]resources.Export, error)
}

type S3Exporter struct{}

func (exporter S3Exporter) ExportAndGetResults(stageId, stageRunId uint64, stageRunner *StageRunner, exportPaths []string, environment map[string]string) ([]resources.Export, error) {
	if len(exportPaths) == 0 {
		return nil, nil
	}

	exportPrefix := fmt.Sprintf("repository/%d/verification/%d/stage/%d/stageRun/%d",
		stageRunner.verification.Changeset.RepositoryId, stageRunner.verification.Id, stageId, stageRunId)
	s3ExporterSettings, err := stageRunner.resourcesConnection.Settings.Read.GetS3ExporterSettings()
	if err != nil {
		return nil, err
	} else if s3ExporterSettings == nil {
		return nil, errors.New("No s3 settings present.")
	}

	args := append([]string{
		shell.Quote(s3ExporterSettings.AccessKey),
		shell.Quote(s3ExporterSettings.SecretKey),
		shell.Quote(s3ExporterSettings.BucketName),
		exportPrefix,
		"us-east-1", // US Standard
	}, exportPaths...)
	writeBuffer, err := stageRunner.copyAndRunExecOnVm(stageRunId, "exportPaths", args, environment)
	if err != nil {
		return nil, err
	}

	var exportOutput export.ExportOutput
	if err = json.Unmarshal(writeBuffer.Bytes(), &exportOutput); err != nil {
		return nil, err
	}

	return exportOutput.Exports, exportOutput.Error
}
