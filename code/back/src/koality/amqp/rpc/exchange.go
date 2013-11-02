package rpc

import (
	"koality/amqp"
	"sync"
)

var (
	initializationSync sync.Once
)

func initializeExchanges() {
	exchangeChannel, err := amqp.GetSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	err = exchangeChannel.ExchangeDeclare(exchangeName, exchangeType,
		exchangeDurable, exchangeAutoDelete, exchangeInternal, exchangeNoWait, nil)
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

func createExchanges() {
	initializationSync.Do(initializeExchanges)
}
