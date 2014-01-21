package ssh

import (
	"fmt"
	"koality/internalapi"
	"net/rpc"
	"regexp"
)

const (
	forcedCommand = `command="%s %d",no-port-forwarding,no-X11-forwarding,no-agent-forwarding`
	validKey      = "^ssh-(?:dss|rsa) [A-Za-z0-9+/]+={0,2}$"
)

func GetForcedCommand(shellPath, publicKey string) (cmd string, err error) {
	matched, err := regexp.MatchString(validKey, publicKey)
	if err != nil || !matched {
		return
	}

	client, err := rpc.Dial("unix", internalapi.RpcSocket)
	if err != nil {
		return
	}

	var userId uint64
	if err = client.Call("PublicKeyVerifier.GetUserIdForKey", &publicKey, &userId); err != nil {
		return
	}
	if err != nil {
		return
	}

	cmd = fmt.Sprintf(forcedCommand, shellPath, userId)
	return
}
