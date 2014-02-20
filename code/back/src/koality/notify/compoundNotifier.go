package notify

import (
	"koality/resources"
	"koality/vm"
)

type CompoundNotifier struct {
	buildStatusNotifiers   []BuildStatusNotifier
	debugInstanceNotifiers []DebugInstanceNotifier
}

func NewCompoundNotifier(buildStatusNotifiers []BuildStatusNotifier, debugInstanceNotifiers []DebugInstanceNotifier) Notifier {
	return &CompoundNotifier{buildStatusNotifiers, debugInstanceNotifiers}
}

func (compoundNotifier *CompoundNotifier) NotifyBuildStatus(build *resources.Build) error {
	var notifyErr error
	for _, notifier := range compoundNotifier.buildStatusNotifiers {
		if err := notifier.NotifyBuildStatus(build); err != nil {
			notifyErr = err
		}
	}
	return notifyErr
}

func (compoundNotifier *CompoundNotifier) NotifyDebugInstance(vm vm.VirtualMachine, build *resources.Build, debugInstance *resources.DebugInstance) error {
	var notifyErr error
	for _, notifier := range compoundNotifier.debugInstanceNotifiers {
		if err := notifier.NotifyDebugInstance(vm, build, debugInstance); err != nil {
			notifyErr = err
		}
	}
	return notifyErr
}
