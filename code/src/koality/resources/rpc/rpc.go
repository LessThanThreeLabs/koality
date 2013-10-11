package rpc

type RpcRequest struct {
	Method    string      `codec:"method"`
	Arguments interface{} `codec:"args"`
}

type RpcResponse struct {
	Value interface{}      `codec:"value"`
	Error RpcResponseError `codec:"error"`
}

type RpcResponseError struct {
	Message   string `codec:"message"`
	Type      string `codec:"type"`
	Traceback string `codec:"traceback"`
}
