package ssh

import (
	"fmt"
	"github.com/LessThanThreeLabs/gocheck"
	"koality/internalapi"
	"koality/repositorymanager"
	"koality/util/pathtranslator"
	"koality/util/ssh/test"
	"koality/vm"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type SshSuite struct {
	client *rpc.Client
}

var _ = gocheck.Suite(&SshSuite{})

func rpcSetup(check *gocheck.C) {
	server := rpc.NewServer()
	services := []interface{}{
		new(test.PublicKeyVerifier),
		new(test.RepositoryReader),
		new(test.UserInfoReader),
		new(test.VmReader),
	}
	for _, service := range services {
		check.Assert(server.Register(service), gocheck.IsNil)
	}

	listener, err := net.Listen("unix", internalapi.RpcSocket)
	check.Assert(err, gocheck.IsNil)
	go server.Accept(listener)
}

func (suite *SshSuite) SetUpSuite(check *gocheck.C) {
	rpcSetup(check)

	// REVIEW(dhuang) is there a better way to do this?
	socketOpen := false
	for i := 0; i < 42 && !socketOpen; i++ {
		if _, err := os.Stat(internalapi.RpcSocket); err == nil {
			socketOpen = true
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !socketOpen {
		check.Fatalf("socket took too long to exist")
	}

	compileCmd := exec.Command("go", "install", path.Join("koality", "util", "xunit", "getXunitResults"))
	err := compileCmd.Run()
	check.Assert(err, gocheck.IsNil)

	suite.client, err = rpc.Dial("unix", internalapi.RpcSocket)
	check.Assert(err, gocheck.IsNil)
}
func (suite *SshSuite) TearDownSuite(check *gocheck.C) {
	if suite.client != nil {
		suite.client.Close()
	}

	os.Remove(internalapi.RpcSocket)
}

func (suite *SshSuite) TestGetForcedCommand(check *gocheck.C) {
	shellPath := "shellpath"
	publicKey := "ssh-rsa abc"
	forcedCommandRes, err := GetForcedCommand(shellPath, publicKey)
	check.Assert(err, gocheck.IsNil)
	check.Assert(forcedCommandRes, gocheck.Equals,
		fmt.Sprintf(forcedCommand, shellPath, 69))
}

func (suite *SshSuite) TestGitShell(check *gocheck.C) {
	userId := uint64(69)
	localUri := "koality-v1"
	repositoryManager := repositorymanager.New(test.RepositoriesPath, nil)

	command := []string{"git-receive-pack", localUri}
	shell := restrictedGitShell{userId, command, suite.client}
	commandToExec, err := shell.GetCommand()
	check.Assert(err, gocheck.IsNil)
	check.Assert(commandToExec, gocheck.DeepEquals, vm.Command{
		Argv: []string{"jgit", "receive-pack",
			repositoryManager.ToPath(&test.Repository),
			fmt.Sprint(userId)},
	})

	shell.command = []string{"git-upload-pack", localUri}
	commandToExec, err = shell.GetCommand()
	check.Assert(err, gocheck.IsNil)
	sshwrapperPath, err := pathtranslator.TranslatePath(pathtranslator.BinaryPath("sshwrapper"))
	check.Assert(err, gocheck.IsNil)
	check.Assert(commandToExec, gocheck.DeepEquals, vm.Command{
		Argv: []string{sshwrapperPath, "git@github.com", "git-upload-pack KoalityCode/koality-v1.git"},
		Envv: append([]string{"PRIVATE_KEY=" + test.PrivateKey}),
	})
}

func (suite *SshSuite) TestSSHForwardingShell(check *gocheck.C) {
	userId := uint64(69)
	command := []string{"ssh", "instanceid", "555"}
	shell := restrictedSSHForwardingShell{userId, command, suite.client}
	commandToExec, err := shell.GetCommand()
	check.Assert(err, gocheck.IsNil)
	check.Assert(commandToExec, gocheck.DeepEquals, vm.Command{
		Argv: []string{"ssh", "-oLogLevel=error",
			"-oStrictHostKeyChecking=no", "-oUserKnownHostsFile=/dev/null",
			"-oServerAliveInterval=20", "koality@koalitycode.com", "-p", "22",
			"''"},
		Envv: append([]string{"PRIVATE_KEY=" + test.PrivateKey}, os.Environ()...),
	})
}
