package vm

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"koality/shell"
	"os"
	"os/user"
	"path"
	"strconv"
)

type SshExecutableMaker struct {
	sshClient *ssh.ClientConn
}

type sshExecutable struct {
	shell.Command
	*ssh.Session
}

type keychain struct {
	key *rsa.PrivateKey
}

func NewSshExecutableMaker(sshConfig SshConfig) (*SshExecutableMaker, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	privateKey, err := ioutil.ReadFile(fmt.Sprintf("%s/.ssh/id_rsa", currentUser.HomeDir))
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privateKey)
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
	return &SshExecutableMaker{sshClient}, nil
}

func (sshExecutableMaker *SshExecutableMaker) MakeExecutable(command shell.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, environment map[string]string) (shell.Executable, error) {
	session, err := sshExecutableMaker.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr

	for key, value := range environment {
		err = session.Setenv(key, value)
		if err != nil {
			return nil, err
		}
	}

	return &sshExecutable{command, session}, nil
}

func (sshExecutableMaker *SshExecutableMaker) Close() error {
	return sshExecutableMaker.sshClient.Close()
}

func (executable *sshExecutable) Start() error {
	return executable.Session.Start(string(executable.Command))
}

func (executable *sshExecutable) Run() error {
	defer executable.Session.Close()
	return executable.Session.Run(string(executable.Command))
}

func (executable *sshExecutable) Wait() error {
	defer executable.Session.Close()
	return executable.Session.Wait()
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

type Scper interface {
	Scp(localFilePath, remoteFilePath string) (shell.Executable, error)
}

type sshScper struct {
	*SshExecutableMaker
}

func NewScper(config ScpConfig) (Scper, error) {
	sshExecutableMaker, err := NewSshExecutableMaker(SshConfig(config))
	if err != nil {
		return nil, err
	}
	return &sshScper{sshExecutableMaker}, nil
}

func (scper *sshScper) Scp(localFilePath, remoteFilePath string) (shell.Executable, error) {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return nil, err
	}
	fileInfo, err := localFile.Stat()
	if err != nil {
		return nil, err
	}
	headerBuffer := bytes.NewBufferString(fmt.Sprintf("C%#o %d %s\n", fileInfo.Mode()&os.ModePerm, fileInfo.Size(), path.Base(remoteFilePath)))
	scpStdin := io.MultiReader(headerBuffer, localFile, bytes.NewReader([]byte{0}))

	// Note: this is more powerful than standard scp, as it will actually create the destination directory for you
	remoteCommand := shell.And(
		shell.Commandf("mkdir -p %s", path.Dir(remoteFilePath)),
		shell.Commandf("scp -qrt %s", path.Dir(remoteFilePath)),
	)
	return scper.SshExecutableMaker.MakeExecutable(remoteCommand, scpStdin, nil, nil, nil)
}

type SshConfig struct {
	Username string
	Hostname string
	Port     int
	Options  map[string]string
}

type ScpConfig SshConfig

func toOptionsList(options map[string]string) []string {
	optionsList := make([]string, len(options))

	index := 0
	for key, value := range options {
		optionsList[index] = fmt.Sprintf("-o%s=%s", key, value)
		index++
	}
	return optionsList
}

func (sshConfig SshConfig) SshArgs(remoteCommand string) []string {
	options := toOptionsList(sshConfig.Options)
	login := fmt.Sprintf("%s@%s", sshConfig.Username, sshConfig.Hostname)
	args := append(options, login, "-p", strconv.Itoa(sshConfig.Port), shell.Quote(remoteCommand))

	return append([]string{"ssh"}, args...)
}

func (scpConfig ScpConfig) ScpArgs(localFilePath, remoteFilePath string) []string {
	options := toOptionsList(scpConfig.Options)
	remotePath := fmt.Sprintf("%s@%s:%s", scpConfig.Username, scpConfig.Hostname, remoteFilePath)
	return append(append([]string{"scp"}, options...), "-P", strconv.Itoa(scpConfig.Port), shell.Quote(localFilePath), shell.Quote(remotePath))
}

type ScpFileCopier struct {
	Scper
}

func (fileCopier *ScpFileCopier) FileCopy(localFilePath, remoteFilePath string) (shell.Executable, error) {
	return fileCopier.Scper.Scp(localFilePath, remoteFilePath)
}
