package amqp

import (
	"github.com/streadway/amqp"
	"sync"
)

const (
	amqpUri = "amqp://localhost:5672/"
)

var (
	initializationSync sync.Once
	sendConnection     *amqp.Connection
	receiveConnection  *amqp.Connection
)

func createConnections() error {
	var err error

	sendConnection, err = amqp.Dial(amqpUri)
	if err != nil {
		return err
	}

	receiveConnection, err = amqp.Dial(amqpUri)
	if err != nil {
		return err
	}

	go handleConnectionClose()

	return nil
}

func handleConnectionClose() {
	sendConnectionClose := make(chan *amqp.Error)
	sendConnection.NotifyClose(sendConnectionClose)

	receiveConnectionClose := make(chan *amqp.Error)
	receiveConnection.NotifyClose(receiveConnectionClose)

	select {
	case <-sendConnectionClose:
	case <-receiveConnectionClose:
	}

	panic("Lost connection to RabbitMQ")
}

func initializeAmqp() {
	err := createConnections()
	if err != nil {
		panic(err)
	}
}

func GetSendConnection() *amqp.Connection {
	initializationSync.Do(initializeAmqp)
	return sendConnection
}

func GetReceiveConnection() *amqp.Connection {
	initializationSync.Do(initializeAmqp)
	return receiveConnection
}
