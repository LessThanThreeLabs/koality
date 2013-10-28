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

	numRequests := 10
	completedRequest := make(chan *rpc.Response, numRequests)

	for index := 0; index < numRequests; index++ {
		go func() {
			request := rpc.NewRequest("getTime")
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
