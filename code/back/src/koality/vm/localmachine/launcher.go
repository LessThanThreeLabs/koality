package localmachine

import (
	"koality/vm"
)

type localMachineLauncher struct{}

var Launcher = localMachineLauncher{}

func (launcher localMachineLauncher) LaunchVirtualMachine() (vm.VirtualMachine, error) {
	return New(), nil
}
