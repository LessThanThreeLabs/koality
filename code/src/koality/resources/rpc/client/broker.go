package client

import (
	"github.com/streadway/amqp"
	"koality/resources/rpc"
)

const (
	exchangeName       = "rpc"
	exchangeType       = "direct"
	exchangeDurable    = true
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false
)

const (
	responseQueueName       = ""
	responseQueueDurable    = true
	responseQueueAutoDelete = true
	responseQueueExclusive  = false
	responseQueueNoWait     = false
	responseQueueAutoAck    = true
	responseQueueNoLocal    = true
)

type RpcBroker struct {
	resourceName     string
	resourceAction   string
	sendChannel      *amqp.Channel
	receiveChannel   *amqp.Channel
	responseQueue    *amqp.Queue
	correlationId    uint64
	responseChannels map[uint64]chan rpc.RpcResponse
}

func NewBroker(sendConnection, receiveConnection *amqp.Connection, resourceName, resourceAction string) *RpcBroker {
	sendChannel, err := sendConnection.Channel()
	if err != nil {
		panic(err)
	}

	err = sendChannel.ExchangeDeclare(exchangeName, exchangeType, exchangeDurable,
		exchangeAutoDelete, exchangeInternal, exchangeNoWait, nil)
	if err != nil {
		panic(err)
	}

	// need to add dead letter exchange

	receiveChannel, err := receiveConnection.Channel()
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(responseQueueName, responseQueueDurable,
		responseQueueAutoDelete, responseQueueExclusive, responseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	responses, err := receiveChannel.Consume(responseQueue.Name, responseQueue.Name, responseQueueAutoAck,
		responseQueueExclusive, responseQueueNoLocal, responseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	go handleResponses(responses)

	responseChannels := make(map[uint64]chan rpc.RpcResponse)
	return &RpcBroker{resourceName, resourceAction, sendChannel, receiveChannel,
		&responseQueue, 0, responseChannels}
}

func (rpcConnection *RpcBroker) CallFunction(rpcRequest *rpc.RpcRequest) {

}

func handleResponses(responses <-chan amqp.Delivery) {
	for response := range responses {
		// do stuff with the response
		var _ = response
	}
}
