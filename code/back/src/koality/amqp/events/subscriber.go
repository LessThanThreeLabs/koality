package events

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack"
	kamqp "koality/amqp"
	"reflect"
)

type Subscriber struct {
	route          string
	eventHandler   *reflect.Value
	receiveChannel *amqp.Channel
	receiveQueue   *amqp.Queue
}

func NewSubscriber(route string, eventHandler interface{}) *Subscriber {
	createExchanges()

	receiveChannel, err := kamqp.GetReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	err = receiveChannel.Qos(receiveQueueQos, 0, false)
	if err != nil {
		panic(err)
	}

	receiveQueue, err := receiveChannel.QueueDeclare(receiveQueueName, receiveQueueDurable,
		receiveQueueAutoDelete, receiveQueueExclusive, receiveQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	err = receiveChannel.QueueBind(receiveQueue.Name, route, exchangeName, receiveQueueBindNoWait, nil)
	if err != nil {
		panic(err)
	}

	reflectedEventHandler := reflect.ValueOf(eventHandler)
	if !reflectedEventHandler.IsValid() {
		panic("Event Subscriber: Unable to reflect on event handler")
	}
	fmt.Println("make sure it has the handler() function we need")

	subscriber := Subscriber{
		route:          route,
		eventHandler:   &reflectedEventHandler,
		receiveChannel: receiveChannel,
		receiveQueue:   &receiveQueue,
	}

	go subscriber.handleDeliveries()

	return &subscriber
}

func (subscriber *Subscriber) handleDeliveries() {
	deliveries, err := subscriber.receiveChannel.Consume(subscriber.receiveQueue.Name, subscriber.receiveQueue.Name,
		receiveQueueAutoAck, receiveQueueExclusive, receiveQueueNoLocal, receiveQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	for delivery := range deliveries {
		go func(delivery amqp.Delivery) {
			if delivery.ContentType != "application/x-msgpack" {
				panic(fmt.Sprintf("Unsupported content type: %s", delivery.ContentType))
			}

			event := new(Event)
			err := msgpack.Unmarshal(delivery.Body, &event)
			if err != nil {
				panic(err)
			}

			fmt.Println(event)
			delivery.Ack(false)
		}(delivery)
	}
}
