package internalapi

import (
	"koality/vm/poolmanager"
	"koality/resources"
	"net"
	"net/rpc"
)

const (
	RpcSocket = "/tmp/koality-rpc.sock"
)

func Start(resourcesConnection *resources.Connection, poolManager *poolmanager.PoolManager, repositoriesPath string, rpcSocket string) error {
	server := rpc.NewServer()
	services := []interface{}{
		&PublicKeyVerifier{resourcesConnection},
		&RepositoryReader{resourcesConnection, repositoriesPath},
		&UserInfoReader{resourcesConnection},
		&VmReader{resourcesConnection, poolManager},
	}
	for _, service := range services {
		if err := server.Register(service); err != nil {
			return err
		}
	}
	listener, err := net.Listen("unix", rpcSocket)
	if err != nil {
		return err
	}

	go server.Accept(listener)
	return nil
}
