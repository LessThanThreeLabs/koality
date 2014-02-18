package settings

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/license/checker"
	"koality/resources"
	"koality/webserver/middleware"
)

type sanitizedAuthenticationSettings struct {
	ManualAccountsAllowed bool     `json:"manualAccountsAllowed"`
	GoogleAccountsAllowed bool     `json:"googleAccountsAllowed"`
	AllowedDomains        []string `json:"allowedDomains"`
}

type sanitizedRepositoryKeyPair struct {
	PublicKey string `json:"publicKey"`
}

type sanitizedS3ExporterSettings struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type sanitizedSmtpServerSettings struct {
	Hostname           string `json:"hostname"`
	Port               uint16 `json:"port"`
	AuthenticationType string `json:"authenticationType"`
	Username           string `json:"username"`
}

type sanitizedHipChatSettings struct {
	AuthenticationToken string   `json:"authenticationToken"`
	Rooms               []string `json:"rooms"`
	NotifyOn            string   `json:"notifyOn"`
}

type sanitizedGitHubEnterpriseSettings struct {
	BaseUri           string `json:"baseUri"`
	OAuthClientId     string `json:"oAuthClientId"`
	OAuthClientSecret string `json:"oAuthClientSecret"`
}

type sanitizedLicenseSettings struct {
	LicenseKey   string `json:"licenseKey"`
	MaxExecutors uint32 `json:"maxExecutors"`
	Status       string `json:"status"`
}

type setAuthenticationRequestData struct {
	ManualAccountsAllowed bool     `json:"manualAccountsAllowed"`
	GoogleAccountsAllowed bool     `json:"googleAccountsAllowed"`
	AllowedDomains        []string `json:"allowedDomains"`
}

type setS3ExporterRequestData struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type setSmtpServerSettingsRequestData struct {
	Hostname           string `json:"hostname"`
	Port               uint16 `json:"port"`
	AuthenticationType string `json:"authenticationType"`
	Username           string `json:"username"`
	Password           string `json:"password"`
}

type setHipChatRequestData struct {
	AuthenticationToken string   `json:"authenticationToken"`
	Rooms               []string `json:"rooms"`
	NotifyOn            string   `json:"notifyOn"`
}

type setGitHubEnterpriseRequestData struct {
	BaseUri           string `json:"baseUri"`
	OAuthClientId     string `json:"oAuthClientId"`
	OAuthClientSecret string `json:"oAuthClientSecret"`
}

type SettingsHandler struct {
	resourcesConnection *resources.Connection
	sessionStore        sessions.Store
	sessionName         string
	passwordHasher      *resources.PasswordHasher
	licenseManager      *licensechecker.LicenseManager
}

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string, passwordHasher *resources.PasswordHasher, licenseManager *licensechecker.LicenseManager) (*SettingsHandler, error) {
	return &SettingsHandler{resourcesConnection, sessionStore, sessionName, passwordHasher, licenseManager}, nil
}

func (settingsHandler *SettingsHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/domainName",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getDomainName)).
		Methods("GET")
	subrouter.HandleFunc("/authentication",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getAuthenticationSettings)).
		Methods("GET")
	subrouter.HandleFunc("/apiKey",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getApiKey)).
		Methods("GET")
	subrouter.HandleFunc("/repositoryKeyPair",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getRepositoryKeyPair)).
		Methods("GET")
	subrouter.HandleFunc("/s3Exporter",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getS3ExporterSettings)).
		Methods("GET")
	subrouter.HandleFunc("/smtp",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getSmtpServerSettings)).
		Methods("GET")
	subrouter.HandleFunc("/hipChat",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getHipChatSettings)).
		Methods("GET")
	subrouter.HandleFunc("/gitHubEnterprise",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getGitHubEnterpriseSettings)).
		Methods("GET")
	subrouter.HandleFunc("/license",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getLicense)).
		Methods("GET")
	subrouter.HandleFunc("/upgradeStatus",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.getUpgradeStatus)).
		Methods("GET")

	subrouter.HandleFunc("/apiKey/reset",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.resetApiKey)).
		Methods("POST")
	subrouter.HandleFunc("/repositoryKeyPair/reset",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.resetRepositoryKeyPair)).
		Methods("POST")
	subrouter.HandleFunc("/wizard",
		settingsHandler.setWizard). // Does not require admin
		Methods("POST")
	subrouter.HandleFunc("/upgrade",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.upgrade)).
		Methods("POST")

	subrouter.HandleFunc("/domainName",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setDomainName)).
		Methods("PUT")
	subrouter.HandleFunc("/authentication",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setAuthenticationSettings)).
		Methods("PUT")
	subrouter.HandleFunc("/s3Exporter",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setS3ExporterSettings)).
		Methods("PUT")
	subrouter.HandleFunc("/smtp",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setSmtpServerSettings)).
		Methods("PUT")
	subrouter.HandleFunc("/hipChat",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setHipChatSettings)).
		Methods("PUT")
	subrouter.HandleFunc("/gitHubEnterprise",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setGitHubEnterpriseSettings)).
		Methods("PUT")

	subrouter.HandleFunc("/s3Exporter",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.clearS3ExporterSettings)).
		Methods("DELETE")
	subrouter.HandleFunc("/hipChat",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.clearHipChatSettings)).
		Methods("DELETE")
	subrouter.HandleFunc("/gitHubEnterprise",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.clearGitHubEnterpriseSettings)).
		Methods("DELETE")
}

func (settingsHandler *SettingsHandler) WireApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/domainName", settingsHandler.getDomainName).Methods("GET")
	subrouter.HandleFunc("/authentication", settingsHandler.getAuthenticationSettings).Methods("GET")
	subrouter.HandleFunc("/apiKey", settingsHandler.getApiKey).Methods("GET")
	subrouter.HandleFunc("/repositoryKeyPair", settingsHandler.getRepositoryKeyPair).Methods("GET")
	subrouter.HandleFunc("/s3Exporter", settingsHandler.getS3ExporterSettings).Methods("GET")
	subrouter.HandleFunc("/hipChat", settingsHandler.getHipChatSettings).Methods("GET")
	subrouter.HandleFunc("/gitHubEnterprise", settingsHandler.getGitHubEnterpriseSettings).Methods("GET")
	subrouter.HandleFunc("/license", settingsHandler.getLicense).Methods("GET")
	subrouter.HandleFunc("/upgradeStatus", settingsHandler.getUpgradeStatus).Methods("GET")

	subrouter.HandleFunc("/apiKey/reset", settingsHandler.resetApiKey).Methods("POST")
	subrouter.HandleFunc("/repositoryKeyPair/reset", settingsHandler.resetRepositoryKeyPair).Methods("POST")
	subrouter.HandleFunc("/upgrade", settingsHandler.resetRepositoryKeyPair).Methods("POST")

	subrouter.HandleFunc("/domainName", settingsHandler.setDomainName).Methods("PUT")
	subrouter.HandleFunc("/s3Exporter", settingsHandler.setS3ExporterSettings).Methods("PUT")
	subrouter.HandleFunc("/hipChat", settingsHandler.setHipChatSettings).Methods("PUT")
	subrouter.HandleFunc("/gitHubEnterprise", settingsHandler.setHipChatSettings).Methods("PUT")

	subrouter.HandleFunc("/s3Exporter", settingsHandler.clearS3ExporterSettings).Methods("DELETE")
	subrouter.HandleFunc("/hipChat", settingsHandler.clearS3ExporterSettings).Methods("DELETE")
	subrouter.HandleFunc("/gitHubEnterprise", settingsHandler.clearS3ExporterSettings).Methods("DELETE")
}

func getSanitizedAuthenticationSettings(authenticationSettings *resources.AuthenticationSettings) *sanitizedAuthenticationSettings {
	return &sanitizedAuthenticationSettings{
		ManualAccountsAllowed: authenticationSettings.ManualAccountsAllowed,
		GoogleAccountsAllowed: authenticationSettings.GoogleAccountsAllowed,
		AllowedDomains:        authenticationSettings.AllowedDomains,
	}
}

func getSanitizedRepositoryKeyPair(repositoryKeyPair *resources.RepositoryKeyPair) *sanitizedRepositoryKeyPair {
	return &sanitizedRepositoryKeyPair{
		PublicKey: repositoryKeyPair.PublicKey,
	}
}

func getSanitizedS3ExporterSettings(settings *resources.S3ExporterSettings) *sanitizedS3ExporterSettings {
	return &sanitizedS3ExporterSettings{
		AccessKey:  settings.AccessKey,
		SecretKey:  settings.SecretKey,
		BucketName: settings.BucketName,
	}
}

func getSanitizedSmtpServerSettings(settings *resources.SmtpServerSettings) *sanitizedSmtpServerSettings {
	var authenticationType, username string
	if settings.Auth.Plain != nil {
		authenticationType = "plain"
		username = settings.Auth.Plain.Username
	} else if settings.Auth.Plain != nil {
		authenticationType = "login"
		username = settings.Auth.Login.Username
	} else if settings.Auth.CramMd5 != nil {
		authenticationType = "cramMd5"
		username = settings.Auth.CramMd5.Username
	}

	return &sanitizedSmtpServerSettings{
		Hostname:           settings.Hostname,
		Port:               settings.Port,
		AuthenticationType: authenticationType,
		Username:           username,
	}
}

func getSanitizedHipChatSettings(settings *resources.HipChatSettings) *sanitizedHipChatSettings {
	return &sanitizedHipChatSettings{
		AuthenticationToken: settings.AuthenticationToken,
		Rooms:               settings.Rooms,
		NotifyOn:            settings.NotifyOn,
	}
}

func getSanitizedGitHubEnterpriseSettings(settings *resources.GitHubEnterpriseSettings) *sanitizedGitHubEnterpriseSettings {
	return &sanitizedGitHubEnterpriseSettings{
		BaseUri:           settings.BaseUri,
		OAuthClientId:     settings.OAuthClientId,
		OAuthClientSecret: settings.OAuthClientSecret,
	}
}

func getSanitizedLicenseSettings(settings *resources.LicenseSettings) *sanitizedLicenseSettings {
	status := "Active"
	if !settings.IsValid {
		status = "Invalid: " + settings.InvalidReason
	} else if settings.InvalidReason != "" {
		status = fmt.Sprintf("Warning (%d/%d): %s", settings.LicenseCheckFailures, licensechecker.MaxLicenseCheckFailures, settings.InvalidReason)
	}
	return &sanitizedLicenseSettings{
		LicenseKey:   settings.LicenseKey,
		MaxExecutors: settings.MaxExecutors,
		Status:       status,
	}
}
