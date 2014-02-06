package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (settingsHandler *SettingsHandler) resetApiKey(writer http.ResponseWriter, request *http.Request) {
	apiKey, err := settingsHandler.resourcesConnection.Settings.Update.ResetApiKey()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	jsonedApiKey, err := json.Marshal(apiKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedApiKey)
}

func (settingsHandler *SettingsHandler) resetRepositoryKeyPair(writer http.ResponseWriter, request *http.Request) {
	repositoryKeyPair, err := settingsHandler.resourcesConnection.Settings.Update.ResetRepositoryKeyPair()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedRepositoryKeyPair := getSanitizedRepositoryKeyPair(repositoryKeyPair)
	jsonedRepositoryKeyPair, err := json.Marshal(sanitizedRepositoryKeyPair)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedRepositoryKeyPair)
}

func (settingsHandler *SettingsHandler) setS3ExporterSettings(writer http.ResponseWriter, request *http.Request) {
	accessKey := request.PostFormValue("accessKey")
	secretKey := request.PostFormValue("secretKey")
	bucketName := request.PostFormValue("bucketName")
	s3ExporterSettings, err := settingsHandler.resourcesConnection.Settings.Update.SetS3ExporterSettings(accessKey, secretKey, bucketName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedS3ExporterSettings := getSanitizedS3ExporterSettings(s3ExporterSettings)
	jsonedS3ExporterSettings, err := json.Marshal(sanitizedS3ExporterSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedS3ExporterSettings)
}
