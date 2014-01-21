package resources

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os/exec"
	"strings"
)

type KeyPairGenerator struct {
}

func NewKeyPairGenerator() (*KeyPairGenerator, error) {
	return &KeyPairGenerator{}, nil
}

func (keyPairGenerator *KeyPairGenerator) GenerateRepositoryKeyPair() (*RepositoryKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}

	privateKeyPem, err := keyPairGenerator.getPrivatePem(privateKey)
	if err != nil {
		return nil, err
	}

	publicKeyPem, err := keyPairGenerator.getPublicPem(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicSshKey, err := keyPairGenerator.getPublicSshKey(publicKeyPem)
	if err != nil {
		return nil, err
	}

	repositoryKeyPair := RepositoryKeyPair{
		PrivateKey: privateKeyPem,
		PublicKey:  publicSshKey,
	}
	return &repositoryKeyPair, nil
}

func (keyPairGenerator *KeyPairGenerator) getPrivatePem(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	return string(pem.EncodeToMemory(&privateKeyBlock)), nil
}

func (keyPairGenerator *KeyPairGenerator) getPublicPem(publicKey *rsa.PublicKey) (string, error) {
	publicKeyDer, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDer,
	}
	return string(pem.EncodeToMemory(&publicKeyBlock)), nil
}

func (keyPairGenerator *KeyPairGenerator) getPublicSshKey(publicPem string) (string, error) {
	command := exec.Command("ssh-keygen", "-m", "PKCS8", "-f", "/dev/stdin", "-i")
	command.Stdin = strings.NewReader(publicPem)
	output, err := command.CombinedOutput()
	if _, ok := err.(*exec.ExitError); ok {
		return "", fmt.Errorf("%v: %s", err, output)
	} else if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
