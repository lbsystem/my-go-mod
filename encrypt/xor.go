package encrypt

import "crypto/sha512"

type MyEncrypto struct {
	Key  []byte
	Temp []byte
}

func (e *MyEncrypto) NewEncrypto(key []byte) *MyEncrypto {
	hash := sha512.New()
	hash.Write(key)
	hash.Write([]byte("fdas19"))
	return &MyEncrypto{Key: hash.Sum(nil)}
}

func (e *MyEncrypto) XorIv(Nonce []byte) {
	encrypted := make([]byte, len(e.Key))
	keyLen := len(Nonce)
	for i := range e.Key {
		encrypted[i] = e.Key[i] ^ Nonce[i%keyLen]
	}
	e.Temp = encrypted
}

func (e *MyEncrypto) XorCipher(data []byte) []byte {

	encrypted := make([]byte, len(data))
	keyLen := len(e.Temp)
	for i := range data {
		encrypted[i] = data[i] ^ e.Temp[i%keyLen]
	}
	return encrypted
}
