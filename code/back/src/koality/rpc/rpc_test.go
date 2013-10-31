package rpc

import (
	"strings"
	"testing"
)

const (
	route = "test-route"
)

var (
	shortString = makeTestString("short", 10)    // =60B
	longString  = makeTestString("long", 200000) // =1MB
)

func makeTestString(text string, numRepeats int) string {
	stringsArray := make([]string, numRepeats)
	for index := 0; index < numRepeats; index++ {
		stringsArray[index] = text
	}
	return strings.Join(stringsArray, " ")
}

func TestSmallRequests(testing *testing.T) {
	client := NewClient(route)
	server := NewServer(route, new(RequestHandler))
	shortRequest := NewRequest("GetShortString", shortString, shortString, shortString)
	sendRequests(client, server, shortRequest, 10000, 1000)
}

func TestLargeRequests(testing *testing.T) {
	client := NewClient(route)
	server := NewServer(route, new(RequestHandler))
	longRequest := NewRequest("GetLongString", longString, longString, longString)
	sendRequests(client, server, longRequest, 1000, 100)
}

func TestMixedRequests(testing *testing.T) {
	client := NewClient(route)
	server := NewServer(route, new(RequestHandler))
	requestsCompleted := make(chan bool)

	shortRequest := NewRequest("GetShortString", shortString, shortString, shortString)
	go func() {
		sendRequests(client, server, shortRequest, 10000, 1000)
		requestsCompleted <- true
	}()

	longRequest := NewRequest("GetLongString", longString, longString, longString)
	go func() {
		sendRequests(client, server, longRequest, 1000, 100)
		requestsCompleted <- true
	}()

	<-requestsCompleted
	<-requestsCompleted
}

func sendRequests(rpcClient *client, rpcServer *server, rpcRequest *Request, numRequests, maxConcurrentRequests int) {
	semaphore := make(chan bool, maxConcurrentRequests)
	completed := make(chan *Response, maxConcurrentRequests)

	go func() {
		for index := 0; index < numRequests; index++ {
			semaphore <- true
			go func() {
				responseChan, err := rpcClient.SendRequest(rpcRequest)
				if err != nil {
					panic(err)
				}

				completed <- <-responseChan
				<-semaphore
			}()
		}
	}()

	for index := 0; index < numRequests; index++ {
		rpcResponse := <-completed
		if rpcResponse.Error != nil {
			panic(rpcResponse.Error)
		}
	}
}

type RequestHandler int

func (requestHandler *RequestHandler) GetShortString(first, second, third string) (string, error) {
	return shortString, nil
}

func (requestHandler *RequestHandler) GetLongString(first, second, third string) (string, error) {
	return longString, nil
}
