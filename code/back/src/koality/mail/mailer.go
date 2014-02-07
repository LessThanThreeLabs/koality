package mail

import (
	"koality/resources"
)

type Mailer interface {
	SendMail(fromAddress string, replyTo, toAddresses []string, subject, body string) error
	SubscribeToEvents(resourcesConnection *resources.Connection) error
	UnsubscribeFromEvents(resourcesConnection *resources.Connection) error
}
