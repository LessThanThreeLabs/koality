package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func (settingsHandler *SettingsHandler) upgrade(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(writer, "need to run an upgrade here")
}

func (settingsHandler *SettingsHandler) setDomainName(writer http.ResponseWriter, request *http.Request) {
	domainNameBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	domainName := string(domainNameBytes)
	_, err = settingsHandler.resourcesConnection.Settings.Update.SetDomainName(domainName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, domainName)
}

func (settingsHandler *SettingsHandler) setS3ExporterSettings(writer http.ResponseWriter, request *http.Request) {
	setS3ExporterRequestData := new(setS3ExporterRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setS3ExporterRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	s3ExporterSettings, err := settingsHandler.resourcesConnection.Settings.Update.SetS3ExporterSettings(setS3ExporterRequestData.AccessKey, setS3ExporterRequestData.SecretKey, setS3ExporterRequestData.BucketName)
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

func (settingsHandler *SettingsHandler) setHipChatSettings(writer http.ResponseWriter, request *http.Request) {
	setHipChatRequestData := new(setHipChatRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setHipChatRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	hipChatSettings, err := settingsHandler.resourcesConnection.Settings.Update.SetHipChatSettings(setHipChatRequestData.AuthenticationToken, setHipChatRequestData.Rooms, setHipChatRequestData.NotifyOn)
	if err != nil {
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

func (settingsHandler *SettingsHandler) setGitHubEnterpriseSettings(writer http.ResponseWriter, request *http.Request) {
	setGitHubEnterpriseRequestData := new(setGitHubEnterpriseRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setGitHubEnterpriseRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	hipChatSettings, err := settingsHandler.resourcesConnection.Settings.Update.SetGitHubEnterpriseSettings(
		setGitHubEnterpriseRequestData.BaseUri, setGitHubEnterpriseRequestData.OAuthClientId, setGitHubEnterpriseRequestData.OAuthClientSecret)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedGitHubEnterpriseSettings := getSanitizedGitHubEnterpriseSettings(hipChatSettings)
	jsonedGitHubEnterpriseSettings, err := json.Marshal(sanitizedGitHubEnterpriseSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedGitHubEnterpriseSettings)
}

func (settingsHandler *SettingsHandler) setWizard(writer http.ResponseWriter, request *http.Request) {
	licenseKey := request.PostFormValue("licenseKey")
	_ = licenseKey // TODO(dhuang) do something with this
	email := request.PostFormValue("email")
	firstName := request.PostFormValue("firstName")
	lastName := request.PostFormValue("lastName")
	password := request.PostFormValue("password")
	passwordHash, passwordSalt, err := settingsHandler.passwordHasher.GenerateHashAndSalt(password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	_, err = settingsHandler.resourcesConnection.Users.Create.Create(email, firstName, lastName, passwordHash, passwordSalt, true)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprintln(writer, "{success:true}")
}
