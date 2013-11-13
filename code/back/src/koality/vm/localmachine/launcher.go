package localmachine

import (
	"koality/vm"
)

type LocalMachineLauncher struct{}

func NewLauncher() *LocalMachineLauncher {
	return new(LocalMachineLauncher)
}

func (launcher *LocalMachineLauncher) LaunchVirtualMachine() (vm.VirtualMachine, error) {
	return New(), nil
}
