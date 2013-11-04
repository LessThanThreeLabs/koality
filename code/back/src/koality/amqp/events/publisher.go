package events

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack"
	kamqp "koality/amqp"
	"time"
	"unicode/utf8"
)

type Publisher struct {
	route       string
	sendChannel *amqp.Channel
}

func NewPublisher(route string) *Publisher {
	createExchanges()

	sendChannel, err := kamqp.GetSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	publisher := Publisher{
		route:       route,
		sendChannel: sendChannel,
	}

	return &publisher
}

func (publisher *Publisher) checkEventIsValid(event *Event) error {
	for _, arg := range event.Args {
		if !utf8.ValidString(fmt.Sprint(arg)) {
			return EventError{Message: "Event argument contains illegal character"}
		}
	}
	return nil
}

func (publisher *Publisher) Publish(event *Event) error {
	err := publisher.checkEventIsValid(event)
	if err != nil {
		return err
	}

	buffer, err := msgpack.Marshal(event)
	if err != nil {
		return err
	}

	publishing := amqp.Publishing{
		Body:            buffer,
		ContentType:     "application/x-msgpack",
		ContentEncoding: "binary",
		DeliveryMode:    amqp.Transient,
		Timestamp:       time.Now(),
	}

	err = publisher.sendChannel.Publish(exchangeName, publisher.route, exchangeMandatory, exchangeImmediate, publishing)
	if err != nil {
		return err
	}

	return nil
}
