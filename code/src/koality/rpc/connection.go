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

func creatConnections() {
	var err error

	sendConnection, err = amqp.Dial(amqpUri)
	if err != nil {
		panic(err)
	}

	receiveConnection, err = amqp.Dial(amqpUri)
	if err != nil {
		panic(err)
	}
}

func createExchanges() {
	exchangeChannel, err := sendConnection.Channel()
	if err != nil {
		panic(err)
	}

	exchangeOptions := amqp.Table(map[string]interface{}{
		"x-dead-letter-exchange": deadLetterExchangeName,
	})
	err = exchangeChannel.ExchangeDeclare(exchangeName, exchangeType,
		exchangeDurable, exchangeAutoDelete, exchangeInternal, exchangeNoWait, exchangeOptions)
	if err != nil {
		panic(err)
	}

	err = exchangeChannel.ExchangeDeclare(deadLetterExchangeName, deadLetterExchangeType,
		deadLetterExchangeDurable, deadLetterExchangeAutoDelete,
		deadLetterExchangeInternal, deadLetterExchangeNoWait, nil)
	if err != nil {
		panic(err)
	}
}

func initializeAmqp() {
	creatConnections()
	createExchanges()
}

func getSendConnection() *amqp.Connection {
	initializationSync.Do(initializeAmqp)
	return sendConnection
}

func getReceiveConnection() *amqp.Connection {
	initializationSync.Do(initializeAmqp)
	return receiveConnection
}
