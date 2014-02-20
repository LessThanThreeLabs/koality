package events

import (
	"fmt"
	"github.com/gorilla/mux"
	licenseclient "koality/license/client"
	"koality/resources"
	"koality/webserver/middleware"
)

type domainNameUpdatedEventData struct {
	DomainName resources.DomainName `json:"domainName"`
}

type authenticationSettingsUpdatedEventData struct {
	ManualAccountsAllowed bool     `json:"manualAccountsAllowed"`
	GoogleAccountsAllowed bool     `json:"googleAccountsAllowed"`
	AllowedDomains        []string `json:"allowedDomains"`
}

type repositoryKeyPairUpdatedEventData struct {
	PublicKey string `json:"publicKey"`
}

type s3ExporterSettingsUpdatedEventData struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type smtpServerSettingsUpdatedEventData struct {
	Hostname           string `json:"hostname"`
	Port               uint16 `json:"port"`
	AuthenticationType string `json:"authenticationType"`
	Username           string `json:"username"`
}

type hipChatSettingsUpdatedEventData struct {
	AuthenticationToken string   `json:"authenticationToken"`
	Rooms               []string `json:"rooms"`
	NotifyOn            string   `json:"notifyOn"`
}

type gitHubEnterpriseSettingsUpdatedEventData struct {
	BaseUri           string `json:"baseUri"`
	OAuthClientId     string `json:"oAuthClientId"`
	OAuthClientSecret string `json:"oAuthClientSecret"`
}

type licenseSettingsUpdatedEventData struct {
	LicenseKey   string `json:"licenseKey"`
	MaxExecutors uint32 `json:"maxExecutors"`
	Status       string `json:"status"`
}

func (eventsHandler *EventsHandler) wireSettingsAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/domainName/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(domainNameUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/authenticationSettings/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(authenticationSettingsUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/repositoryKeyPair/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(repositoryKeyPairUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/s3ExporterSettings/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(s3ExporterSettingsUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/smtpServerSettings/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(smtpServerSettingsUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/hipChatSettings/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(hipChatSettingsUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/gitHubEnterprise/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(gitHubEnterpriseSettingsUpdatedSubscriptions, true, false))).
		Methods("POST")
	subrouter.HandleFunc("/licenseSettings/subscribe",
		middleware.IsAdminWrapper(eventsHandler.resourcesConnection,
			eventsHandler.createSubscription(licenseSettingsUpdatedSubscriptions, true, false))).
		Methods("POST")

	// subrouter.HandleFunc("/created/{subscriptionId:[0-9]+}",
	// 	middleware.IsLoggedInWrapper(
	// 		eventsHandler.deleteSubscription(userCreatedSubscriptions))).
	// 	Methods("DELETE")
	// subrouter.HandleFunc("/deleted/{subscriptionId:[0-9]+}",
	// 	middleware.IsLoggedInWrapper(
	// 		eventsHandler.deleteSubscription(userDeletedSubscriptions))).
	// 	Methods("DELETE")
	// subrouter.HandleFunc("/name/{subscriptionId:[0-9]+}",
	// 	middleware.IsLoggedInWrapper(
	// 		eventsHandler.deleteSubscription(userNameUpdatedSubscriptions))).
	// 	Methods("DELETE")
	// subrouter.HandleFunc("/admin/{subscriptionId:[0-9]+}",
	// 	middleware.IsLoggedInWrapper(
	// 		eventsHandler.deleteSubscription(userAdminUpdatedSubscriptions))).
	// 	Methods("DELETE")
	// subrouter.HandleFunc("/sshKeyAdded/{subscriptionId:[0-9]+}",
	// 	middleware.IsLoggedInWrapper(
	// 		eventsHandler.deleteSubscription(userSshKeyAddedSubscriptions))).
	// 	Methods("DELETE")
	// subrouter.HandleFunc("/sshKeyRemoved/{subscriptionId:[0-9]+}",
	// 	middleware.IsLoggedInWrapper(
	// 		eventsHandler.deleteSubscription(userSshKeyRemovedSubscriptions))).
	// 	Methods("DELETE")
}

func (eventsHandler *EventsHandler) listenForSettingsEvents() error {
	_, err := eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToDomainNameUpdatedEvents(eventsHandler.handleDomainNameUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToAuthenticationSettingsUpdatedEvents(eventsHandler.handleAuthenticationSettingsUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToRepositoryKeyPairUpdatedEvents(eventsHandler.handleRepositoryKeyPairUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToS3ExporterSettingsUpdatedEvents(eventsHandler.handleS3ExporterSettingsUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToS3ExporterSettingsClearedEvents(eventsHandler.handleS3ExporterSettingsClearedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToSmtpServerSettingsUpdatedEvents(eventsHandler.handleSmtpServerSettingsUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToHipChatSettingsUpdatedEvents(eventsHandler.handleHipChatSettingsUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToHipChatSettingsClearedEvents(eventsHandler.handleHipChatSettingsClearedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToGitHubEnterpriseSettingsUpdatedEvents(eventsHandler.handleGitHubEnterpriseSettingsUpdatedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToGitHubEnterpriseSettingsClearedEvents(eventsHandler.handleGitHubEnterpriseSettingsClearedEvent)
	if err != nil {
		return err
	}

	_, err = eventsHandler.resourcesConnection.Settings.Subscription.SubscribeToLicenseSettingsUpdatedEvents(eventsHandler.handleLicenseSettingsUpdatedEvent)
	if err != nil {
		return err
	}

	return nil
}

func (eventsHandler *EventsHandler) handleDomainNameUpdatedEvent(domainName resources.DomainName) {
	data := domainNameUpdatedEventData{domainName}
	eventsHandler.handleEvent(domainNameUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleAuthenticationSettingsUpdatedEvent(authenticationSettings *resources.AuthenticationSettings) {
	data := authenticationSettingsUpdatedEventData{authenticationSettings.ManualAccountsAllowed,
		authenticationSettings.GoogleAccountsAllowed, authenticationSettings.AllowedDomains}
	eventsHandler.handleEvent(authenticationSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleRepositoryKeyPairUpdatedEvent(keyPair *resources.RepositoryKeyPair) {
	data := repositoryKeyPairUpdatedEventData{keyPair.PublicKey}
	eventsHandler.handleEvent(repositoryKeyPairUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleS3ExporterSettingsUpdatedEvent(s3Settings *resources.S3ExporterSettings) {
	data := s3ExporterSettingsUpdatedEventData{s3Settings.AccessKey, s3Settings.SecretKey, s3Settings.BucketName}
	eventsHandler.handleEvent(s3ExporterSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleS3ExporterSettingsClearedEvent() {
	data := map[int]int{}
	eventsHandler.handleEvent(s3ExporterSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleSmtpServerSettingsUpdatedEvent(smtpServerSettings *resources.SmtpServerSettings) {
	var authenticationType, username string
	if smtpServerSettings.Auth.Plain != nil {
		authenticationType = "plain"
		username = smtpServerSettings.Auth.Plain.Username
	} else if smtpServerSettings.Auth.Plain != nil {
		authenticationType = "login"
		username = smtpServerSettings.Auth.Login.Username
	} else if smtpServerSettings.Auth.CramMd5 != nil {
		authenticationType = "cramMd5"
		username = smtpServerSettings.Auth.CramMd5.Username
	}

	data := smtpServerSettingsUpdatedEventData{smtpServerSettings.Hostname, smtpServerSettings.Port, authenticationType, username}
	eventsHandler.handleEvent(smtpServerSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleHipChatSettingsUpdatedEvent(hipChatSettings *resources.HipChatSettings) {
	data := hipChatSettingsUpdatedEventData{hipChatSettings.AuthenticationToken, hipChatSettings.Rooms, hipChatSettings.NotifyOn}
	eventsHandler.handleEvent(hipChatSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleHipChatSettingsClearedEvent() {
	data := map[int]int{}
	eventsHandler.handleEvent(hipChatSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleGitHubEnterpriseSettingsUpdatedEvent(gitHubEnterpriseSettings *resources.GitHubEnterpriseSettings) {
	data := gitHubEnterpriseSettingsUpdatedEventData{gitHubEnterpriseSettings.BaseUri,
		gitHubEnterpriseSettings.OAuthClientId, gitHubEnterpriseSettings.OAuthClientSecret}
	eventsHandler.handleEvent(gitHubEnterpriseSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleGitHubEnterpriseSettingsClearedEvent() {
	data := map[int]int{}
	eventsHandler.handleEvent(gitHubEnterpriseSettingsUpdatedSubscriptions, 0, data)
}

func (eventsHandler *EventsHandler) handleLicenseSettingsUpdatedEvent(licenseSettings *resources.LicenseSettings) {
	status := "Active"
	if !licenseSettings.IsValid {
		status = "Invalid: " + licenseSettings.InvalidReason
	} else if licenseSettings.InvalidReason != "" {
		status = fmt.Sprintf("Warning (%d/%d): %s", licenseSettings.LicenseCheckFailures, licenseclient.MaxLicenseCheckFailures, licenseSettings.InvalidReason)
	}

	data := licenseSettingsUpdatedEventData{licenseSettings.LicenseKey, licenseSettings.MaxExecutors, status}
	eventsHandler.handleEvent(licenseSettingsUpdatedSubscriptions, 0, data)
}
