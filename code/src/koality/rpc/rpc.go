package rpc

import (
	"fmt"
)

type Request struct {
	Method    string        `codec:"method"`
	Arguments []interface{} `codec:"args"`
}

func NewRequest(method string, arguments ...interface{}) *Request {
	return &Request{method, arguments}
}

type Response struct {
	Value interface{}   `codec:"value"`
	Error ResponseError `codec:"error"`
}

type ResponseError struct {
	Type      string `codec:"type"`
	Message   string `codec:"message"`
	Traceback string `codec:"traceback"`
}

func (err ResponseError) Error() string {
	if err.Traceback == "" {
		return fmt.Sprintf("%s: %s", err.Type, err.Message)
	} else {
		return fmt.Sprintf("%s: %s\n%s", err.Type, err.Message, err.Traceback)
	}
}
