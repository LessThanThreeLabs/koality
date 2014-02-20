package notify

import (
	"fmt"
	"koality/resources"
	"koality/vm"
	"strings"
)

type Notifier interface {
	BuildStatusNotifier
	DebugInstanceNotifier
}

type BuildStatusNotifier interface {
	NotifyBuildStatus(build *resources.Build) error
}

type DebugInstanceNotifier interface {
	NotifyDebugInstance(vm vm.VirtualMachine, build *resources.Build, debugInstance *resources.DebugInstance) error
}

var htmlSanitizer = strings.NewReplacer("\n", "<br>", " ", "&nbsp;", "\t", "&nbsp;&nbsp;&nbsp;&nbsp;")

func getBuildUri(domainName resources.DomainName, repositoryId, buildId uint64) string {
	return fmt.Sprintf("https://%s/repository/%d?build=%d", domainName, repositoryId, buildId)
}
