package mail

import (
	"fmt"
	"koality/resources"
	"net/smtp"
	"sync"
)

type Mailer struct {
	smtpServerSettings *resources.SmtpServerSettings
	rwLock             *sync.RWMutex
	subscriptionId     resources.SubscriptionId
}

func NewMailer(smtpServerSettings *resources.SmtpServerSettings) *Mailer {
	return &Mailer{
		smtpServerSettings: smtpServerSettings,
		rwLock:             new(sync.RWMutex),
	}
}

func (mailer *Mailer) SendMail(fromAddress string, toAddresses []string, subject, body string) error {
	auth, err := mailer.getAuth()
	if err != nil {
		return err
	}

	serverAddress := fmt.Sprintf("%s:%d", mailer.smtpServerSettings.Hostname, mailer.smtpServerSettings.Port)
	return smtp.SendMail(serverAddress, auth, fromAddress, toAddresses, mailer.formatMessage(fromAddress, toAddresses, subject, body))
}

func (mailer *Mailer) getAuth() (smtp.Auth, error) {
	mailer.rwLock.RLock()
	defer mailer.rwLock.RUnlock()

	if mailer.smtpServerSettings == nil {
		return nil, NoAuthProvidedError{}
	}

	authSettings := mailer.smtpServerSettings.Auth

	if authSettings.Plain != nil {
		return smtp.PlainAuth(authSettings.Plain.Identity, authSettings.Plain.Username,
			authSettings.Plain.Password, authSettings.Plain.Host), nil
	} else if authSettings.CramMd5 != nil {
		return smtp.CRAMMD5Auth(authSettings.CramMd5.Username, authSettings.CramMd5.Secret), nil
	} else if authSettings.Login != nil {
		return &loginAuth{authSettings.Login.Username, authSettings.Login.Password}, nil
	} else {
		return nil, NoAuthProvidedError{}
	}
}

func (mailer *Mailer) formatMessage(fromAddress string, toAddresses []string, subject, body string) []byte {
	message := fmt.Sprintf("From: %s\r\n", fromAddress)
	for _, toAddress := range toAddresses {
		message += fmt.Sprintf("To: %s\r\n", toAddress)
	}

	message += fmt.Sprintf("Subject: %s\r\nMIME-version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s", subject, body)
	fmt.Printf("%q\n", message)
	return []byte(message)
}

func (mailer *Mailer) updateSmtpServerSettings(smtpServerSettings *resources.SmtpServerSettings) {
	mailer.rwLock.Lock()
	defer mailer.rwLock.Unlock()

	*mailer.smtpServerSettings = *smtpServerSettings
}

func (mailer *Mailer) SubscribeToEvents(resourcesConnection *resources.Connection) error {
	mailer.rwLock.Lock()
	defer mailer.rwLock.Unlock()

	if mailer.subscriptionId != 0 {
		return fmt.Errorf("Mailer already subscribed to events, subscriptionId: %d", mailer.subscriptionId)
	}

	subscriptionId, err := resourcesConnection.Settings.Subscription.SubscribeToSmtpServerSettingsUpdatedEvents(mailer.updateSmtpServerSettings)
	if err != nil {
		return err
	}

	mailer.subscriptionId = subscriptionId
	return nil
}

func (mailer *Mailer) UnsubscribeFromEvents(resourcesConnection *resources.Connection) error {
	mailer.rwLock.Lock()
	defer mailer.rwLock.Unlock()

	if mailer.subscriptionId == 0 {
		return fmt.Errorf("Mailer not subscribed to events")
	}

	if err := resourcesConnection.Settings.Subscription.UnsubscribeFromSmtpServerSettingsUpdatedEvents(mailer.subscriptionId); err != nil {
		return err
	}

	mailer.subscriptionId = 0
	return nil
}

type NoAuthProvidedError struct{}

func (err NoAuthProvidedError) Error() string {
	return "No smtp authentication method provided"
}
