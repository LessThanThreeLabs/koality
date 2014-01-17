package settings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type Encrypter struct {
	blockCipher cipher.Block
}

func NewEncrypter() (*Encrypter, error) {
	aesBlockCipher, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	return &Encrypter{aesBlockCipher}, nil
}

func (encrypter *Encrypter) generateRandomIv(length int) ([]byte, error) {
	iv := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

func (encrypter *Encrypter) EncryptValue(plainTextValue []byte) ([]byte, error) {
	iv, err := encrypter.generateRandomIv(encrypter.blockCipher.BlockSize())
	if err != nil {
		return nil, err
	}

	encryptedValue := make([]byte, len(plainTextValue))
	cipherStream := cipher.NewCTR(encrypter.blockCipher, iv)
	cipherStream.XORKeyStream(encryptedValue, plainTextValue)
	return append(iv, encryptedValue...), nil
}

func (encrypter *Encrypter) DecryptValue(encryptedValue []byte) ([]byte, error) {
	blockSize := encrypter.blockCipher.BlockSize()
	iv := encryptedValue[:blockSize]
	encryptedValue = encryptedValue[blockSize:]

	decryptedValue := make([]byte, len(encryptedValue))
	cipherStream := cipher.NewCTR(encrypter.blockCipher, iv)
	cipherStream.XORKeyStream(decryptedValue, encryptedValue)
	return decryptedValue, nil
}
