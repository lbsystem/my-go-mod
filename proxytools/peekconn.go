package proxytools

import (
	"bufio"
	"net"
	"time"
)

type BufferedConn struct {
	r   *bufio.Reader
	raw net.Conn
}

// 创建一个新的bufferedConn
func NewMyConn(c net.Conn) BufferedConn {
	return BufferedConn{bufio.NewReader(c), c}
}

// 实现net.Conn的Read方法
func (b BufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// 实现net.Conn的Write方法
func (b BufferedConn) Write(p []byte) (n int, err error) {
	return b.raw.Write(p)
}

// 实现net.Conn的Close方法
func (b BufferedConn) Close() error {
	return b.raw.Close()
}

// 实现net.Conn的LocalAddr方法
func (b BufferedConn) LocalAddr() net.Addr {
	return b.raw.LocalAddr()
}

// 实现net.Conn的RemoteAddr方法
func (b BufferedConn) RemoteAddr() net.Addr {
	return b.raw.RemoteAddr()
}

// 实现net.Conn的SetDeadline方法
func (b BufferedConn) SetDeadline(t time.Time) error {
	return b.raw.SetDeadline(t)
}

// 实现net.Conn的SetReadDeadline方法
func (b BufferedConn) SetReadDeadline(t time.Time) error {
	return b.raw.SetReadDeadline(t)
}

// 实现net.Conn的SetWriteDeadline方法
func (b BufferedConn) SetWriteDeadline(t time.Time) error {
	return b.raw.SetWriteDeadline(t)
}

// 额外的Peek方法
func (b BufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}
