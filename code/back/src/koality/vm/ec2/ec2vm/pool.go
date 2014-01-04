package ec2vm

import (
	"koality/resources"
	"koality/vm"
)

type Ec2VirtualMachinePool struct {
	vm.VirtualMachinePool
	Ec2VirtualMachineLauncher *Ec2VirtualMachineLauncher
}

func NewPool(virtualMachineLauncher *Ec2VirtualMachineLauncher) *Ec2VirtualMachinePool {
	virtualMachinePool := vm.NewPool(virtualMachineLauncher.Ec2Pool.Id, virtualMachineLauncher, virtualMachineLauncher.Ec2Pool.NumReadyInstances, virtualMachineLauncher.Ec2Pool.NumMaxInstances)
	return &Ec2VirtualMachinePool{virtualMachinePool, virtualMachineLauncher}
}

func (virtualMachinePool *Ec2VirtualMachinePool) UpdateSettings(ec2Pool resources.Ec2Pool) {
	*virtualMachinePool.Ec2VirtualMachineLauncher.Ec2Pool = ec2Pool

	virtualMachinePool.SetMaxSize(ec2Pool.NumMaxInstances)
	virtualMachinePool.SetMinReady(ec2Pool.NumReadyInstances)
}
