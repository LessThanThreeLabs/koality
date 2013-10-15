package rpc

import (
	"fmt"
	"koality/resources"
	"koality/rpc"
)

func NewConnection() *resources.Connection {
	usersHandler := createUsersHandler()
	return &resources.Connection{usersHandler}
}

func createUsersHandler() resources.UsersHandler {
	// usersReadRpcConnection := rpc.NewClient("users.read")
	usersReadRpcConnection := rpc.NewClient("info")
	usersReadHandler := UsersReadRpcHandler{usersReadRpcConnection}

	return resources.UsersHandler{usersReadHandler}
}

type UsersReadRpcHandler struct {
	rpcClient *rpc.Client
}

func (readHandler UsersReadRpcHandler) Get(id int) (resources.User, error) {
	request := rpc.MakeRequest("getTime")
	responseChan, err := readHandler.rpcClient.SendRequest(request)
	if err != nil {
		return resources.User{}, err
	}

	response := <-responseChan
	fmt.Printf("%f\n", response.Value)

	return resources.User{id, "JordanNPotter@gmail.com", "Jordan", "Potter"}, nil
}
