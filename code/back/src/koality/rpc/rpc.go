package rpc

import (
	"fmt"
	"reflect"
)

type Request struct {
	Method string        `codec:"method"`
	Args   []interface{} `codec:"args"`
}

func NewRequest(method string, args ...interface{}) *Request {
	for _, arg := range args {
		fmt.Println(reflect.TypeOf(arg))
	}
	return &Request{method, args}
}

type InvalidRequestError struct {
	Message string
}

func (err InvalidRequestError) Error() string {
	return err.Message
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
	if err.Message == "" {
		return err.Type
	} else if err.Traceback == "" {
		return fmt.Sprintf("%s: %s", err.Type, err.Message)
	} else {
		return fmt.Sprintf("%s: %s\n%s", err.Type, err.Message, err.Traceback)
	}
}
