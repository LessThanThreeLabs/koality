package rpc

import (
	"fmt"
)

type Request struct {
	Method string        `codec:"method"`
	Args   []interface{} `codec:"args"`
}

func NewRequest(method string, args ...interface{}) *Request {
	return &Request{method, args}
}

type RequestError struct {
	Message string
}

func (err RequestError) Error() string {
	return err.Message
}

type Response struct {
	Values []interface{} `codec:"values"`
	Error  error         `codec:"error"`
}

type ResponseError struct {
	Type    string `codec:"type"`
	Message string `codec:"message"`
}

func (err ResponseError) Error() string {
	if err.Message == "" {
		return err.Type
	} else {
		return fmt.Sprintf("%s: %s", err.Type, err.Message)
	}
}
