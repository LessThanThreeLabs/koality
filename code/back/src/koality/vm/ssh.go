package vm

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/LessThanThreeLabs/go.crypto/ssh"
	"github.com/LessThanThreeLabs/go.crypto/ssh/terminal"
	"io"
	"koality/shell"
	"os"
	"strconv"
)

type SshExecutor struct {
	sshClient *ssh.ClientConn
}

type sshExecution struct {
	*ssh.Session
}

type keychain struct {
	key *rsa.PrivateKey
}

func NewSshExecutor(sshConfig SshConfig) (*SshExecutor, error) {
	block, _ := pem.Decode([]byte(sshConfig.PrivateKey))
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	clientKey := keychain{rsaKey}
	clientConfig := ssh.ClientConfig{
		User: sshConfig.Username,
		Auth: []ssh.ClientAuth{ssh.ClientAuthKeyring(&clientKey)},
	}
	address := fmt.Sprintf("%s:%d", sshConfig.Hostname, sshConfig.Port)
	sshClient, err := ssh.Dial("tcp", address, &clientConfig)
	if err != nil {
		return nil, err
	}
	return &SshExecutor{sshClient}, nil
}

func (executor *SshExecutor) Execute(executable shell.Executable) (shell.Execution, error) {
	session, err := executor.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	session.Stdin = executable.Stdin
	session.Stdout = executable.Stdout
	session.Stderr = executable.Stderr

	terminalHeight, terminalWidth, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		terminalHeight, terminalWidth = 80, 40
	}

	terminalModes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err = session.RequestPty("xterm", terminalHeight, terminalWidth, terminalModes); err != nil {
		session.Close()
		return nil, err
	}

	envCommands := make([]shell.Command, 0, len(executable.Environment))
	for key, value := range executable.Environment {
		envCommands = append(envCommands, shell.Commandf("export %s=%s", key, shell.Quote(value)))
	}

	commandWithEnv := shell.Chain(append(envCommands, executable.Command)...)

	err = session.Start(string(commandWithEnv))
	if err != nil {
		session.Close()
		return nil, err
	}

	return &sshExecution{session}, nil
}

func (sshExecutor *SshExecutor) Close() error {
	return sshExecutor.sshClient.Close()
}

func (execution *sshExecution) Wait() error {
	return execution.Session.Wait()
}

func (k *keychain) Key(i int) (ssh.PublicKey, error) {
	if i != 0 {
		return nil, nil
	}
	return ssh.NewPublicKey(&k.key.PublicKey)
}

func (k *keychain) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	hashFunc := crypto.SHA1
	h := hashFunc.New()
	h.Write(data)
	digest := h.Sum(nil)
	return rsa.SignPKCS1v15(rand, k.key, hashFunc, digest)
}

type SshConfig struct {
	Username   string
	Hostname   string
	Port       int
	PrivateKey string
	Options    map[string]string
}

func toOptionsList(options map[string]string) []string {
	optionsList := make([]string, len(options))

	index := 0
	for key, value := range options {
		optionsList[index] = fmt.Sprintf("-o%s=%s", key, value)
		index++
	}
	return optionsList
}

// TODO: make this use the specified private key somehow, or maybe just get rid of this
func (sshConfig SshConfig) SshArgs(remoteCommand string) []string {
	options := toOptionsList(sshConfig.Options)
	login := fmt.Sprintf("%s@%s", sshConfig.Username, sshConfig.Hostname)
	args := append(options, login, "-p", strconv.Itoa(sshConfig.Port), shell.Quote(remoteCommand))

	return append([]string{"ssh"}, args...)
}
