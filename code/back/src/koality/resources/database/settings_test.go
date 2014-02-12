package database

import (
	"bytes"
	"koality/resources"
	"testing"
	"time"
)

func TestSettingsDomainName(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	domainNameEventReceived := make(chan bool, 1)
	var domainNameUpdatedEventDomainName resources.DomainName
	domainNameUpdatedHandler := func(domainName resources.DomainName) {
		domainNameUpdatedEventDomainName = domainName
		domainNameEventReceived <- true
	}
	if _, err = connection.Settings.Subscription.SubscribeToDomainNameUpdatedEvents(domainNameUpdatedHandler); err != nil {
		test.Fatal(err)
	}

	domainName := "koality.yourcompany.com"

	domainNameSetting, err := connection.Settings.Update.SetDomainName(domainName)
	if err != nil {
		test.Fatal(err)
	}

	if domainNameSetting.String() != domainName {
		test.Fatalf("Domain name setting corrupted, was %s, expected %s", domainNameSetting, domainName)
	}

	select {
	case <-domainNameEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear domain name updated event")
	}

	if domainNameSetting != domainNameUpdatedEventDomainName {
		test.Fatal("Bad domainName in domain name updated event")
	}

	domainNameSetting2, err := connection.Settings.Read.GetDomainName()
	if err != nil {
		test.Fatal(err)
	}

	if domainNameSetting != domainNameSetting2 {
		test.Fatal("Received bad domain name setting from read method")
	}
}

func TestSettingsAuthenticationSettings(test *testing.T) {
	allowedDomainsEqual := func(domains1, domains2 []string) bool {
		if len(domains1) != len(domains2) {
			return false
		}
		for index := range domains1 {
			if domains1[index] != domains2[index] {
				return false
			}
		}
		return true
	}
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	authenticationSettingsEventReceived := make(chan bool, 1)
	var authenticationSettingsUpdatedEventSettings *resources.AuthenticationSettings
	authenticationSettingsUpdatedHandler := func(authenticationSettings *resources.AuthenticationSettings) {
		authenticationSettingsUpdatedEventSettings = authenticationSettings
		authenticationSettingsEventReceived <- true
	}
	if _, err = connection.Settings.Subscription.SubscribeToAuthenticationSettingsUpdatedEvents(authenticationSettingsUpdatedHandler); err != nil {
		test.Fatal(err)
	}

	manualAllowed := true
	googleAllowed := false
	allowedDomains := []string{"koality.yourcompany.com", "127.0.0.1", "local-host"}

	authenticationSettings, err := connection.Settings.Update.SetAuthenticationSettings(manualAllowed, googleAllowed, allowedDomains)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-authenticationSettingsEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear authentication settings updated event")
	}

	if authenticationSettingsUpdatedEventSettings.ManualAccountsAllowed != authenticationSettings.ManualAccountsAllowed {
		test.Fatal("Bad authenticationSettings.ManualAccountsAllowed in authentication settings updated event")
	} else if authenticationSettingsUpdatedEventSettings.GoogleAccountsAllowed != authenticationSettings.GoogleAccountsAllowed {
		test.Fatal("Bad authenticationSettings.GoogleAccountsAllowed in authentication settings updated event")
	} else if !allowedDomainsEqual(authenticationSettingsUpdatedEventSettings.AllowedDomains, authenticationSettings.AllowedDomains) {
		test.Fatal("Bad authenticationSettings.AllowedDomains in authentication settings updated event")
	}

	if authenticationSettings.ManualAccountsAllowed != manualAllowed {
		test.Fatal("ManualAccountsAllowed mismatch")
	} else if authenticationSettings.GoogleAccountsAllowed != googleAllowed {
		test.Fatal("GoogleAccountsAllowed mismatch")
	} else if !allowedDomainsEqual(authenticationSettings.AllowedDomains, allowedDomains) {
		test.Fatal("AllowedDomains mismatch")
	}

	authenticationSettings2, err := connection.Settings.Read.GetAuthenticationSettings()
	if err != nil {
		test.Fatal(err)
	}

	if authenticationSettings.ManualAccountsAllowed != authenticationSettings2.ManualAccountsAllowed {
		test.Fatal("Bad authenticationSettings.ManualAccountsAllowed in authentication settings updated event")
	} else if authenticationSettings.GoogleAccountsAllowed != authenticationSettings2.GoogleAccountsAllowed {
		test.Fatal("Bad authenticationSettings.GoogleAccountsAllowed in authentication settings updated event")
	} else if !allowedDomainsEqual(authenticationSettings.AllowedDomains, authenticationSettings2.AllowedDomains) {
		test.Fatal("Bad authenticationSettings.AllowedDomains in authentication settings updated event")
	}
}

func TestSettingsResetRepositoryKeyPair(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	repositoryKeyPair, err := connection.Settings.Read.GetRepositoryKeyPair()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		test.Fatal("Expected repository key pair to have default")
	} else if err != nil {
		test.Fatal(err)
	}

	keyPairUpdatedEventReceived := make(chan bool, 1)
	var keyPairUpdatedEventKeyPair *resources.RepositoryKeyPair
	keyPairUpdatedHandler := func(keyPair *resources.RepositoryKeyPair) {
		keyPairUpdatedEventKeyPair = keyPair
		keyPairUpdatedEventReceived <- true
	}
	if _, err = connection.Settings.Subscription.SubscribeToRepositoryKeyPairUpdatedEvents(keyPairUpdatedHandler); err != nil {
		test.Fatal(err)
	}

	repositoryKeyPair, err = connection.Settings.Update.ResetRepositoryKeyPair()
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-keyPairUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear repository key pair updated event")
	}

	if keyPairUpdatedEventKeyPair.PrivateKey != repositoryKeyPair.PrivateKey {
		test.Fatal("Bad repositoryKeyPair.PrivateKey in repository key pair updated event")
	} else if keyPairUpdatedEventKeyPair.PublicKey != repositoryKeyPair.PublicKey {
		test.Fatal("Bad repositoryKeyPair.PublicKey in repository key pair updated event")
	}

	repositoryKeyPair2, err := connection.Settings.Read.GetRepositoryKeyPair()
	if err != nil {
		test.Fatal(err)
	}

	if repositoryKeyPair.PrivateKey != repositoryKeyPair2.PrivateKey {
		test.Fatal("PrivateKey mismatch")
	} else if repositoryKeyPair.PublicKey != repositoryKeyPair2.PublicKey {
		test.Fatal("PublicKey mismatch")
	}

	repositoryKeyPair3, err := connection.Settings.Update.ResetRepositoryKeyPair()
	if err != nil {
		test.Fatal(err)
	}

	if repositoryKeyPair.PrivateKey == repositoryKeyPair3.PrivateKey {
		test.Fatal("Expected new PrivateKey")
	} else if repositoryKeyPair.PublicKey == repositoryKeyPair3.PublicKey {
		test.Fatal("Expected new PublicKey")
	}
}

func TestSettingsS3ExporterSettings(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	s3ExporterSettingsUpdatedEventReceived := make(chan bool, 1)
	var s3ExporterSettingsUpdatedEventSettings *resources.S3ExporterSettings
	s3ExporterSettingsUpdatedHandler := func(s3ExporterSettings *resources.S3ExporterSettings) {
		s3ExporterSettingsUpdatedEventSettings = s3ExporterSettings
		s3ExporterSettingsUpdatedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToS3ExporterSettingsUpdatedEvents(s3ExporterSettingsUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	s3ExporterSettingsClearedEventReceived := make(chan bool, 1)
	s3ExporterSettingsClearedHandler := func() {
		s3ExporterSettingsClearedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToS3ExporterSettingsClearedEvents(s3ExporterSettingsClearedHandler)
	if err != nil {
		test.Fatal(err)
	}

	accessKey := "aaaabbbbccccddddeeee"
	secretKey := "0000111122223333444455556666777788889999"
	bucketName := "some-bucket-name"
	s3ExporterSettings, err := connection.Settings.Update.SetS3ExporterSettings(accessKey, secretKey, bucketName)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-s3ExporterSettingsUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear s3 exporter settings updated event")
	}

	if s3ExporterSettingsUpdatedEventSettings.AccessKey != s3ExporterSettings.AccessKey {
		test.Fatal("Bad s3ExporterSettings.AccessKey in s3 exporter settings updated event")
	} else if s3ExporterSettingsUpdatedEventSettings.SecretKey != s3ExporterSettings.SecretKey {
		test.Fatal("Bad s3ExporterSettings.SecretKey in s3 exporter settings updated event")
	} else if s3ExporterSettingsUpdatedEventSettings.BucketName != s3ExporterSettings.BucketName {
		test.Fatal("Bad s3ExporterSettings.BucketName in s3 exporter settings updated event")
	}

	if s3ExporterSettings.AccessKey != accessKey {
		test.Fatal("AccessKey mismatch")
	} else if s3ExporterSettings.SecretKey != secretKey {
		test.Fatal("SecretKey mismatch")
	} else if s3ExporterSettings.BucketName != bucketName {
		test.Fatal("BucketName mismatch")
	}

	s3ExporterSettings2, err := connection.Settings.Read.GetS3ExporterSettings()
	if err != nil {
		test.Fatal(err)
	}

	if s3ExporterSettings.AccessKey != s3ExporterSettings2.AccessKey {
		test.Fatal("AccessKey mismatch")
	} else if s3ExporterSettings.SecretKey != s3ExporterSettings2.SecretKey {
		test.Fatal("SecretKey mismatch")
	} else if s3ExporterSettings.BucketName != s3ExporterSettings2.BucketName {
		test.Fatal("BucketName mismatch")
	}

	err = connection.Settings.Delete.ClearS3ExporterSettings()
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-s3ExporterSettingsClearedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear s3 exporter settings cleared event")
	}

	_, err = connection.Settings.Read.GetS3ExporterSettings()
	if _, ok := err.(resources.NoSuchSettingError); !ok {
		test.Fatal("Expected NoSuchSettingError when trying to get s3 exporter settings that have been cleared")
	}
}

func TestSettingsHipChatSettings(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	hipChatSettingsUpdatedEventReceived := make(chan bool, 1)
	var hipChatSettingsUpdatedEventSettings *resources.HipChatSettings
	hipChatSettingsUpdatedHandler := func(hipChatSettings *resources.HipChatSettings) {
		hipChatSettingsUpdatedEventSettings = hipChatSettings
		hipChatSettingsUpdatedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToHipChatSettingsUpdatedEvents(hipChatSettingsUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	hipChatSettingsClearedEventReceived := make(chan bool, 1)
	hipChatSettingsClearedHandler := func() {
		hipChatSettingsClearedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToHipChatSettingsClearedEvents(hipChatSettingsClearedHandler)
	if err != nil {
		test.Fatal(err)
	}

	authenticationToken := "abcdefghijklmnopqrstuvwxyzabcd"
	rooms := []string{"engineering", "koalas"}
	notifyOn := "all"
	hipChatSettings, err := connection.Settings.Update.SetHipChatSettings(authenticationToken, rooms, notifyOn)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-hipChatSettingsUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear hipchat settings updated event")
	}

	if hipChatSettingsUpdatedEventSettings.AuthenticationToken != hipChatSettings.AuthenticationToken {
		test.Fatal("Bad hipChatSettings.AuthenticationToken in hipchat settings updated event")
	} else if hipChatSettingsUpdatedEventSettings.Rooms[0] != hipChatSettings.Rooms[0] {
		test.Fatal("Bad hipChatSettings.Rooms in hipchat settings updated event")
	} else if hipChatSettingsUpdatedEventSettings.Rooms[1] != hipChatSettings.Rooms[1] {
		test.Fatal("Bad hipChatSettings.Rooms in hipchat settings updated event")
	} else if hipChatSettingsUpdatedEventSettings.NotifyOn != hipChatSettings.NotifyOn {
		test.Fatal("Bad hipChatSettings.NotifyOn in hipchat settings updated event")
	}

	if hipChatSettings.AuthenticationToken != authenticationToken {
		test.Fatal("AuthenticationToken mismatch")
	} else if hipChatSettings.Rooms[0] != rooms[0] {
		test.Fatal("Rooms mismatch")
	} else if hipChatSettings.Rooms[1] != rooms[1] {
		test.Fatal("Rooms mismatch")
	} else if hipChatSettings.NotifyOn != notifyOn {
		test.Fatal("NotifyOn mismatch")
	}

	hipChatSettings2, err := connection.Settings.Read.GetHipChatSettings()
	if err != nil {
		test.Fatal(err)
	}

	if hipChatSettings.AuthenticationToken != hipChatSettings2.AuthenticationToken {
		test.Fatal("AuthenticationToken mismatch")
	} else if hipChatSettings.Rooms[0] != hipChatSettings2.Rooms[0] {
		test.Fatal("Rooms mismatch")
	} else if hipChatSettings.Rooms[1] != hipChatSettings2.Rooms[1] {
		test.Fatal("Rooms mismatch")
	} else if hipChatSettings.NotifyOn != hipChatSettings2.NotifyOn {
		test.Fatal("NotifyOn mismatch")
	}

	err = connection.Settings.Delete.ClearHipChatSettings()
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-hipChatSettingsClearedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear hipchat settings cleared event")
	}

	_, err = connection.Settings.Read.GetHipChatSettings()
	if _, ok := err.(resources.NoSuchSettingError); !ok {
		test.Fatal("Expected NoSuchSettingError when trying to get hipchat settings that have been cleared")
	}
}

func TestSettingsResetCookieStoreKeys(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	cookieStoreKeys, err := connection.Settings.Read.GetCookieStoreKeys()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		test.Fatal("Expected cookie store keys to have default")
	} else if err != nil {
		test.Fatal(err)
	}

	keysUpdatedEventReceived := make(chan bool, 1)
	var keysUpdatedEventKeys *resources.CookieStoreKeys
	keysUpdatedHandler := func(keys *resources.CookieStoreKeys) {
		keysUpdatedEventKeys = keys
		keysUpdatedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToCookieStoreKeysUpdatedEvents(keysUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	cookieStoreKeys, err = connection.Settings.Update.ResetCookieStoreKeys()
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-keysUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear cookie store keys updated event")
	}

	if bytes.Compare(keysUpdatedEventKeys.Authentication, cookieStoreKeys.Authentication) != 0 {
		test.Fatal("Bad cookieStoreKeys.Authentication in cookie store keys updated event")
	} else if bytes.Compare(keysUpdatedEventKeys.Encryption, cookieStoreKeys.Encryption) != 0 {
		test.Fatal("Bad cookieStoreKeys.Encryption in cookie store keys updated event")
	}

	cookieStoreKeys2, err := connection.Settings.Read.GetCookieStoreKeys()
	if err != nil {
		test.Fatal(err)
	}

	if bytes.Compare(cookieStoreKeys.Authentication, cookieStoreKeys2.Authentication) != 0 {
		test.Fatal("Authentication mismatch")
	} else if bytes.Compare(cookieStoreKeys.Encryption, cookieStoreKeys2.Encryption) != 0 {
		test.Fatal("Encryption mismatch")
	}

	cookieStoreKeys3, err := connection.Settings.Update.ResetCookieStoreKeys()
	if err != nil {
		test.Fatal(err)
	}

	if bytes.Compare(cookieStoreKeys.Authentication, cookieStoreKeys3.Authentication) == 0 {
		test.Fatal("Expected new Authentication")
	} else if bytes.Compare(cookieStoreKeys.Encryption, cookieStoreKeys3.Encryption) == 0 {
		test.Fatal("Expected new Encryption")
	}
}

func TestSettingsSmtpAuth(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	smtpServerSettingsUpdatedEventReceived := make(chan bool, 3)
	var smtpServerSettingsUpdatedEventSettings *resources.SmtpServerSettings
	smtpServerSettingsUpdatedHandler := func(smtpServerSettings *resources.SmtpServerSettings) {
		smtpServerSettingsUpdatedEventSettings = smtpServerSettings
		smtpServerSettingsUpdatedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToSmtpServerSettingsUpdatedEvents(smtpServerSettingsUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	hostname := "a.hostName"
	port := uint16(1234)
	plainIdentity := ""
	plainUsername := "bbland"
	plainPassword := "Ap@$$w0Rd!"
	plainHost := "smtp.gmail.com"
	smtpServerSettings, err := connection.Settings.Update.SetSmtpAuthPlain(hostname, port, plainIdentity, plainUsername, plainPassword, plainHost)
	if err != nil {
		test.Fatal(err)
	}

	if smtpServerSettings.Auth.CramMd5 != nil {
		test.Fatal("Bad smtp server settings, CramMd5 should be nil")
	} else if smtpServerSettings.Auth.Login != nil {
		test.Fatal("Bad smtp server settings, Login should be nil")
	}

	select {
	case <-smtpServerSettingsUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear smtp server settings updated event")
	}

	if smtpServerSettings.Hostname != smtpServerSettingsUpdatedEventSettings.Hostname {
		test.Fatal("Bad smtpServerSettings.Hostname in smtp server settings update event")
	} else if smtpServerSettings.Port != smtpServerSettingsUpdatedEventSettings.Port {
		test.Fatal("Bad smtpServerSettings.Port in smtp server settings update event")
	} else if smtpServerSettingsUpdatedEventSettings.Auth.CramMd5 != nil {
		test.Fatal("Bad smtp server settings event, CramMd5 should be nil")
	} else if smtpServerSettingsUpdatedEventSettings.Auth.Login != nil {
		test.Fatal("Bad smtp server settings event, Login should be nil")
	} else if smtpServerSettings.Auth.Plain.Identity != smtpServerSettingsUpdatedEventSettings.Auth.Plain.Identity {
		test.Fatal("Bad smtpServerSettings.Auth.Plain.Identity in smtp server settings update event")
	} else if smtpServerSettings.Auth.Plain.Username != smtpServerSettingsUpdatedEventSettings.Auth.Plain.Username {
		test.Fatal("Bad smtpServerSettings.Auth.Plain.Username in smtp server settings update event")
	} else if smtpServerSettings.Auth.Plain.Password != smtpServerSettingsUpdatedEventSettings.Auth.Plain.Password {
		test.Fatal("Bad smtpServerSettings.Auth.Plain.Password in smtp server settings update event")
	} else if smtpServerSettings.Auth.Plain.Host != smtpServerSettingsUpdatedEventSettings.Auth.Plain.Host {
		test.Fatal("Bad smtpServerSettings.Auth.Plain.Host in smtp server settings update event")
	}

	cramMd5Username := "a Username"
	cramMd5Secret := "$Up3r_sEcR3+"

	smtpServerSettings, err = connection.Settings.Update.SetSmtpAuthCramMd5(hostname, port, cramMd5Username, cramMd5Secret)
	if err != nil {
		test.Fatal(err)
	}

	if smtpServerSettings.Auth.Plain != nil {
		test.Fatal("Bad smtp server settings, Plain should be nil")
	} else if smtpServerSettings.Auth.Login != nil {
		test.Fatal("Bad smtp server settings, Login should be nil")
	}

	select {
	case <-smtpServerSettingsUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear smtp server settings updated event")
	}

	if smtpServerSettingsUpdatedEventSettings.Auth.Plain != nil {
		test.Fatal("Bad smtp server settings event, Plain should be nil")
	} else if smtpServerSettingsUpdatedEventSettings.Auth.Login != nil {
		test.Fatal("Bad smtp server settings event, Login should be nil")
	} else if smtpServerSettings.Auth.CramMd5.Username != smtpServerSettingsUpdatedEventSettings.Auth.CramMd5.Username {
		test.Fatal("Bad smtpServerSettings.Auth.CramMd5.Auth.Username in smtp server settings update event")
	} else if smtpServerSettings.Auth.CramMd5.Secret != smtpServerSettingsUpdatedEventSettings.Auth.CramMd5.Secret {
		test.Fatal("Bad smtpServerSettings.Auth.CramMd5.Secret in smtp server settings update event")
	}

	loginUsername := "a Username"
	loginPassword := "4n0T#3r_p4$Sw0rd"

	smtpServerSettings, err = connection.Settings.Update.SetSmtpAuthLogin(hostname, port, loginUsername, loginPassword)
	if err != nil {
		test.Fatal(err)
	}

	if smtpServerSettings.Auth.Plain != nil {
		test.Fatal("Bad smtp server settings, Plain should be nil")
	} else if smtpServerSettings.Auth.CramMd5 != nil {
		test.Fatal("Bad smtp server settings, CramMd5 should be nil")
	}

	select {
	case <-smtpServerSettingsUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear smtp server settings updated event")
	}

	if smtpServerSettingsUpdatedEventSettings.Auth.Plain != nil {
		test.Fatal("Bad smtp server settings event, Plain should be nil")
	} else if smtpServerSettingsUpdatedEventSettings.Auth.CramMd5 != nil {
		test.Fatal("Bad smtp server settings event, CramMd5 should be nil")
	} else if smtpServerSettings.Auth.Login.Username != smtpServerSettingsUpdatedEventSettings.Auth.Login.Username {
		test.Fatal("Bad smtpServerSettings.Auth.Login.Username in smtp server settings update event")
	} else if smtpServerSettings.Auth.Login.Password != smtpServerSettingsUpdatedEventSettings.Auth.Login.Password {
		test.Fatal("Bad smtpServerSettings.Auth.Login.Password in smtp server settings update event")
	}
}

func TestSettingsGitHubEnterprise(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	gitHubEnterpriseSettingsUpdatedEventReceived := make(chan bool, 1)
	var gitHubEnterpriseSettingsUpdatedEventSettings *resources.GitHubEnterpriseSettings
	gitHubEnterpriseSettingsUpdatedHandler := func(gitHubEnterpriseSettings *resources.GitHubEnterpriseSettings) {
		gitHubEnterpriseSettingsUpdatedEventSettings = gitHubEnterpriseSettings
		gitHubEnterpriseSettingsUpdatedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToGitHubEnterpriseSettingsUpdatedEvents(gitHubEnterpriseSettingsUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	gitHubEnterpriseSettingsClearedEventReceived := make(chan bool, 1)
	gitHubEnterpriseSettingsClearedHandler := func() {
		gitHubEnterpriseSettingsClearedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToGitHubEnterpriseSettingsClearedEvents(gitHubEnterpriseSettingsClearedHandler)
	if err != nil {
		test.Fatal(err)
	}

	uri := "aaaabbbbccccddddeeee"
	oAuthClientId := "0000111122223333444455556666777788889999"
	oAuthClientSecret := "some-bucket-name"
	gitHubEnterpriseSettings, err := connection.Settings.Update.SetGitHubEnterpriseSettings(uri, oAuthClientId, oAuthClientSecret)
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-gitHubEnterpriseSettingsUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear gitHub enterprise settings updated event")
	}

	if gitHubEnterpriseSettingsUpdatedEventSettings.BaseUri != gitHubEnterpriseSettings.BaseUri {
		test.Fatal("Bad gitHubEnterpriseSettings.BaseUri in gitHub enterprise settings updated event")
	} else if gitHubEnterpriseSettingsUpdatedEventSettings.OAuthClientId != gitHubEnterpriseSettings.OAuthClientId {
		test.Fatal("Bad gitHubEnterpriseSettings.OAuthClientId in gitHub enterprise settings updated event")
	} else if gitHubEnterpriseSettingsUpdatedEventSettings.OAuthClientSecret != gitHubEnterpriseSettings.OAuthClientSecret {
		test.Fatal("Bad gitHubEnterpriseSettings.OAuthClientSecret in gitHub enterprise settings updated event")
	}

	if gitHubEnterpriseSettings.BaseUri != uri {
		test.Fatal("Url mismatch")
	} else if gitHubEnterpriseSettings.OAuthClientId != oAuthClientId {
		test.Fatal("OAuthClientId mismatch")
	} else if gitHubEnterpriseSettings.OAuthClientSecret != oAuthClientSecret {
		test.Fatal("OAuthClientSecret mismatch")
	}

	gitHubEnterpriseSettings2, err := connection.Settings.Read.GetGitHubEnterpriseSettings()
	if err != nil {
		test.Fatal(err)
	}

	if gitHubEnterpriseSettings.BaseUri != gitHubEnterpriseSettings2.BaseUri {
		test.Fatal("Url mismatch")
	} else if gitHubEnterpriseSettings.OAuthClientId != gitHubEnterpriseSettings2.OAuthClientId {
		test.Fatal("OAuthClientId mismatch")
	} else if gitHubEnterpriseSettings.OAuthClientSecret != gitHubEnterpriseSettings2.OAuthClientSecret {
		test.Fatal("OAuthClientSecret mismatch")
	}

	err = connection.Settings.Delete.ClearGitHubEnterpriseSettings()
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-gitHubEnterpriseSettingsClearedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear gitHub enterprise settings cleared event")
	}

	_, err = connection.Settings.Read.GetGitHubEnterpriseSettings()
	if _, ok := err.(resources.NoSuchSettingError); !ok {
		test.Fatal("Expected NoSuchSettingError when trying to get gitHub enterprise settings that have been cleared")
	}
}

func TestSettingsApiKey(test *testing.T) {
	if err := PopulateDatabase(); err != nil {
		test.Fatal(err)
	}

	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()

	apiKey, err := connection.Settings.Read.GetApiKey()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		test.Fatal("Expected api key to have default")
	} else if err != nil {
		test.Fatal(err)
	}

	keyUpdatedEventReceived := make(chan bool, 1)
	var keyUpdatedEventKey resources.ApiKey
	keyUpdatedHandler := func(key resources.ApiKey) {
		keyUpdatedEventKey = key
		keyUpdatedEventReceived <- true
	}
	_, err = connection.Settings.Subscription.SubscribeToApiKeyUpdatedEvents(keyUpdatedHandler)
	if err != nil {
		test.Fatal(err)
	}

	apiKey, err = connection.Settings.Update.ResetApiKey()
	if err != nil {
		test.Fatal(err)
	}

	select {
	case <-keyUpdatedEventReceived:
	case <-time.After(10 * time.Second):
		test.Fatal("Failed to hear api key updated event")
	}

	if keyUpdatedEventKey != apiKey {
		test.Fatal("Bad apiKey in api key updated event")
	}

	apiKey2, err := connection.Settings.Read.GetApiKey()
	if err != nil {
		test.Fatal(err)
	}

	if apiKey != apiKey2 {
		test.Fatal("Key mismatch")
	}

	apiKey3, err := connection.Settings.Update.ResetApiKey()
	if err != nil {
		test.Fatal(err)
	}

	if apiKey == apiKey3 {
		test.Fatal("Expected new Key")
	}
}
