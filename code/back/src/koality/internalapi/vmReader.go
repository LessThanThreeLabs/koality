package internalapi

import (
	"koality/resources"
	"koality/vm"
	"koality/vm/poolmanager"
)

type VmReader struct {
	resourcesConnection *resources.Connection
	poolManager         *poolmanager.PoolManager
}

type VmReaderArg struct {
	VmInstanceId string
	PoolId       uint64
}

func (vmReader VmReader) GetShellCommandFromId(arg VmReaderArg, command *vm.Command) error {
	pool, err := vmReader.poolManager.GetPool(arg.PoolId)
	if err != nil {
		return err
	}

	vm, err := pool.GetExisting(arg.VmInstanceId)
	if err != nil {
		return err
	}

	*command = vm.GetStartShellCommand()
	return err
}
