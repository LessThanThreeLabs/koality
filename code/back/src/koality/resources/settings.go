package resources

type RepositoryKeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type SettingsHandler struct {
	Read         SettingsReadHandler
	Update       SettingsUpdateHandler
	Subscription SettingsSubscriptionHandler
}

type SettingsReadHandler interface {
	GetRepositoryKeyPair() (*RepositoryKeyPair, error)
}

type SettingsUpdateHandler interface {
	ResetRepositoryKeyPair() (*RepositoryKeyPair, error)
}

type RepositoryKeyPairUpdatedHandler func(keyPair *RepositoryKeyPair)

type SettingsSubscriptionHandler interface {
	SubscribeToRepositoryKeyPairUpdatedEvents(updateHandler RepositoryKeyPairUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromRepositoryKeyPairUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalSettingsSubscriptionHandler interface {
	FireRepositoryKeyPairUpdatedEvent(keyPair *RepositoryKeyPair)
	SettingsSubscriptionHandler
}

type NoSuchSettingError struct {
	Message string
}

func (err NoSuchSettingError) Error() string {
	return err.Message
}
