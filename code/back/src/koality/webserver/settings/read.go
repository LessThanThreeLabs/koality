package settings

import (
	"encoding/json"
	"fmt"
	"koality/resources"
	"net/http"
)

func (settingsHandler *SettingsHandler) getDomainName(writer http.ResponseWriter, request *http.Request) {
	domainName, err := settingsHandler.resourcesConnection.Settings.Read.GetDomainName()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, domainName)
}

func (settingsHandler *SettingsHandler) getAuthenticationSettings(writer http.ResponseWriter, request *http.Request) {
	authenticationSettings, err := settingsHandler.resourcesConnection.Settings.Read.GetAuthenticationSettings()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedAuthenticationSettings := getSanitizedAuthenticationSettings(authenticationSettings)
	jsonedAuthenticationSettings, err := json.Marshal(sanitizedAuthenticationSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedAuthenticationSettings)
}

func (settingsHandler *SettingsHandler) getApiKey(writer http.ResponseWriter, request *http.Request) {
	apiKey, err := settingsHandler.resourcesConnection.Settings.Read.GetApiKey()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, apiKey)
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
	if _, ok := err.(resources.NoSuchSettingError); ok {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "{}")
		return
	} else if err != nil {
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

func (settingsHandler *SettingsHandler) getHipChatSettings(writer http.ResponseWriter, request *http.Request) {
	hipChatSettings, err := settingsHandler.resourcesConnection.Settings.Read.GetHipChatSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "{}")
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedHipChatSettings := getSanitizedHipChatSettings(hipChatSettings)
	jsonedHipChatSettings, err := json.Marshal(sanitizedHipChatSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedHipChatSettings)
}

func (settingsHandler *SettingsHandler) getGitHubEnterpriseSettings(writer http.ResponseWriter, request *http.Request) {
	gitHubEnterpriseSettings, err := settingsHandler.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "{}")
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedGitHubEnterpriseSettings := getSanitizedGitHubEnterpriseSettings(gitHubEnterpriseSettings)
	jsonedGitHubEnterpriseSettings, err := json.Marshal(sanitizedGitHubEnterpriseSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedGitHubEnterpriseSettings)
}
