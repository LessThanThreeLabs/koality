package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	route             string
	sendChannel       *amqp.Channel
	receiveChannel    *amqp.Channel
	responseQueue     *amqp.Queue
	correlationIdLock *sync.Mutex
	nextCorrelationId uint64
	responseChannels  map[uint64]chan<- *Response
	msgpackHandle     *codec.MsgpackHandle
}

func NewClient(route string) *Client {
	sendChannel, err := getSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	receiveChannel, err := getReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(responseQueueName, responseQueueDurable,
		responseQueueAutoDelete, responseQueueExclusive, responseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	// TODO: add dead letter queue here

	deliveries, err := receiveChannel.Consume(responseQueue.Name, responseQueue.Name, responseQueueAutoAck,
		responseQueueExclusive, responseQueueNoLocal, responseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	client := Client{
		route:             route,
		sendChannel:       sendChannel,
		receiveChannel:    receiveChannel,
		responseQueue:     &responseQueue,
		nextCorrelationId: 0,
		correlationIdLock: new(sync.Mutex),
		responseChannels:  make(map[uint64]chan<- *Response),
		msgpackHandle:     new(codec.MsgpackHandle),
	}

	go client.handleDeliveries(deliveries)

	return &client
}

func (client *Client) getNextCorrelationId() uint64 {
	client.correlationIdLock.Lock()
	correlationId := client.nextCorrelationId
	client.nextCorrelationId++
	client.correlationIdLock.Unlock()
	return correlationId
}

func (client *Client) SendRequest(rpcRequest *Request) (<-chan *Response, error) {
	var buffer []byte
	codec.NewEncoderBytes(&buffer, client.msgpackHandle).Encode(rpcRequest)

	correlationId := client.getNextCorrelationId()

	message := amqp.Publishing{
		Body:            buffer,
		ContentType:     "application/x-msgpack",
		ContentEncoding: "binary",
		DeliveryMode:    amqp.Transient,
		CorrelationId:   strconv.FormatUint(correlationId, 36),
		ReplyTo:         client.responseQueue.Name,
		Timestamp:       time.Now(),
	}

	err := client.sendChannel.Publish(exchangeName, client.route, exchangeMandatory, exchangeImmediate, message)
	if err != nil {
		return nil, err
	}

	responseChannel := make(chan *Response)
	client.responseChannels[correlationId] = responseChannel
	return responseChannel, nil
}

func (client *Client) handleDeliveries(deliveries <-chan amqp.Delivery) {
	for delivery := range deliveries {

		go func(delivery amqp.Delivery) {
			if delivery.ContentType != "application/x-msgpack" {
				panic(fmt.Sprintf("Unsupported content type: %s", delivery.ContentType))
			}

			response := new(Response)
			codec.NewDecoderBytes(delivery.Body, client.msgpackHandle).Decode(response)

			correlationId, err := strconv.ParseUint(delivery.CorrelationId, 36, 64)
			if err != nil {
				panic(err)
			}

			responseChannel, ok := client.responseChannels[correlationId]
			if !ok {
				panic(fmt.Sprintf("Unexpected correlation id: %d", correlationId))
			}

			responseChannel <- response
			close(responseChannel)
			delete(client.responseChannels, correlationId)
		}(delivery)
	}
}
