package resources

type DomainName string

func (domainName DomainName) String() string {
	return string(domainName)
}

type RepositoryKeyPair struct {
	PrivateKey string
	PublicKey  string
}

type S3ExporterSettings struct {
	AccessKey  string
	SecretKey  string
	BucketName string
}

type CookieStoreKeys struct {
	Authentication []byte
	Encryption     []byte
}

type SmtpServerSettings struct {
	Hostname string
	Port     uint16
	Auth     SmtpAuthSettings
}

type SmtpAuthSettings struct {
	Plain   *SmtpPlainAuthSettings
	CramMd5 *SmtpCramMd5AuthSettings
	Login   *SmtpLoginAuthSettings
}

type SmtpPlainAuthSettings struct {
	Identity string
	Username string
	Password string
	Host     string
}

type SmtpCramMd5AuthSettings struct {
	Username string
	Secret   string
}

type SmtpLoginAuthSettings struct {
	Username string
	Password string
}

type GitHubEnterpriseSettings struct {
	BaseUrl           string
	OAuthClientId     string
	OAuthClientSecret string
}

type ApiKey string

func (apiKey ApiKey) String() string {
	return string(apiKey)
}

type SettingsHandler struct {
	Read         SettingsReadHandler
	Update       SettingsUpdateHandler
	Delete       SettingsDeleteHandler
	Subscription SettingsSubscriptionHandler
}

type SettingsReadHandler interface {
	GetDomainName() (DomainName, error)
	GetRepositoryKeyPair() (*RepositoryKeyPair, error)
	GetS3ExporterSettings() (*S3ExporterSettings, error)
	GetCookieStoreKeys() (*CookieStoreKeys, error)
	GetSmtpServerSettings() (*SmtpServerSettings, error)
	GetGitHubEnterpriseSettings() (*GitHubEnterpriseSettings, error)
	GetApiKey() (ApiKey, error)
}

type SettingsUpdateHandler interface {
	SetDomainName(domainName string) (DomainName, error)
	ResetRepositoryKeyPair() (*RepositoryKeyPair, error)
	SetS3ExporterSettings(accessKey, secretKey, bucketName string) (*S3ExporterSettings, error)
	ResetCookieStoreKeys() (*CookieStoreKeys, error)
	SetSmtpAuthPlain(hostname string, port uint16, identity, username, password, host string) (*SmtpServerSettings, error)
	SetSmtpAuthCramMd5(hostname string, port uint16, username, secret string) (*SmtpServerSettings, error)
	SetSmtpAuthLogin(hostname string, port uint16, username, password string) (*SmtpServerSettings, error)
	SetGitHubEnterpriseSettings(baseUrl, oAuthClientId, oAuthClientSecret string) (*GitHubEnterpriseSettings, error)
	ResetApiKey() (ApiKey, error)
}

type SettingsDeleteHandler interface {
	ClearS3ExporterSettings() error
	ClearGitHubEnterpriseSettings() error
}

type DomainNameUpdatedHandler func(domainName DomainName)
type RepositoryKeyPairUpdatedHandler func(keyPair *RepositoryKeyPair)
type S3ExporterSettingsUpdatedHandler func(s3Settings *S3ExporterSettings)
type S3ExporterSettingsClearedHandler func()
type CookieStoreKeysUpdatedHandler func(keys *CookieStoreKeys)
type SmtpServerSettingsUpdatedHandler func(auth *SmtpServerSettings)
type GitHubEnterpriseSettingsUpdatedHandler func(gitHubEnterpriseSettings *GitHubEnterpriseSettings)
type GitHubEnterpriseSettingsClearedHandler func()
type ApiKeyUpdatedHandler func(key ApiKey)

type SettingsSubscriptionHandler interface {
	SubscribeToDomainNameUpdatedEvents(updateHandler DomainNameUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromDomainNameUpdatedEvents(subscriptionId SubscriptionId) error
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
	SubscribeToGitHubEnterpriseSettingsUpdatedEvents(updateHandler GitHubEnterpriseSettingsUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubEnterpriseSettingsUpdatedEvents(subscriptionId SubscriptionId) error
	SubscribeToGitHubEnterpriseSettingsClearedEvents(updateHandler GitHubEnterpriseSettingsClearedHandler) (SubscriptionId, error)
	UnsubscribeFromGitHubEnterpriseSettingsClearedEvents(subscriptionId SubscriptionId) error
	SubscribeToApiKeyUpdatedEvents(updateHandler ApiKeyUpdatedHandler) (SubscriptionId, error)
	UnsubscribeFromApiKeyUpdatedEvents(subscriptionId SubscriptionId) error
}

type InternalSettingsSubscriptionHandler interface {
	FireDomainNameUpdatedEvent(domainName DomainName)
	FireRepositoryKeyPairUpdatedEvent(keyPair *RepositoryKeyPair)
	FireS3ExporterSettingsUpdatedEvent(s3ExporterSettings *S3ExporterSettings)
	FireS3ExporterSettingsClearedEvent()
	FireCookieStoreKeysUpdatedEvent(keys *CookieStoreKeys)
	FireSmtpServerSettingsUpdatedEvent(auth *SmtpServerSettings)
	FireGitHubEnterpriseSettingsUpdatedEvent(gitHubEnterpriseSettings *GitHubEnterpriseSettings)
	FireGitHubEnterpriseSettingsClearedEvent()
	FireApiKeyUpdatedEvent(key ApiKey)
	SettingsSubscriptionHandler
}
