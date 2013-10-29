package rpc

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/ugorji/go/codec"
	"reflect"
	"unicode"
	"unicode/utf8"
)

type server struct {
	route          string
	requestHandler *reflect.Value
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

	reflectedRequestHandler := reflect.ValueOf(requestHandler)
	if !reflectedRequestHandler.IsValid() {
		panic("RPC Server: Unable to reflect on request handler")
	} else if reflectedRequestHandler.Elem().Kind() != reflect.Struct {
		panic("Must pass a struct for the request handler")
	}

	server := server{
		route:          route,
		requestHandler: &reflectedRequestHandler,
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

func (server *server) getMethod(methodName string) (*reflect.Value, error) {
	firstRuneOfMethod, _ := utf8.DecodeRune([]byte(methodName))
	if firstRuneOfMethod == utf8.RuneError {
		return nil, InvalidRequestError{"404 (method does not exist)"}
	} else if !unicode.IsUpper(firstRuneOfMethod) {
		return nil, InvalidRequestError{"404 (method does not exist)"}
	}

	method := server.requestHandler.MethodByName(methodName)
	if !method.IsValid() {
		return nil, InvalidRequestError{"404 (method does not exist)"}
	}

	return &method, nil
}

func (server *server) getMethodArgs(method *reflect.Value, args []interface{}) ([]reflect.Value, error) {
	if len(args) != method.Type().NumIn() {
		return nil, InvalidRequestError{"400 (mismatched number of arguments)"}
	}

	argValues := make([]reflect.Value, len(args))
	for index, arg := range args {
		argValues[index] = reflect.ValueOf(arg)
	}

	for index, argValue := range argValues {
		if argValue.Kind() != method.Type().In(index).Kind() {
			return nil, InvalidRequestError{fmt.Sprintf("400 (mismatched parameter for index %d. Recevied %s but expected %s)",
				index, argValue.Kind(), method.Type().In(index).Kind())}
		}
	}

	return argValues, nil
}

func (server *server) handleRequest(rpcRequest *Request, replyToQueueName string) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("Handle unexpected error")
		}
	}()

	method, err := server.getMethod(rpcRequest.Method)
	if err != nil {
		fmt.Println(err)
	}

	methodArgs, err := server.getMethodArgs(method, rpcRequest.Args)
	if err != nil {
		fmt.Println(err)
	}

	returnValues := method.Call(methodArgs)
	for index, returnValue := range returnValues {
		if !returnValue.CanInterface() {
			fmt.Printf("Received bad return type for return value %d", index)
			return
		}
	}

	fmt.Printf("%d\n", returnValues[0].Interface())
}
