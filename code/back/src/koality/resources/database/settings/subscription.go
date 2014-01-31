package settings

import (
	"koality/resources"
)

type SubscriptionHandler struct {
	repositoryKeyPairUpdatedSubscriptionManager  resources.SubscriptionManager
	s3ExporterSettingsUpdatedSubscriptionManager resources.SubscriptionManager
	s3ExporterSettingsClearedSubscriptionManager resources.SubscriptionManager
	cookieStoreKeysUpdatedSubscriptionManager    resources.SubscriptionManager
	smtpAuthSettingsUpdatedSubscriptionManager   resources.SubscriptionManager
	apiKeyUpdatedSubscriptionManager             resources.SubscriptionManager
}

func NewInternalSubscriptionHandler() (resources.InternalSettingsSubscriptionHandler, error) {
	return &SubscriptionHandler{}, nil
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

func (subscriptionHandler *SubscriptionHandler) SubscribeToCookieStoreKeysUpdatedEvents(updateHandler resources.CookieStoreKeysUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.cookieStoreKeysUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromCookieStoreKeysUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.cookieStoreKeysUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireCookieStoreKeysUpdatedEvent(keys *resources.CookieStoreKeys) {
	subscriptionHandler.cookieStoreKeysUpdatedSubscriptionManager.Fire(keys)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToSmtpAuthSettingsUpdatedEvents(updateHandler resources.SmtpAuthSettingsUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.smtpAuthSettingsUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromSmtpAuthSettingsUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.smtpAuthSettingsUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireSmtpAuthSettingsUpdatedEvent(auth *resources.SmtpAuthSettings) {
	subscriptionHandler.smtpAuthSettingsUpdatedSubscriptionManager.Fire(auth)
}

func (subscriptionHandler *SubscriptionHandler) SubscribeToApiKeyUpdatedEvents(updateHandler resources.ApiKeyUpdatedHandler) (resources.SubscriptionId, error) {
	return subscriptionHandler.apiKeyUpdatedSubscriptionManager.Add(updateHandler)
}

func (subscriptionHandler *SubscriptionHandler) UnsubscribeFromApiKeyUpdatedEvents(subscriptionId resources.SubscriptionId) error {
	return subscriptionHandler.apiKeyUpdatedSubscriptionManager.Remove(subscriptionId)
}

func (subscriptionHandler *SubscriptionHandler) FireApiKeyUpdatedEvent(key *resources.ApiKey) {
	subscriptionHandler.apiKeyUpdatedSubscriptionManager.Fire(key)
}
