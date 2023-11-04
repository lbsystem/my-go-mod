package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"net"
	"crypto/rand"
	"crypto/md5"
)

type CTRConn struct {
	net.Conn
	writeStream cipher.Stream
	readStream  cipher.Stream
}

func NewCTRConn(conn net.Conn, key, writeIV, readIV []byte) (*CTRConn, error) {
	// 创建 AES 密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 对 writeIV 进行 MD5 哈希
	hashedWriteIV := md5.Sum(writeIV)
	// 对 readIV 进行 MD5 哈希
	hashedReadIV := md5.Sum(readIV)

	// 创建发送方向的 CTR 密码流
	writeStream := cipher.NewCTR(block, hashedWriteIV[:])

	// 创建接收方向的 CTR 密码流
	readStream := cipher.NewCTR(block, hashedReadIV[:])

	return &CTRConn{
		Conn:        conn,
		writeStream: writeStream,
		readStream:  readStream,
	}, nil
}

func (c *CTRConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if n > 0 {
		c.readStream.XORKeyStream(b[:n], b[:n])
	}
	return n, err
}

func (c *CTRConn) Write(b []byte) (int, error) {
	buf := make([]byte, len(b))
	copy(buf, b) // 复制原始数据以避免更改输入切片
	c.writeStream.XORKeyStream(buf, buf)
	return c.Conn.Write(buf)
}

// Helper function to generate a random IV for read and write operations
func GenerateIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}
