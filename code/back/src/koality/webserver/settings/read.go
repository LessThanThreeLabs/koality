package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (settingsHandler *SettingsHandler) getApiKey(writer http.ResponseWriter, request *http.Request) {
	apiKey, err := settingsHandler.resourcesConnection.Settings.Read.GetApiKey()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedApiKey := getSanitizedApiKey(apiKey)
	jsonedApiKey, err := json.Marshal(sanitizedApiKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedApiKey)
}

func (settingsHandler *SettingsHandler) getRepositoryKeyPair(writer http.ResponseWriter, request *http.Request) {
	repositoryKeyPair, err := settingsHandler.resourcesConnection.Settings.Read.GetRepositoryKeyPair()
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

func (settingsHandler *SettingsHandler) getS3ExporterSettings(writer http.ResponseWriter, request *http.Request) {
	s3ExporterSettings, err := settingsHandler.resourcesConnection.Settings.Read.GetS3ExporterSettings()
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
