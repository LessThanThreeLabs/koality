package client

import (
	"github.com/streadway/amqp"
	"koality/resources"
	"time"
)

const (
	amqpUri = "amqp://localhost:5672/"
)

func NewConnection() *resources.Connection {
	sendConnection, err := amqp.Dial(amqpUri)
	if err != nil {
		panic(err)
	}

	receiveConnection, err := amqp.Dial(amqpUri)
	if err != nil {
		panic(err)
	}

	usersHandler := createUsersHandler(sendConnection, receiveConnection)
	return &resources.Connection{usersHandler}
}

func createUsersHandler(sendConnection, receiveConnection *amqp.Connection) resources.UsersHandler {
	usersReadRpcConnection := NewBroker(sendConnection, receiveConnection, "users", "read")
	usersReadHandler := UsersReadRpcHandler{usersReadRpcConnection}

	return resources.UsersHandler{usersReadHandler}
}

type UsersReadRpcHandler struct {
	rpcBroker *RpcBroker
}

func (readHandler UsersReadRpcHandler) Get(id int) (resources.User, error) {
	responseChannel := make(chan resources.User)

	go func() {
		fakeUser := resources.User{id, "JordanNPotter@gmail.com", "Jordan", "Potter"}
		time.Sleep(1000 * time.Millisecond)
		responseChannel <- fakeUser
	}()

	return <-responseChannel, nil
}
