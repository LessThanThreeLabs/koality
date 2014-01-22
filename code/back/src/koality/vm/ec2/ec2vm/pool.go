package ec2vm

import (
	"koality/resources"
	"koality/vm"
)

type Ec2VirtualMachinePool struct {
	vm.VirtualMachinePool
	Ec2VirtualMachineManager *Ec2VirtualMachineManager
}

func NewPool(virtualMachineManager *Ec2VirtualMachineManager) *Ec2VirtualMachinePool {
	virtualMachinePool := vm.NewPool(virtualMachineManager.Ec2Pool.Id, virtualMachineManager, virtualMachineManager.Ec2Pool.NumReadyInstances, virtualMachineManager.Ec2Pool.NumMaxInstances)
	return &Ec2VirtualMachinePool{virtualMachinePool, virtualMachineManager}
}

func (virtualMachinePool *Ec2VirtualMachinePool) UpdateSettings(ec2Pool resources.Ec2Pool) {
	*virtualMachinePool.Ec2VirtualMachineManager.Ec2Pool = ec2Pool

	virtualMachinePool.SetMaxSize(ec2Pool.NumMaxInstances)
	virtualMachinePool.SetMinReady(ec2Pool.NumReadyInstances)
}
