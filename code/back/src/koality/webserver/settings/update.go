package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"koality/resources"
	"koality/upgrade"
	"koality/webserver/util"
	"net/http"
	"time"
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
	upgradeVersionBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	upgradeVersion := string(upgradeVersionBytes)
	upgradeReader, err := settingsHandler.licenseManager.DownloadUpgrade(upgradeVersion)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	installerPath, err := upgrade.PrepareUpgrade(upgradeReader)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")

	time.AfterFunc(2*time.Second, func() {
		if err := upgrade.RunUpgrade(installerPath); err != nil {
			fmt.Printf("Failed to run upgrade, %v\n", err)
		}
	})
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

func (settingsHandler *SettingsHandler) setAuthenticationSettings(writer http.ResponseWriter, request *http.Request) {
	setAuthenticationRequestData := new(setAuthenticationRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setAuthenticationRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	authenticationSettings, err := settingsHandler.resourcesConnection.Settings.Update.SetAuthenticationSettings(setAuthenticationRequestData.ManualAccountsAllowed, setAuthenticationRequestData.GoogleAccountsAllowed, setAuthenticationRequestData.AllowedDomains)
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

func (settingsHandler *SettingsHandler) setSmtpServerSettings(writer http.ResponseWriter, request *http.Request) {
	smtpSettings := new(setSmtpServerSettingsRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(smtpSettings); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if smtpSettings.Password == "" {
		oldSmtpServerSettings, err := settingsHandler.resourcesConnection.Settings.Read.GetSmtpServerSettings()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(writer, err)
			return
		}

		if oldSmtpServerSettings.Auth.Plain != nil {
			if oldSmtpServerSettings.Auth.Plain.Username == smtpSettings.Username {
				smtpSettings.Password = oldSmtpServerSettings.Auth.Plain.Password
			} else {
				writer.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(writer, "Must provide an SMTP password if you change the username")
				return
			}
		} else if oldSmtpServerSettings.Auth.Login != nil {
			if oldSmtpServerSettings.Auth.Login.Username == smtpSettings.Username {
				smtpSettings.Password = oldSmtpServerSettings.Auth.Login.Password
			} else {
				writer.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(writer, "Must provide an SMTP password if you change the username")
				return
			}
		} else if oldSmtpServerSettings.Auth.CramMd5 != nil {
			if oldSmtpServerSettings.Auth.CramMd5.Username == smtpSettings.Username {
				smtpSettings.Password = oldSmtpServerSettings.Auth.CramMd5.Secret
			} else {
				writer.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(writer, "Must provide an SMTP secret if you change the username")
				return
			}
		}
	}

	var smtpServerSettings *resources.SmtpServerSettings
	var err error
	switch smtpSettings.AuthenticationType {
	case "plain":
		smtpServerSettings, err = settingsHandler.resourcesConnection.Settings.Update.SetSmtpAuthPlain(smtpSettings.Hostname, smtpSettings.Port, "", smtpSettings.Username, smtpSettings.Password, smtpSettings.Hostname)
	case "login":
		smtpServerSettings, err = settingsHandler.resourcesConnection.Settings.Update.SetSmtpAuthLogin(smtpSettings.Hostname, smtpSettings.Port, smtpSettings.Username, smtpSettings.Password)
	case "cramMd5":
		smtpServerSettings, err = settingsHandler.resourcesConnection.Settings.Update.SetSmtpAuthCramMd5(smtpSettings.Hostname, smtpSettings.Port, smtpSettings.Username, smtpSettings.Password)
	default:
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Invalid SMTP authentication type provided")
		return
	}

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	sanitizedSmtpServerSettings := getSanitizedSmtpServerSettings(smtpServerSettings)
	jsonedSmtpServerSettings, err := json.Marshal(sanitizedSmtpServerSettings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedSmtpServerSettings)
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
	email := request.PostFormValue("email")
	firstName := request.PostFormValue("firstName")
	lastName := request.PostFormValue("lastName")
	password := request.PostFormValue("password")

	if err := settingsHandler.licenseManager.SetLicenseKey(licenseKey); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	passwordHash, passwordSalt, err := settingsHandler.passwordHasher.GenerateHashAndSalt(password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	user, err := settingsHandler.resourcesConnection.Users.Create.Create(email, firstName, lastName, passwordHash, passwordSalt, true)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	session, _ := settingsHandler.sessionStore.Get(request, settingsHandler.sessionName)

	util.Login(user.Id, false, session, writer, request)
	fmt.Fprintln(writer, "{success:true}")
}
