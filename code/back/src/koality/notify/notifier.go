package notify

import (
	"koality/build/runner"
	"koality/resources"
	"koality/vm"
)

type Notifier interface {
	Notify(vm vm.VirtualMachine, build *resources.Build, buildData *runner.BuildData) error
}
