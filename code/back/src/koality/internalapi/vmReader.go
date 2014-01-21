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

func (vmReader VmReader) GetVmFromId(arg VmReaderArg, vm *vm.VirtualMachine) error {
	pool, err := vmReader.poolManager.GetPool(arg.PoolId)
	if err != nil {
		return err
	}

	*vm, err = pool.GetExisting(arg.VmInstanceId)
	return err
}
