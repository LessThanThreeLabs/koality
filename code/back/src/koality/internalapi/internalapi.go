package internalapi

import (
	"koality/resources"
	"net"
	"net/rpc"
)

const (
	RpcSocket     = "/tmp/koality-rpc.sock"
)

func Setup(resourcesConnection *resources.Connection, rpcSocket string) error {
	server := rpc.NewServer()
	server.Register(&PublicKeyVerifier{resourcesConnection})
	listener, err := net.Listen("unix", rpcSocket)
	if err != nil {
		return err
	}

	go server.Accept(listener)
	return nil
}