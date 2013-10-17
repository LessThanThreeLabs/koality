package main

import (
	"fmt"
	// "koality/resources/rpc"
	"koality/rpc"
	"reflect"
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
		fmt.Printf("%v\n", <-completedRequest)
	}
}

// func testModelServer() {
// 	connection := rpc.NewConnection()

// 	fmt.Printf("Getting user 17...\n")
// 	user, err := connection.Users.Read.Get(1000)
// 	if err != nil {
// 		fmt.Println("Error: %v", err)
// 	} else {
// 		fmt.Println(user)
// 	}
// }

func testReflection() {
	c := map[string]int{"there": 14}
	arr := []interface{}{7, "hello", c}
	fmt.Println(arr)

	blah := Blah(7)

	method, _ := reflect.TypeOf(blah).MethodByName("hello")
	fmt.Println(method.Func.Type().In(1))

	arr2 := []reflect.Value{
		reflect.ValueOf(arr[0]),
		reflect.ValueOf(arr[1]),
	}
	method.Func.Call(arr2)
}

type Blah int

func (blah Blah) hello(a int, b string) {
	fmt.Println("hello there!")
}
