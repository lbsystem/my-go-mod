package proxytools

import (
	"context"
	"errors"

	"fmt"

	// "io"
	"net"
	"sync"
	"time"

	tproxy "github.com/KatelynHaworth/go-tproxy"
)

type UdpConn struct {
	net.PacketConn
	localAddr  net.Addr
	remote     net.Addr
	singal     chan int
	mu         sync.Mutex
	MTU        int
	ctx        context.Context
	cancel     context.CancelFunc
	tmpData    []byte
	timeout    *time.Timer
	isNotfirst bool
}

var connMap = make(map[string]*UdpConn)
var connMapMu sync.Mutex
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
	if u.isNotfirst {
		u.tmpData = b
	}

	select {
	case <-u.ctx.Done():
		u.timeout = nil
		return 0, errors.New("time out")
	case n, ok := <-u.singal:
		if !ok {
			return 0, fmt.Errorf("connection closed")
		}
		if !u.isNotfirst {
			u.isNotfirst = true
			return copy(b, u.tmpData[:n]), nil
		} else {
			return n, nil
		}

	}

}
func (u *UdpConn) SetReadDeadline(t time.Duration) error {

	if u.timeout == nil {
		u.timeout = time.NewTimer(t)
		go u.handleTimeout()
		return nil
	} else {
		u.timeout.Reset(t)
	}

	return nil

}
func (u *UdpConn) Close() error {
	fmt.Println("close")
	close(u.singal)
	u.mu.Lock()
	defer u.mu.Unlock()

	connMapMu.Lock()
	delete(connMap, u.remote.String())
	connMapMu.Unlock()

	u.cancel() // 取消相关的 goroutine

	return nil
}

type UdpForConn struct {
	Conn   net.PacketConn
	addr   string
	mu     sync.Mutex
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
func (u *UdpForConn) runAccept(mode ...string) {
	b := make([]byte, 1500)
	defer u.mu.Unlock()
	for {
		var addr net.Addr
		var originalDst net.Addr
		var n int
		var err error
		u.mu.Lock()
		isTproxy := (len(mode) > 0 && mode[0] == "tproxy")
		if isTproxy {
			n, addr, originalDst, err = tproxy.ReadFromUDP(u.TpConn, b)

			if err != nil {
				fmt.Println(err.Error())
				u.mu.Unlock()
				continue
			}
		} else {
			n, addr, err = u.Conn.ReadFrom(b)
			if err != nil {
				fmt.Println(err.Error())
				u.mu.Unlock()
				continue
			}
		}
		connMapMu.Lock()
		conn, ok := connMap[addr.String()]
		if !ok {
			if isTproxy {
				conn = NewUdpConn(u.Conn, originalDst, addr, u.MTU)
			} else {
				conn = NewUdpConn(u.Conn, u.Conn.LocalAddr(), addr, u.MTU)
			}

			connMap[addr.String()] = conn
			cc <- conn
			newconn, ok := connMap[addr.String()]
			if ok {
				fmt.Println("new conn")
				copy(newconn.tmpData, b[:n])
				u.mu.Unlock()
				newconn.singal <- n
			} else {
				u.mu.Unlock()
			}

		}
		connMapMu.Unlock()
		if ok {
			n := copy(conn.tmpData, b[:n])
			u.mu.Unlock()
			conn.singal <- n
		}

	}
}
func NewUdpForConn(addr string, mtu int, mode ...string) (*UdpForConn, error) {
	var conn *net.UDPConn
	var newconn *UdpForConn
	u, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	if len(mode) > 0 && mode[0] == "tproxy" {

		conn, err = tproxy.ListenUDP("udp", u)
		if err != nil {
			return nil, err
		}
		newconn = &UdpForConn{addr: addr, MTU: mtu, Conn: conn, TpConn: conn}
	} else {
		conn, err = net.ListenUDP("udp", u)
		if err != nil {
			return nil, err
		}
		newconn = &UdpForConn{addr: addr, MTU: mtu, Conn: conn}
	}

	go newconn.runAccept(mode...)
	return newconn, nil
}

func ChangeUdpToConn(conn net.PacketConn, addr string, mtu int) (*UdpForConn, error) {

	newconn := &UdpForConn{addr: addr, MTU: mtu, Conn: conn}
	go newconn.runAccept()
	return newconn, nil
}

func Test() {
	fmt.Println("ok")
	udpCUdpForConn, err := NewUdpForConn("0.0.0.0:8080", 1400, "tproxy")
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
			c, err2 := net.Dial("tcp", "xxx.xxx:33326")
			fmt.Printf("uc.LocalAddr(): %v\n", uc.localAddr)
			if err2 != nil {
				fmt.Println(err2.Error())
				uc.Close()
				return
			}

			// go io.Copy(c, uc)
			// io.Copy(uc, c)
			defer c.Close()
			defer uc.Close()

			b := make([]byte, 3500)
			for {
				uc.SetReadDeadline(time.Second * 15)
				n, err2 := uc.Read(b)
				c.Write(b[:n])
				if err2 != nil {
					fmt.Println(err2.Error())
					fmt.Println("will close")

					return
				}
				fmt.Println(b[:n])
			}
		}()

	}

}
