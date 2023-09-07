type MyEncrypto struct{
	Key []byte
	
}


func (e *MyEncrypto) XorIv(Nonce []byte ){
	encrypted := make([]byte, len(e.Key))
	keyLen := len(Nonce)
	for i := range e.Key {
		encrypted[i] = e.Key[i] ^ Nonce[i%keyLen]
	}
	e.Key=encrypted
}

func (e *MyEncrypto) XorCipher(data []byte) []byte {
	packetLength := len(data)
	packetLengthByte := byte(packetLength) // 包长转成ASCII字符

	// 用包长改变密钥
	modifiedKey := make([]byte, len(e.Key))
	for i, keyByte := range e.Key {
		modifiedKey[i] = keyByte ^ packetLengthByte
	}

	// 用修改后的密钥进行XOR加密
	encrypted := make([]byte, packetLength)
	keyLen := len(modifiedKey)
	for i := range data {
		j := i % keyLen
		encrypted[i] = data[i] ^ modifiedKey[j]
	}
	return encrypted
}