package resources

type RepositoryKeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type S3ExporterSettings struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	BucketName string `json:"bucketName"`
}

type SettingsHandler struct {
	Read         SettingsReadHandler
	Update       SettingsUpdateHandler
	Delete       SettingsDeleteHandler
	Subscription SettingsSubscriptionHandler
}

type SettingsReadHandler interface {
	GetRepositoryKeyPair() (*RepositoryKeyPair, error)
	GetS3ExporterSettings() (*S3ExporterSettings, error)
}

type SettingsUpdateHandler interface {
	ResetRepositoryKeyPair() (*RepositoryKeyPair, error)
	SetS3ExporterSettings(accessKey, secretKey, bucketName string) (*S3ExporterSettings, error)
}

type SettingsDeleteHandler interface {
	ClearS3ExporterSettings() error
}

type RepositoryKeyPairUpdatedHandler func(keyPair *RepositoryKeyPair)
type S3ExporterSettingsUpdatedHandler func(s3Settings *S3ExporterSettings)
type S3ExporterSettingsClearedHandler func()

type SettingsSubscriptionHandler interface {
	SubscribeToRepositoryKeyPairUpdatedEvents(updateHandler RepositoryKeyPairUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromRepositoryKeyPairUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToS3ExporterSettingsUpdatedEvents(updateHandler S3ExporterSettingsUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromS3ExporterSettingsUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToS3ExporterSettingsClearedEvents(updateHandler S3ExporterSettingsClearedHandler) (SubscriptionId, error)
	UnsubscribeFromS3ExporterSettingsClearedEvents(subscriptionId SubscriptionId) error
}

type InternalSettingsSubscriptionHandler interface {
	FireRepositoryKeyPairUpdatedEvent(keyPair *RepositoryKeyPair)
	FireS3ExporterSettingsUpdatedEvent(s3ExporterSettings *S3ExporterSettings)
	FireS3ExporterSettingsClearedEvent()
	SettingsSubscriptionHandler
}

type NoSuchSettingError struct {
	Message string
}

func (err NoSuchSettingError) Error() string {
	return err.Message
}
