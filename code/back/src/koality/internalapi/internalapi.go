package internalapi

import (
	"koality/resources"
	"koality/vm/poolmanager"
	"net"
	"net/rpc"
)

const (
	RpcSocket = "/tmp/koality-rpc.sock"
)

func Setup(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager, rpcSocket string) error {
	server := rpc.NewServer()
	services := []interface{}{
		&PublicKeyVerifier{resourcesConnection},
		&RepositoryReader{resourcesConnection},
		&UserInfoReader{resourcesConnection},
		&VmReader{resourcesConnection, poolManager},
	}
	for _, service := range services {
		server.Register(service)
	}
	listener, err := net.Listen("unix", rpcSocket)
	if err != nil {
		return err
	}

	go server.Accept(listener)
	return nil
}
