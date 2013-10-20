package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
)

type server struct {
	route          string
	sendChannel    *amqp.Channel
	receiveChannel *amqp.Channel
	responseQueue  *amqp.Queue
	msgpackHandle  *codec.MsgpackHandle
}

func NewServer(route string) *server {
	sendChannel, err := getSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	receiveChannel, err := getReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(route, serverResponseQueueDurable,
		serverResponseQueueAutoDelete, serverResponseQueueExclusive, serverResponseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	err = receiveChannel.QueueBind(responseQueue.Name, route, exchangeName, serverResponseQueueBindNoWait, nil)
	if err != nil {
		panic(err)
	}

	server := server{
		route:          route,
		sendChannel:    sendChannel,
		receiveChannel: receiveChannel,
		responseQueue:  &responseQueue,
		msgpackHandle:  new(codec.MsgpackHandle),
	}

	go server.handleDeliveries()

	return &server
}

func (server *server) handleDeliveries() {
	deliveries, err := server.receiveChannel.Consume(server.responseQueue.Name, server.responseQueue.Name,
		serverResponseQueueAutoAck, serverResponseQueueExclusive,
		serverResponseQueueNoLocal, serverResponseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	for delivery := range deliveries {
		go func(delivery amqp.Delivery) {
			if delivery.ContentType != "application/x-msgpack" {
				panic(fmt.Sprintf("Unsupported content type: %s", delivery.ContentType))
			}

			response := new(Response)
			err := codec.NewDecoderBytes(delivery.Body, server.msgpackHandle).Decode(response)
			if err != nil {
				panic(err)
			}

			// TODO: handle request
		}(delivery)
	}
}
