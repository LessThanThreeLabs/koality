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
	handler := &RequestHandler{}
	rpcServer := rpc.NewServer("info", handler)
	var _ = rpcServer

	time.Sleep(time.Minute)
}

type RequestHandler struct {
}