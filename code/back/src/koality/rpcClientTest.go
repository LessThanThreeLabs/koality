package main

import (
	"fmt"
	"koality/rpc"
	"time"
)

func main() {
	fmt.Println(">> ", time.Now())
	testExample()
	fmt.Println(">> ", time.Now())
}

func testExample() {
	rpcClient := rpc.NewClient("info")

	numRequests := 1
	completedRequest := make(chan *rpc.Response, numRequests)

	for index := 0; index < numRequests; index++ {
		go func() {
			var first int64 = 1
			var second int64 = 2

			request := rpc.NewRequest("Add", first, second)
			responseChan, err := rpcClient.SendRequest(request)
			if err != nil {
				panic(err)
			}

			completedRequest <- <-responseChan
		}()
	}

	for index := 0; index < numRequests; index++ {
		fmt.Println(<-completedRequest)
	}
}
