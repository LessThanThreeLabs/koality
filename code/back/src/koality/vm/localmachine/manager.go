package localmachine

import (
	"koality/vm"
)

type localMachineManager struct{}

var Manager = localMachineManager{}

func (manager localMachineManager) NewVirtualMachine() (vm.VirtualMachine, error) {
	return New()
}

func (manager localMachineManager) GetVirtualMachine(rootDir string) (vm.VirtualMachine, error) {
	return FromDir(rootDir)
}
