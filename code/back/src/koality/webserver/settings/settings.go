package settings

import (
	"github.com/gorilla/mux"
	"koality/resources"
	"koality/webserver/middleware"
)

type sanitizedRepositoryKeyPair struct {
	PublicKey string `json:"publicKey"`
}

type sanitizedS3ExporterSettings struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type sanitizedApiKey struct {
	Key string `json:"key"`
}

type SettingsHandler struct {
	resourcesConnection *resources.Connection
}

func New(resourcesConnection *resources.Connection) (*SettingsHandler, error) {
	return &SettingsHandler{resourcesConnection}, nil
}

func (settingsHandler *SettingsHandler) WireAppSubroutes(subrouter *mux.Router) {
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

	subrouter.HandleFunc("/s3Exporter",
		middleware.IsAdminWrapper(settingsHandler.resourcesConnection, settingsHandler.setS3ExporterSettings)).
		Methods("PUT")
}

func (settingsHandler *SettingsHandler) WireApiSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/apiKey", settingsHandler.getApiKey).Methods("GET")
	subrouter.HandleFunc("/repositoryKeyPair", settingsHandler.getRepositoryKeyPair).Methods("GET")
	subrouter.HandleFunc("/s3Exporter", settingsHandler.getS3ExporterSettings).Methods("GET")

	subrouter.HandleFunc("/apiKey/reset", settingsHandler.resetApiKey).Methods("POST")
	subrouter.HandleFunc("/repositoryKeyPair/reset", settingsHandler.resetRepositoryKeyPair).Methods("POST")

	subrouter.HandleFunc("/s3Exporter", settingsHandler.setS3ExporterSettings).Methods("PUT")
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

func getSanitizedApiKey(sshKey *resources.ApiKey) *sanitizedApiKey {
	return &sanitizedApiKey{
		Key: sshKey.Key,
	}
}
