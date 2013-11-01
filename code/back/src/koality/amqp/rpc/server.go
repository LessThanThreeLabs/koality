package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vmihailenco/msgpack"
	kamqp "koality/amqp"
	"reflect"
	"time"
	"unicode"
	"unicode/utf8"
)

type server struct {
	route          string
	requestHandler *reflect.Value
	sendChannel    *amqp.Channel
	receiveChannel *amqp.Channel
	responseQueue  *amqp.Queue
}

func NewServer(route string, requestHandler interface{}) *server {
	sendChannel, err := kamqp.GetSendConnection().Channel()
	if err != nil {
		panic(err)
	}

	receiveChannel, err := kamqp.GetReceiveConnection().Channel()
	if err != nil {
		panic(err)
	}

	err = receiveChannel.Qos(serverResponseQueueQos, 0, false)
	if err != nil {
		panic(err)
	}

	err = sendChannel.ExchangeDeclare(exchangeName, exchangeType,
		exchangeDurable, exchangeAutoDelete, exchangeInternal, exchangeNoWait, nil)
	if err != nil {
		panic(err)
	}

	err = sendChannel.ExchangeDeclare(deadLetterExchangeName, deadLetterExchangeType,
		deadLetterExchangeDurable, deadLetterExchangeAutoDelete,
		deadLetterExchangeInternal, deadLetterExchangeNoWait, nil)
	if err != nil {
		panic(err)
	}

	responseQueue, err := receiveChannel.QueueDeclare(serverResponseQueueNamePrefix+route, serverResponseQueueDurable,
		serverResponseQueueAutoDelete, serverResponseQueueExclusive, serverResponseQueueNoWait, nil)
	if err != nil {
		panic(err)
	}

	err = receiveChannel.QueueBind(responseQueue.Name, route, exchangeName, serverResponseQueueBindNoWait, nil)
	if err != nil {
		panic(err)
	}

	reflectedRequestHandler := reflect.ValueOf(requestHandler)
	if !reflectedRequestHandler.IsValid() {
		panic("RPC Server: Unable to reflect on request handler")
	}

	server := server{
		route:          route,
		requestHandler: &reflectedRequestHandler,
		sendChannel:    sendChannel,
		receiveChannel: receiveChannel,
		responseQueue:  &responseQueue,
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
			err := msgpack.Unmarshal(delivery.Body, &rpcRequest)
			if err != nil {
				panic(err)
			}

			server.handleRequest(rpcRequest, delivery.ReplyTo, delivery.CorrelationId)
			delivery.Ack(false)
		}(delivery)
	}
}

func (server *server) getMethod(methodName string) (*reflect.Value, error) {
	firstRuneOfMethod, _ := utf8.DecodeRune([]byte(methodName))
	if firstRuneOfMethod == utf8.RuneError {
		return nil, ResponseError{Type: "MethodDoesNotExist", Message: "Method does not exist"}
	} else if !unicode.IsUpper(firstRuneOfMethod) {
		return nil, ResponseError{Type: "MethodDoesNotExist", Message: "Method does not exist"}
	}

	method := server.requestHandler.MethodByName(methodName)
	if !method.IsValid() {
		return nil, ResponseError{Type: "MethodDoesNotExist", Message: "Method does not exist"}
	}

	return &method, nil
}

func (server *server) getMethodArgs(method *reflect.Value, args []interface{}) ([]reflect.Value, error) {
	if len(args) != method.Type().NumIn() {
		return nil, ResponseError{Type: "MethodCallFailed", Message: "Mismatched number of arguments"}
	}

	argValues := make([]reflect.Value, len(args))
	for index, arg := range args {
		argValues[index] = reflect.ValueOf(arg)
		if !argValues[index].IsValid() {
			return nil, ResponseError{Type: "MethodCallFailed", Message: "Failed to reflect on argument"}
		}
	}

	for index, argValue := range argValues {
		if argValue.Kind() != method.Type().In(index).Kind() {
			errorMessage := fmt.Sprintf("Mismatched parameter for index %d. Recevied %s but expected %s)",
				index, argValue.Kind(), method.Type().In(index).Kind())
			return nil, ResponseError{Type: "MethodCallFailed", Message: errorMessage}
		}
	}

	return argValues, nil
}

func (server *server) getReturnValues(method *reflect.Value, methodArgs []reflect.Value) ([]interface{}, error) {
	returnValues := method.Call(methodArgs)

	returnInterfaces := make([]interface{}, len(returnValues))
	for index, returnValue := range returnValues {
		if !returnValue.CanInterface() {
			return nil, ResponseError{Type: "MethodCallFailed", Message: "Received bad value for return type"}
		}
		returnInterfaces[index] = returnValue.Interface()
	}

	return returnInterfaces, nil
}

func (server *server) handleRequest(rpcRequest *Request, replyToQueueName, correlationId string) {
	method, err := server.getMethod(rpcRequest.Method)
	if err != nil {
		server.sendResponse(nil, err, replyToQueueName, correlationId)
		return
	}

	methodArgs, err := server.getMethodArgs(method, rpcRequest.Args)
	if err != nil {
		server.sendResponse(nil, err, replyToQueueName, correlationId)
		return
	}

	returnValues, err := server.getReturnValues(method, methodArgs)
	if err != nil {
		server.sendResponse(nil, err, replyToQueueName, correlationId)
		return
	}

	server.sendResponse(returnValues, err, replyToQueueName, correlationId)
}

func (server *server) sendResponse(values []interface{}, err error, replyToQueueName, correlationId string) {
	response := Response{values, err}

	buffer, err := msgpack.Marshal(response)
	if err != nil {
		panic(err)
	}

	publishing := amqp.Publishing{
		Body:            buffer,
		ContentType:     "application/x-msgpack",
		ContentEncoding: "binary",
		DeliveryMode:    amqp.Transient,
		CorrelationId:   correlationId,
		Timestamp:       time.Now(),
	}

	err = server.sendChannel.Publish("", replyToQueueName, exchangeMandatory, exchangeImmediate, publishing)
	if err != nil {
		panic(err)
	}
}
