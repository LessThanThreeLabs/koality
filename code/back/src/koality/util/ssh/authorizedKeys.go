package ssh

import (
	"errors"
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
	if err != nil {
		return
	} else if !matched {
		err = errors.New("provided publicKey is not a valid key")
		return
	}

	client, err := rpc.Dial("unix", internalapi.RpcSocket)
	if err != nil {
		return
	}

	var userId uint64
	if err = client.Call("PublicKeyVerifier.GetUserIdForKey", &publicKey, &userId); err != nil {
		return
	} else if userId == 0 {
		err = errors.New("no user with the provided public key \"" + publicKey + "\" exists")
		return
	}

	cmd = fmt.Sprintf(forcedCommand, shellPath, userId)
	return
}
