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

type CookieStoreKeys struct {
	Authentication []byte `json:"authentication"`
	Encryption     []byte `json:"encryption"`
}

type SmtpServerSettings struct {
	Hostname string           `json:"hostname"`
	Port     uint16           `json:"port"`
	Auth     SmtpAuthSettings `json:"auth"`
}

type SmtpAuthSettings struct {
	Plain   *SmtpPlainAuthSettings   `json:"plain"`
	CramMd5 *SmtpCramMd5AuthSettings `json:"cram-md5"`
	Login   *SmtpLoginAuthSettings   `json:"login"`
}

type SmtpPlainAuthSettings struct {
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

type SmtpCramMd5AuthSettings struct {
	Username string `json:"username"`
	Secret   string `json:"secret"`
}

type SmtpLoginAuthSettings struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ApiKey struct {
	Key string `json:"key"`
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
	GetCookieStoreKeys() (*CookieStoreKeys, error)
	GetSmtpServerSettings() (*SmtpServerSettings, error)
	GetApiKey() (*ApiKey, error)
}

type SettingsUpdateHandler interface {
	ResetRepositoryKeyPair() (*RepositoryKeyPair, error)
	SetS3ExporterSettings(accessKey, secretKey, bucketName string) (*S3ExporterSettings, error)
	ResetCookieStoreKeys() (*CookieStoreKeys, error)
	SetSmtpAuthPlain(hostname string, port uint16, identity, username, password, host string) (*SmtpServerSettings, error)
	SetSmtpAuthCramMd5(hostname string, port uint16, username, secret string) (*SmtpServerSettings, error)
	SetSmtpAuthLogin(hostname string, port uint16, username, password string) (*SmtpServerSettings, error)
	ResetApiKey() (*ApiKey, error)
}

type SettingsDeleteHandler interface {
	ClearS3ExporterSettings() error
}

type RepositoryKeyPairUpdatedHandler func(keyPair *RepositoryKeyPair)
type S3ExporterSettingsUpdatedHandler func(s3Settings *S3ExporterSettings)
type S3ExporterSettingsClearedHandler func()
type CookieStoreKeysUpdatedHandler func(keys *CookieStoreKeys)
type SmtpServerSettingsUpdatedHandler func(auth *SmtpServerSettings)
type ApiKeyUpdatedHandler func(key *ApiKey)

type SettingsSubscriptionHandler interface {
	SubscribeToRepositoryKeyPairUpdatedEvents(updateHandler RepositoryKeyPairUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromRepositoryKeyPairUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToS3ExporterSettingsUpdatedEvents(updateHandler S3ExporterSettingsUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromS3ExporterSettingsUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToS3ExporterSettingsClearedEvents(updateHandler S3ExporterSettingsClearedHandler) (SubscriptionId, error)
	UnsubscribeFromS3ExporterSettingsClearedEvents(subscriptionId SubscriptionId) error
	SubscribeToCookieStoreKeysUpdatedEvents(updateHandler CookieStoreKeysUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromCookieStoreKeysUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToSmtpServerSettingsUpdatedEvents(updateHandler SmtpServerSettingsUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromSmtpServerSettingsUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToApiKeyUpdatedEvents(updateHandler ApiKeyUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromApiKeyUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalSettingsSubscriptionHandler interface {
	FireRepositoryKeyPairUpdatedEvent(keyPair *RepositoryKeyPair)
	FireS3ExporterSettingsUpdatedEvent(s3ExporterSettings *S3ExporterSettings)
	FireS3ExporterSettingsClearedEvent()
	FireCookieStoreKeysUpdatedEvent(keys *CookieStoreKeys)
	FireSmtpServerSettingsUpdatedEvent(auth *SmtpServerSettings)
	FireApiKeyUpdatedEvent(key *ApiKey)
	SettingsSubscriptionHandler
}
