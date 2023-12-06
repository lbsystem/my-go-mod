package proxytools

import (
	"context"
	"errors"
	"fmt"

	// "io"
	"net"
	"sync"
	"time"
)

type UdpConn struct {
	net.PacketConn
	localAddr  net.Addr
	remote     net.Addr
	singal     chan int
	isNotFirst bool
	MTU        int
	ctx        context.Context
	cancel     context.CancelFunc
	tmpData    []byte
	timeout    *time.Timer
}

var connMap sync.Map

var cc = make(chan *UdpConn, 32)

func NewUdpConn(udpConn net.PacketConn, localAddr, remote net.Addr, mtu int) *UdpConn {
	ctx, cancel := context.WithCancel(context.Background())

	return &UdpConn{
		PacketConn: udpConn,
		localAddr:  localAddr,
		remote:     remote,
		singal:     make(chan int, 1),
		MTU:        mtu,
		ctx:        ctx,
		cancel:     cancel,
		tmpData:    make([]byte, 1500),
	}
}
func (u *UdpConn) LocalAddr() net.Addr {
	return u.localAddr
}

// RemoteAddr returns the remote network address, if known.
func (u *UdpConn) RemoteAddr() net.Addr {
	return u.remote
}

func (u *UdpConn) Write(b []byte) (n int, err error) {
	totalLen := len(b)
	for i := 0; i < totalLen; i += u.MTU {
		end := i + u.MTU
		if end > totalLen {
			end = totalLen
		}
		nn, err := u.PacketConn.WriteTo(b[i:end], u.remote)
		if err != nil {
			return n, err
		}
		n += nn
	}
	return n, nil
}
func (u *UdpConn) handleTimeout() {
	_, ok := <-u.timeout.C
	if ok {
		u.cancel()
	}

}
func (u *UdpConn) Read(b []byte) (n int, err error) {

	if !u.isNotFirst {
		u.isNotFirst = true
		n, ok := <-u.singal
		if ok {
			nn := copy(b, u.tmpData[:n])
			u.tmpData = b
			return nn, nil
		} else {
			return 0, errors.New("error")
		}
	}
	select {
	case <-u.ctx.Done():
		u.timeout = nil
		return 0, errors.New("time out")
	case n, ok := <-u.singal:
		if !ok {
			return 0, fmt.Errorf("connection closed")
		}
		return n, nil
	}
}
func (u *UdpConn) SetReadDeadline(t time.Duration) error {
	if u.timeout == nil {
		u.timeout = time.NewTimer(t)
		go u.handleTimeout()
		return nil
	}
	u.timeout.Reset(t)
	return nil
}
func (u *UdpConn) Close() error {
	fmt.Println("close")
	close(u.singal)
	connMap.Delete(u.remote.String())
	u.cancel() // 取消相关的 goroutine
	return nil
}

type UdpForConn struct {
	Conn   net.PacketConn
	addr   string
	MTU    int
	TpConn *net.UDPConn
}

func (u *UdpForConn) Accept() (*UdpConn, error) {
	newConn, ok := <-cc
	if !ok {
		return nil, fmt.Errorf("accept channel closed")
	}
	return newConn, nil
}
func (u *UdpForConn) runAccept() {
	b := make([]byte, 1500)

	for {
		var addr net.Addr
		var conn *UdpConn
		var n int
		var err error

		n, addr, err = u.Conn.ReadFrom(b)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		v, ok := connMap.Load(addr.String())
		if !ok {
			fmt.Println("new conn")
			conn = NewUdpConn(u.Conn, u.Conn.LocalAddr(), addr, u.MTU)
			connMap.Store(addr.String(), conn)
			cc <- conn
		}
		if ok {
			newconn, _ := v.(*UdpConn)
			fmt.Println("old conn")
			copy(newconn.tmpData, b[:n])
			newconn.singal <- n
		} else {
			n := copy(conn.tmpData, b[:n])
			conn.singal <- n
		}
	}
}
func NewUdpForConn(addr string, mtu int) (*UdpForConn, error) {
	var conn *net.UDPConn
	var newconn *UdpForConn
	u, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err = net.ListenUDP("udp", u)
	if err != nil {
		return nil, err
	}
	newconn = &UdpForConn{addr: addr, MTU: mtu, Conn: conn}

	go newconn.runAccept()
	return newconn, nil
}

func ChangeUdpToConn(conn net.PacketConn, addr string, mtu int) (*UdpForConn, error) {

	newconn := &UdpForConn{addr: addr, MTU: mtu, Conn: conn}
	go newconn.runAccept()
	return newconn, nil
}

func main() {

	udpCUdpForConn, err := NewUdpForConn("0.0.0.0:8080", 1400)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		fmt.Println("start")
		uc, err := udpCUdpForConn.Accept()
		fmt.Print("accept")
		if err != nil {
			fmt.Println(err.Error())
		}

		go func() {
			defer uc.Close()
			b := make([]byte, 3500)
			for {
				uc.SetReadDeadline(time.Second * 3)
				n, err2 := uc.Read(b)
				if err2 != nil {
					fmt.Println(err2.Error())
					fmt.Println("will close")
					break
				}
				fmt.Println(b[:n])
				uc.Write(b[:n])
			}
		}()
	}

}
