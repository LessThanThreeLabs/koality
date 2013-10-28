package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
)

type server struct {
	route          string
	requestHandler interface{}
	sendChannel    *amqp.Channel
	receiveChannel *amqp.Channel
	responseQueue  *amqp.Queue
	msgpackHandle  *codec.MsgpackHandle
}

func NewServer(route string, requestHandler interface{}) *server {
	sendChannel, err := getSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	receiveChannel, err := getReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	err = receiveChannel.Qos(serverResponseQueueQos, 0, false)
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(serverResponseQueueName, serverResponseQueueDurable,
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
		requestHandler: requestHandler,
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

			rpcRequest := new(Request)
			err := codec.NewDecoderBytes(delivery.Body, server.msgpackHandle).Decode(rpcRequest)
			if err != nil {
				panic(err)
			}

			server.handleRequest(rpcRequest, delivery.ReplyTo)
			delivery.Ack(false)
		}(delivery)
	}
}

func (server *server) handleRequest(rpcRequest *Request, replyToQueueName string) {
	fmt.Println(rpcRequest, replyToQueueName)
	fmt.Println("Need to use reflection to call correct function of handler here")
}
