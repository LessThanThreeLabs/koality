package settings

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"koality/webserver/middleware"
)

type sanitizedAuthenticationSettings struct {
	ManualLoginAllowed bool     `json:"manualLoginAllowed"`
	GoogleLoginAllowed bool     `json:"googleLoginAllowed"`
	AllowedDomains     []string `json:"allowedDomains"`
}

type sanitizedRepositoryKeyPair struct {
	PublicKey string `json:"publicKey"`
}

type sanitizedS3ExporterSettings struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type setS3ExporterRequestData struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type SettingsHandler struct {
	resourcesConnection *resources.Connection
}

func New(resourcesConnection *resources.Connection) (*SettingsHandler, error) {
	return &SettingsHandler{resourcesConnection}, nil
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

	subrouter.HandleFunc("/apiKey/reset",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.resetApiKey)).
		Methods("POST")
	subrouter.HandleFunc("/repositoryKeyPair/reset",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.resetRepositoryKeyPair)).
		Methods("POST")

	subrouter.HandleFunc("/domainName",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setDomainName)).
		Methods("PUT")
	subrouter.HandleFunc("/s3Exporter",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setS3ExporterSettings)).
		Methods("PUT")

	subrouter.HandleFunc("/s3Exporter",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.clearS3ExporterSettings)).
		Methods("DELETE")
}

func (settingsHandler *SettingsHandler) WireApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/domainName", settingsHandler.getDomainName).Methods("GET")
	subrouter.HandleFunc("/authentication", settingsHandler.getAuthenticationSettings).Methods("GET")
	subrouter.HandleFunc("/apiKey", settingsHandler.getApiKey).Methods("GET")
	subrouter.HandleFunc("/repositoryKeyPair", settingsHandler.getRepositoryKeyPair).Methods("GET")
	subrouter.HandleFunc("/s3Exporter", settingsHandler.getS3ExporterSettings).Methods("GET")

	subrouter.HandleFunc("/apiKey/reset", settingsHandler.resetApiKey).Methods("POST")
	subrouter.HandleFunc("/repositoryKeyPair/reset", settingsHandler.resetRepositoryKeyPair).Methods("POST")

	subrouter.HandleFunc("/domainName", settingsHandler.setDomainName).Methods("PUT")
	subrouter.HandleFunc("/s3Exporter", settingsHandler.setS3ExporterSettings).Methods("PUT")

	subrouter.HandleFunc("/s3Exporter", settingsHandler.clearS3ExporterSettings).Methods("DELETE")
}

func getSanitizedAuthenticationSettings(authenticationSettings *resources.AuthenticationSettings) *sanitizedAuthenticationSettings {
	return &sanitizedAuthenticationSettings{
		ManualLoginAllowed: authenticationSettings.ManualLoginAllowed,
		GoogleLoginAllowed: authenticationSettings.GoogleLoginAllowed,
		AllowedDomains:     authenticationSettings.AllowedDomains,
	}
}

func getSanitizedRepositoryKeyPair(repositoryKeyPair *resources.RepositoryKeyPair) *sanitizedRepositoryKeyPair {
	return &sanitizedRepositoryKeyPair{
		PublicKey: repositoryKeyPair.PublicKey,
	}
}

func getSanitizedS3ExporterSettings(sshKey *resources.S3ExporterSettings) *sanitizedS3ExporterSettings {
	return &sanitizedS3ExporterSettings{
		AccessKey:  sshKey.AccessKey,
		SecretKey:  sshKey.SecretKey,
		BucketName: sshKey.BucketName,
	}
}
