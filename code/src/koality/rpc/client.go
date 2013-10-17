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
	deadLetterQueue   *amqp.Queue
	correlationIdLock *sync.Mutex
	nextCorrelationId uint64
	responseChannels  map[string]chan<- *Response
	msgpackHandle     *codec.MsgpackHandle
}

func NewClient(route string) *Client {
	sendChannel, err := getSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	returns := make(chan amqp.Return)
	sendChannel.NotifyReturn(returns)

	receiveChannel, err := getReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(responseQueueName, responseQueueDurable,
		responseQueueAutoDelete, responseQueueExclusive, responseQueueNoWait, nil)
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

	deliveries, err := receiveChannel.Consume(responseQueue.Name, responseQueue.Name, responseQueueAutoAck,
		responseQueueExclusive, responseQueueNoLocal, responseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	deadLetterDeliveries, err := receiveChannel.Consume(deadLetterQueue.Name, deadLetterQueue.Name, deadLetterQueueAutoAck,
		deadLetterQueueExclusive, deadLetterQueueNoLocal, deadLetterQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	client := Client{
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

	go client.handleReturns(returns)
	go client.handleDeliveries(deliveries)
	go client.handleDeadLetterDeliveries(deadLetterDeliveries)

	return &client
}

func (client *Client) getNextCorrelationId() string {
	client.correlationIdLock.Lock()
	correlationId := strconv.FormatUint(client.nextCorrelationId, 36)
	client.nextCorrelationId++
	client.correlationIdLock.Unlock()
	return correlationId
}

func (client *Client) SendRequest(rpcRequest *Request) (<-chan *Response, error) {
	var buffer []byte
	err := codec.NewEncoderBytes(&buffer, client.msgpackHandle).Encode(rpcRequest)
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

func (client *Client) handleDeliveries(deliveries <-chan amqp.Delivery) {
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

func (client *Client) handleDeadLetterDeliveries(deadLetterDeliveries <-chan amqp.Delivery) {
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

func (client *Client) handleReturns(returns <-chan amqp.Return) {
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
