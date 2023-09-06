package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type AESGCM struct {
	key    []byte
	aesgcm cipher.AEAD
}

func NewAESGCM(key []byte) (*AESGCM, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AESGCM{
		key:    key,
		aesgcm: aesgcm,
	}, nil
}

func (a *AESGCM) Encrypt(plainText []byte) ([]byte, error) {
	nonce := make([]byte, a.aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	cipherText := a.aesgcm.Seal(nonce, nonce, plainText, nil)
	return cipherText, nil
}

func (a *AESGCM) Decrypt(cipherText []byte) ([]byte, error) {
	nonce := cipherText[:a.aesgcm.NonceSize()]
	plainText, err := a.aesgcm.Open(nil, nonce, cipherText[a.aesgcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}
