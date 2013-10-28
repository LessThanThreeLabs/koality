package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

type client struct {
	route             string
	sendChannel       *amqp.Channel
	receiveChannel    *amqp.Channel
	responseQueue     *amqp.Queue
	deadLetterQueue   *amqp.Queue
	correlationIdLock *sync.Mutex
	nextCorrelationId uint64
	responseChannels  map[string]chan<- *Response
	msgpackHandle     *codec.MsgpackHandle
}

func NewClient(route string) *client {
	sendChannel, err := getSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	receiveChannel, err := getReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(clientResponseQueueName, clientResponseQueueDurable,
		clientResponseQueueAutoDelete, clientResponseQueueExclusive, clientResponseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	deadLetterQueue, err := receiveChannel.QueueDeclare(deadLetterQueueName, deadLetterQueueDurable,
		deadLetterQueueAutoDelete, deadLetterQueueExclusive, deadLetterQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	err = receiveChannel.QueueBind(deadLetterQueue.Name, "", deadLetterExchangeName, deadLetterQueueBindNoWait, nil)
	if err != nil {
		panic(err)
	}

	client := client{
		route:             route,
		sendChannel:       sendChannel,
		receiveChannel:    receiveChannel,
		responseQueue:     &responseQueue,
		deadLetterQueue:   &deadLetterQueue,
		nextCorrelationId: 0,
		correlationIdLock: new(sync.Mutex),
		responseChannels:  make(map[string]chan<- *Response),
		msgpackHandle:     new(codec.MsgpackHandle),
	}

	go client.handleDeliveries()
	go client.handleDeadLetterDeliveries()
	go client.handleReturns()

	return &client
}

func (client *client) getNextCorrelationId() string {
	client.correlationIdLock.Lock()
	defer client.correlationIdLock.Unlock()

	correlationId := strconv.FormatUint(client.nextCorrelationId, 36)
	client.nextCorrelationId++
	return correlationId
}

func (client *client) checkRequestIsValid(rpcRequest *Request) error {
	for arg := range rpcRequest.Args {
		fmt.Println("need to check unicode better?...")
		if !utf8.ValidString(fmt.Sprint(arg)) {
			return &InvalidRequestError{Message: "Request argument contains illegal character"}
		}
	}
	return nil
}

func (client *client) SendRequest(rpcRequest *Request) (<-chan *Response, error) {
	err := client.checkRequestIsValid(rpcRequest)
	if err != nil {
		return nil, err
	}

	var buffer []byte
	err = codec.NewEncoderBytes(&buffer, client.msgpackHandle).Encode(rpcRequest)
	if err != nil {
		return nil, err
	}

	correlationId := client.getNextCorrelationId()

	publishing := amqp.Publishing{
		Body:            buffer,
		ContentType:     "application/x-msgpack",
		ContentEncoding: "binary",
		DeliveryMode:    amqp.Transient,
		CorrelationId:   correlationId,
		ReplyTo:         client.responseQueue.Name,
		Timestamp:       time.Now(),
	}

	err = client.sendChannel.Publish(exchangeName, client.route, exchangeMandatory, exchangeImmediate, publishing)
	if err != nil {
		return nil, err
	}

	responseChannel := make(chan *Response)
	client.responseChannels[correlationId] = responseChannel
	return responseChannel, nil
}

func (client *client) handleDeliveries() {
	deliveries, err := client.receiveChannel.Consume(client.responseQueue.Name, client.responseQueue.Name,
		clientResponseQueueAutoAck, clientResponseQueueExclusive,
		clientResponseQueueNoLocal, clientResponseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	for delivery := range deliveries {
		go func(delivery amqp.Delivery) {
			if delivery.ContentType != "application/x-msgpack" {
				panic(fmt.Sprintf("Unsupported content type: %s", delivery.ContentType))
			}

			response := new(Response)
			err := codec.NewDecoderBytes(delivery.Body, client.msgpackHandle).Decode(response)
			if err != nil {
				panic(err)
			}

			correlationId := delivery.CorrelationId
			responseChannel, ok := client.responseChannels[correlationId]
			if !ok {
				panic(fmt.Sprintf("Unexpected correlation id: %s", correlationId))
			}

			responseChannel <- response
			close(responseChannel)
			delete(client.responseChannels, correlationId)
		}(delivery)
	}
}

func (client *client) handleDeadLetterDeliveries() {
	deadLetterDeliveries, err := client.receiveChannel.Consume(client.deadLetterQueue.Name, client.deadLetterQueue.Name,
		deadLetterQueueAutoAck, deadLetterQueueExclusive,
		deadLetterQueueNoLocal, deadLetterQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	for deadLetterDelivery := range deadLetterDeliveries {

		go func(deadLetterDelivery amqp.Delivery) {
			correlationId := deadLetterDelivery.CorrelationId
			responseChannel, ok := client.responseChannels[correlationId]
			if !ok {
				return
			}

			response := Response{
				Value: nil,
				Error: ResponseError{
					Type:      "TimedOutError",
					Message:   "Request ttl expired",
					Traceback: "",
				},
			}

			responseChannel <- &response
			close(responseChannel)
			delete(client.responseChannels, correlationId)
		}(deadLetterDelivery)
	}
}

func (client *client) handleReturns() {
	returns := make(chan amqp.Return)
	client.sendChannel.NotifyReturn(returns)

	for badReturn := range returns {

		go func(badReturn amqp.Return) {
			correlationId := badReturn.CorrelationId
			responseChannel, ok := client.responseChannels[correlationId]
			if !ok {
				panic(fmt.Sprintf("Unexpected correlation id: %d", correlationId))
			}

			response := Response{
				Value: nil,
				Error: ResponseError{
					Type:      "BadRequestError",
					Message:   fmt.Sprintf("Bad request - %s", badReturn.ReplyText),
					Traceback: "",
				},
			}

			responseChannel <- &response
			close(responseChannel)
			delete(client.responseChannels, correlationId)
		}(badReturn)
	}
}
