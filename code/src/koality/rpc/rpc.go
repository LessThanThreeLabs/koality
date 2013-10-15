package rpc

type Request struct {
	Method    string        `codec:"method"`
	Arguments []interface{} `codec:"args"`
}

func MakeRequest(method string, arguments ...interface{}) *Request {
	return &Request{method, arguments}
}

type Response struct {
	Value interface{}   `codec:"value"`
	Error ResponseError `codec:"error"`
}

type ResponseError struct {
	Message   string `codec:"message"`
	Type      string `codec:"type"`
	Traceback string `codec:"traceback"`
}
