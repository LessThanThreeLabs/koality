package events

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack"
	kamqp "koality/amqp"
	"sync"
)

type Subscriber struct {
	route              string
	receiveChannel     *amqp.Channel
	receiveQueue       *amqp.Queue
	eventChannels      []chan *Event
	eventChannelsMutex *sync.RWMutex
}

func NewSubscriber(route string) *Subscriber {
	createExchanges()

	receiveChannel, err := kamqp.GetReceiveConnection().Channel()
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

	subscriber := Subscriber{
		route:              route,
		receiveChannel:     receiveChannel,
		receiveQueue:       &receiveQueue,
		eventChannels:      make([]chan *Event, 0),
		eventChannelsMutex: new(sync.RWMutex),
	}

	go subscriber.handleDeliveries()

	return &subscriber
}

func (subscriber *Subscriber) Subscribe() chan *Event {
	subscriber.eventChannelsMutex.Lock()
	defer subscriber.eventChannelsMutex.Unlock()

	eventChannel := make(chan *Event, 100)
	subscriber.eventChannels = append(subscriber.eventChannels, eventChannel)

	return eventChannel
}

func (subscriber *Subscriber) Unsubscribe(eventChannel chan *Event) error {
	subscriber.eventChannelsMutex.Lock()
	defer subscriber.eventChannelsMutex.Unlock()

	for index, otherEventChannel := range subscriber.eventChannels {
		if eventChannel == otherEventChannel {
			subscriber.eventChannels = append(subscriber.eventChannels[:index], subscriber.eventChannels[index+1:]...)
			close(eventChannel)
			return nil
		}
	}

	return errors.New("Events Subscriber: No such event channel")
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

			subscriber.deliverEvent(event)
		}(delivery)
	}
}

func (subscriber *Subscriber) deliverEvent(event *Event) {
	subscriber.eventChannelsMutex.RLock()
	defer subscriber.eventChannelsMutex.RUnlock()

	for _, eventChannel := range subscriber.eventChannels {
		eventChannel <- event
	}
}
