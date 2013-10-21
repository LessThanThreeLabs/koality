package rpc

import (
	"github.com/streadway/amqp"
	"sync"
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

func createExchanges() error {
	exchangeChannel, err := sendConnection.Channel()
	if err != nil {
		return err
	}

	err = exchangeChannel.ExchangeDeclare(exchangeName, exchangeType,
		exchangeDurable, exchangeAutoDelete, exchangeInternal, exchangeNoWait, nil)
	if err != nil {
		return err
	}

	err = exchangeChannel.ExchangeDeclare(deadLetterExchangeName, deadLetterExchangeType,
		deadLetterExchangeDurable, deadLetterExchangeAutoDelete,
		deadLetterExchangeInternal, deadLetterExchangeNoWait, nil)
	if err != nil {
		return err
	}

	return nil
}

func initializeAmqp() {
	err := createConnections()
	if err != nil {
		panic(err)
	}

	err = createExchanges()
	if err != nil {
		panic(err)
	}
}

func getSendConnection() *amqp.Connection {
	initializationSync.Do(initializeAmqp)
	return sendConnection
}

func getReceiveConnection() *amqp.Connection {
	initializationSync.Do(initializeAmqp)
	return receiveConnection
}
