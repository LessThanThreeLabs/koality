package amqp

import (
	"koality/resources"
	"time"
)

func NewConnection() resources.Connection {
	usersReadHandler := RpcUsersReadHandler{}
	usersHandler := resources.UsersHandler{&usersReadHandler}

	return resources.Connection{usersHandler}
}

type RpcUsersReadHandler struct {
}

func (readHandler *RpcUsersReadHandler) Get(id int) (resources.User, error) {
	responseChannel := make(chan resources.User)

	go func() {
		fakeUser := resources.User{18, "JordanNPotter2@gmail.com", "Jordan2", "Potter2"}
		time.Sleep(1000 * time.Millisecond)
		responseChannel <- fakeUser
	}()

	return <-responseChannel, nil
}
