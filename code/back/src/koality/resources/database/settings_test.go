package database

import (
	"bytes"
	"koality/resources"
	"testing"
	"time"
)

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
	_, err = connection.Settings.Subscription.SubscribeToRepositoryKeyPairUpdatedEvents(keyPairUpdatedHandler)
	if err != nil {
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
	var keyUpdatedEventKey *resources.ApiKey
	keyUpdatedHandler := func(key *resources.ApiKey) {
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

	if keyUpdatedEventKey.Key != apiKey.Key {
		test.Fatal("Bad apiKey.Key in api key updated event")
	}

	apiKey2, err := connection.Settings.Read.GetApiKey()
	if err != nil {
		test.Fatal(err)
	}

	if apiKey.Key != apiKey2.Key {
		test.Fatal("Key mismatch")
	}

	apiKey3, err := connection.Settings.Update.ResetApiKey()
	if err != nil {
		test.Fatal(err)
	}

	if apiKey.Key == apiKey3.Key {
		test.Fatal("Expected new Key")
	}
}
