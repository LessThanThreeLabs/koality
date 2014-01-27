package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (settingsHandler *SettingsHandler) ResetApiKey(writer http.ResponseWriter, request *http.Request) {
	apiKey, err := settingsHandler.resourcesConnection.Settings.Update.ResetApiKey()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedApiKey := getSanitizedApiKey(apiKey)
	jsonedApiKey, err := json.Marshal(sanitizedApiKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedApiKey)
}

func (settingsHandler *SettingsHandler) ResetRepositoryKeyPair(writer http.ResponseWriter, request *http.Request) {
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
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedRepositoryKeyPair)
}

func (settingsHandler *SettingsHandler) SetS3ExporterSettings(writer http.ResponseWriter, request *http.Request) {
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
		fmt.Print(writer, err)
		return
	}
	fmt.Fprintf(writer, "%s", jsonedS3ExporterSettings)
}
