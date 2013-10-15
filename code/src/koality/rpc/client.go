package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
	"strconv"
	"time"
)

const (
	amqpUri = "amqp://localhost:5672/"
)

const (
	exchangeName       = "rpc"
	exchangeType       = "direct"
	exchangeDurable    = false // TODO: make durable?
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false
	exchangeMandatory  = true
	exchangeImmediate  = false
)

const (
	responseQueueName       = ""
	responseQueueDurable    = false
	responseQueueAutoDelete = true
	responseQueueExclusive  = true
	responseQueueNoWait     = false
	responseQueueAutoAck    = true
	responseQueueNoLocal    = true
)

var (
	sendConnection    *amqp.Connection
	receiveConnection *amqp.Connection
)

func init() {
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

type Client struct {
	route             string
	sendChannel       *amqp.Channel
	receiveChannel    *amqp.Channel
	responseQueue     *amqp.Queue
	nextCorrelationId uint64
	responseChannels  map[uint64]chan<- *Response
	msgpackHandle     *codec.MsgpackHandle
}

func NewClient(route string) *Client {
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

	deliveries, err := receiveChannel.Consume(responseQueue.Name, responseQueue.Name, responseQueueAutoAck,
		responseQueueExclusive, responseQueueNoLocal, responseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	msgpackHandle := new(codec.MsgpackHandle)

	responseChannels := make(map[uint64]chan<- *Response)

	client := Client{route, sendChannel, receiveChannel, &responseQueue, 0, responseChannels, msgpackHandle}
	go client.handleDeliveries(deliveries)
	return &client
}

func (client *Client) getNextCorrelationId() uint64 {
	correlationId := client.nextCorrelationId
	client.nextCorrelationId++
	return correlationId
}

func (client *Client) SendRequest(rpcRequest *Request) <-chan *Response {
	var buffer []byte
	codec.NewEncoderBytes(&buffer, client.msgpackHandle).Encode(rpcRequest)

	correlationId := client.getNextCorrelationId()
	responseChannel := make(chan *Response)
	client.responseChannels[correlationId] = responseChannel

	message := amqp.Publishing{
		Body:          buffer,
		ContentType:   "application/x-msgpack",
		DeliveryMode:  amqp.Transient, // TODO: this the right choice?
		CorrelationId: strconv.FormatUint(correlationId, 36),
		ReplyTo:       client.responseQueue.Name,
		Timestamp:     time.Now(),
	}

	err := client.sendChannel.Publish(exchangeName, client.route, exchangeMandatory, exchangeImmediate, message)
	if err != nil {
		// should gracefully handle these, perhaps with a general
		// defer that could catch any error and return "unable to send"
		panic(err)
	}

	return responseChannel
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
