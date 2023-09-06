package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net"
)

type CTRConn struct {
	net.Conn
	stream cipher.Stream
}

func NewCTRConn(conn net.Conn, key, iv []byte) (*CTRConn, error) {
	// 创建 AES 密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	fmt.Println(aes.BlockSize)
	// 生成随机 IV
	// iv := make([]byte, aes.BlockSize)
	// if _, err := io.ReadFull(rand.Reader, iv); err != nil {
	// 	return nil, err
	// }
	// 创建 CTR 密码流
	stream := cipher.NewCTR(block, iv)
	return &CTRConn{
		Conn:   conn,
		stream: stream,
	}, nil
}

func (c *CTRConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if n > 0 {c.stream.XORKeyStream(b[:n], b[:n])}
	return n, err
}

func (c *CTRConn) Write(b []byte) (int, error) {
	c.stream.XORKeyStream(b, b)
	return c.Conn.Write(b)
}
