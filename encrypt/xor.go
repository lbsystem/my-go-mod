package encrypt

import (
	"crypto/sha512"
	"encoding/binary"
)



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

func (e *MyEncrypto) XorCipher1(data []byte) []byte {

	encrypted := make([]byte, len(data))
	keyLen := len(e.Temp)
	for i := range data {
		encrypted[i] = data[i] ^ e.Temp[i%keyLen]
	}
	return encrypted
}
func (e *MyEncrypto) XorCipher(data []byte) []byte {
    keyLen := len(e.Temp)
    encrypted := make([]byte, len(data))

    // Process 8 bytes at a time using uint64 for platforms that support it
    for i := 0; i < len(data)-7; i += 8 {
        keyChunk := binary.LittleEndian.Uint64(e.Temp[i%keyLen : i%keyLen+8])
        dataChunk := binary.LittleEndian.Uint64(data[i : i+8])
        encryptedChunk := keyChunk ^ dataChunk
        binary.LittleEndian.PutUint64(encrypted[i:i+8], encryptedChunk)
    }

    // Process remaining bytes
    for i := len(data) &^ 7; i < len(data); i++ {
        encrypted[i] = data[i] ^ e.Temp[i%keyLen]
    }

    return encrypted
}
