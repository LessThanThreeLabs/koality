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
	usersReadRpcConnection := rpc.NewClient("rpc:users.read")
	usersReadHandler := UsersReadRpcHandler{usersReadRpcConnection}

	return resources.UsersHandler{usersReadHandler}
}

type UsersReadRpcHandler struct {
	rpcClient *rpc.Client
}

func (readHandler UsersReadRpcHandler) Get(id int) (*resources.User, error) {
	request := rpc.NewRequest("get_user_from_id", id)
	responseChan, err := readHandler.rpcClient.SendRequest(request)
	if err != nil {
		return nil, err
	}

	response := <-responseChan
	fmt.Printf("\n%v\n\n", response)

	return &resources.User{id, "JordanNPotter@gmail.com", "Jordan", "Potter"}, nil
}
