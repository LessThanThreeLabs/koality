package settings

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	domainNameUpdatedSubscriptionManager               resources.SubscriptionManager
	authenticationSettingsUpdatedSubscriptionManager   resources.SubscriptionManager
	repositoryKeyPairUpdatedSubscriptionManager        resources.SubscriptionManager
	s3ExporterSettingsUpdatedSubscriptionManager       resources.SubscriptionManager
	s3ExporterSettingsClearedSubscriptionManager       resources.SubscriptionManager
	hipChatSettingsUpdatedSubscriptionManager          resources.SubscriptionManager
	hipChatSettingsClearedSubscriptionManager          resources.SubscriptionManager
	cookieStoreKeysUpdatedSubscriptionManager          resources.SubscriptionManager
	smtpServerSettingsUpdatedSubscriptionManager       resources.SubscriptionManager
	gitHubEnterpriseSettingsUpdatedSubscriptionManager resources.SubscriptionManager
	gitHubEnterpriseSettingsClearedSubscriptionManager resources.SubscriptionManager
	licenseSettingsUpdatedSubscriptionManager          resources.SubscriptionManager
	apiKeyUpdatedSubscriptionManager                   resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalSettingsSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToDomainNameUpdatedEvents(updateHandler resources.DomainNameUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.domainNameUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromDomainNameUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.domainNameUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireDomainNameUpdatedEvent(domainName resources.DomainName) {
	subscriptionHandler.domainNameUpdatedSubscriptionManager.Fire(domainName)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToAuthenticationSettingsUpdatedEvents(updateHandler resources.AuthenticationSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.authenticationSettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromAuthenticationSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.authenticationSettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireAuthenticationSettingsUpdatedEvent(authenticationSettings *resources.AuthenticationSettings) {
	subscriptionHandler.authenticationSettingsUpdatedSubscriptionManager.Fire(authenticationSettings)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToRepositoryKeyPairUpdatedEvents(updateHandler resources.RepositoryKeyPairUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.repositoryKeyPairUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromRepositoryKeyPairUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.repositoryKeyPairUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireRepositoryKeyPairUpdatedEvent(keyPair *resources.RepositoryKeyPair) {
	subscriptionHandler.repositoryKeyPairUpdatedSubscriptionManager.Fire(keyPair)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToS3ExporterSettingsUpdatedEvents(updateHandler resources.S3ExporterSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.s3ExporterSettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromS3ExporterSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.s3ExporterSettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireS3ExporterSettingsUpdatedEvent(s3ExporterSettings *resources.S3ExporterSettings) {
	subscriptionHandler.s3ExporterSettingsUpdatedSubscriptionManager.Fire(s3ExporterSettings)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToS3ExporterSettingsClearedEvents(updateHandler resources.S3ExporterSettingsClearedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.s3ExporterSettingsClearedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromS3ExporterSettingsClearedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.s3ExporterSettingsClearedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireS3ExporterSettingsClearedEvent() {
	subscriptionHandler.s3ExporterSettingsClearedSubscriptionManager.Fire()
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToHipChatSettingsUpdatedEvents(updateHandler resources.HipChatSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.hipChatSettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromHipChatSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.hipChatSettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireHipChatSettingsUpdatedEvent(hipChatSettings *resources.HipChatSettings) {
	subscriptionHandler.hipChatSettingsUpdatedSubscriptionManager.Fire(hipChatSettings)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToHipChatSettingsClearedEvents(updateHandler resources.HipChatSettingsClearedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.hipChatSettingsClearedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromHipChatSettingsClearedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.hipChatSettingsClearedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireHipChatSettingsClearedEvent() {
	subscriptionHandler.hipChatSettingsClearedSubscriptionManager.Fire()
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToCookieStoreKeysUpdatedEvents(updateHandler resources.CookieStoreKeysUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.cookieStoreKeysUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCookieStoreKeysUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.cookieStoreKeysUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCookieStoreKeysUpdatedEvent(keys *resources.CookieStoreKeys) {
	subscriptionHandler.cookieStoreKeysUpdatedSubscriptionManager.Fire(keys)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSmtpServerSettingsUpdatedEvents(updateHandler resources.SmtpServerSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.smtpServerSettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSmtpServerSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.smtpServerSettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireSmtpServerSettingsUpdatedEvent(auth *resources.SmtpServerSettings) {
	subscriptionHandler.smtpServerSettingsUpdatedSubscriptionManager.Fire(auth)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToGitHubEnterpriseSettingsUpdatedEvents(updateHandler resources.GitHubEnterpriseSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.gitHubEnterpriseSettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromGitHubEnterpriseSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.gitHubEnterpriseSettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireGitHubEnterpriseSettingsUpdatedEvent(gitHubEnterpriseSettings *resources.GitHubEnterpriseSettings) {
	subscriptionHandler.gitHubEnterpriseSettingsUpdatedSubscriptionManager.Fire(gitHubEnterpriseSettings)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToGitHubEnterpriseSettingsClearedEvents(updateHandler resources.GitHubEnterpriseSettingsClearedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.gitHubEnterpriseSettingsClearedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromGitHubEnterpriseSettingsClearedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.gitHubEnterpriseSettingsClearedSubscriptionManager.Remove(subscriptionId)
}
func (subscriptionHandler *SubscriptionHandler) FireGitHubEnterpriseSettingsClearedEvent() {
	subscriptionHandler.gitHubEnterpriseSettingsClearedSubscriptionManager.Fire()
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToLicenseSettingsUpdatedEvents(updateHandler resources.LicenseSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.apiKeyUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromLicenseSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.apiKeyUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireLicenseSettingsUpdatedEvent(licenseSettings *resources.LicenseSettings) {
	subscriptionHandler.apiKeyUpdatedSubscriptionManager.Fire(licenseSettings)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToApiKeyUpdatedEvents(updateHandler resources.ApiKeyUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.apiKeyUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromApiKeyUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.apiKeyUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireApiKeyUpdatedEvent(key resources.ApiKey) {
	subscriptionHandler.apiKeyUpdatedSubscriptionManager.Fire(key)
}
