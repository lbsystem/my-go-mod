package rawConn

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lbsystem/my-go-mod/encrypt"
	"github.com/lbsystem/my-go-mod/udp/codec"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/bpf"
	"golang.org/x/net/ipv4"
)

type RawPackConn struct {
	*ipv4.RawConn
	ipHeader    *ipv4.Header
	srcIP       *net.UDPAddr
	fd          int
	nonce       []byte
	encryptMode bool
}

func (c *RawPackConn) Close() error {
	return unix.Close(c.fd)
}

var en *encrypt.MyEncrypto

func (c *RawPackConn) ReadFrom(b []byte) (int, net.Addr, error) {
	h, newByte, _, err := c.RawConn.ReadFrom(b)
	if err != nil {
		return 0, nil, err
	}
	port1:=int(binary.BigEndian.Uint16(newByte[0:2]))
	var n int
	if c.encryptMode {
		iv := make([]byte, 2)
		binary.LittleEndian.PutUint16(iv, uint16(h.ID))
		en.XorIv(iv)
		newByte = en.XorCipher(newByte[8:])
		n = copy(b, newByte)
	} else {
		n = copy(b, newByte[8:])
	}

	return n, &net.UDPAddr{
		IP:   h.Src,
		Port: port1,
	}, err
}

func (c *RawPackConn) WriteTo(b []byte, dstaddr net.Addr) (int, error) {
	
	dstAdrr, ok := dstaddr.(*net.UDPAddr)
	if !ok {
		return 0, errors.New("dstaddr is not a net.UDPAddr")
	}
	ipId := codec.GenerateRandomPort()
	packetLen := len(b)
	if c.encryptMode {
		binary.LittleEndian.PutUint16(c.nonce, uint16(ipId))
		en.XorIv(c.nonce)
		b = en.XorCipher(b)
	}

	c.ipHeader.Dst = dstAdrr.IP
	c.ipHeader.ID = ipId
	c.ipHeader.TotalLen = 20 + 8 + packetLen
	b = codec.BuildUDPPacket(c.srcIP, dstAdrr, b)

	err := c.RawConn.WriteTo(c.ipHeader, b, nil)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return packetLen, nil
}

func UDPAddrToSockaddr(addr *net.UDPAddr) unix.Sockaddr {
	if addr.IP.To4() != nil {
		// IPv4
		var addr4 [4]byte
		copy(addr4[:], addr.IP.To4())
		return &unix.SockaddrInet4{
			Port: addr.Port,
			Addr: addr4,
		}
	} else if addr.IP.To16() != nil {
		// IPv6
		var addr16 [16]byte
		copy(addr16[:], addr.IP.To16())
		return &unix.SockaddrInet6{
			Port: addr.Port,
			Addr: addr16,
		}
	}
	return nil
}

func (c *RawPackConn) UnixWriteTo(b []byte, dstaddr net.Addr) (int, error) {
	dstAdrr, ok := dstaddr.(*net.UDPAddr)
	if !ok {
		return 0, errors.New("dstaddr is not a net.UDPAddr")
	}
	packetLen := len(b)
	c.ipHeader.Dst = dstAdrr.IP
	c.ipHeader.ID = codec.GenerateRandomPort()
	c.ipHeader.TotalLen = 20 + 8 + packetLen
	b = codec.BuildUDPPacket(c.srcIP, dstAdrr, b)
	hb, err := c.ipHeader.Marshal()
	if err != nil {
		fmt.Println(err.Error())
	}
	finalyB := append(hb, b...)
	sockAddr := UDPAddrToSockaddr(dstAdrr)
	if sockAddr == nil {
		return 0, errors.New("Failed to convert net.UDPAddr to unix.Sockaddr")
	}

	return packetLen, unix.Sendto(c.fd, finalyB, 0, sockAddr)
}

type Filter []bpf.Instruction

// 设置fd 为非阻塞
func setFDNonblocking(fd int) error {
	if err := unix.SetNonblock(fd, true); err != nil {
		log.Fatalf("setnonblock: %s", err)
	}
	return nil
}

func NewRawConn(addr string, port int, setNonblocking, setEnCrypto bool) *RawPackConn {

	if setEnCrypto {
		en = &encrypt.MyEncrypto{Key: []byte("1234567890abcdef")}
	}

	filter := Filter{
		bpf.LoadAbsolute{Off: 22, Size: 2},                               // load the destination port
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(port), SkipFalse: 1}, // if Val != 8972 skip next instruction
		bpf.RetConstant{Val: 0xffff},                                     // return 0xffff bytes (or less) from packet
		bpf.RetConstant{Val: 0x0}}
	conn, err := net.ListenPacket("ip4:udp", addr)
	if err != nil {
		panic(err)
	}
	cc := conn.(*net.IPConn)
	cc.SetReadBuffer(2 * 1024 * 1024)
	cc.SetWriteBuffer(2 * 1024 * 1024)
	pconn, _ := ipv4.NewRawConn(conn)

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.SOCK_DGRAM)
	if err != nil {
		fmt.Println("create fd is ")
		panic(err)
	}
	// 设置接收缓冲区大小为 4096 字节
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, 2*1024*1024); err != nil {
		log.Fatalf("Setting receive buffer error: %s", err)
	}

	// 设置发送缓冲区大小为 4096 字节
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, 2*1024*1024); err != nil {
		log.Fatalf("Setting send buffer error: %s", err)
	}

	if setNonblocking {
		fdInt := int(fd)
		if err := setFDNonblocking(fdInt); err != nil {
			log.Fatalf("setNonblocking failed: %v", err)
		}
	}
	var assembled []bpf.RawInstruction
	if assembled, err = bpf.Assemble(filter); err != nil {
		log.Print(err)
		return nil
	}
	port1 := strconv.Itoa(port)
	srcIPandPort, err := net.ResolveUDPAddr("udp", addr+":"+port1)
	if err != nil {
		fmt.Println(err.Error())
		panic("net.ResolveUDPAddr(addr+port1)")
	}
	pconn.SetBPF(assembled)
	return &RawPackConn{pconn, &ipv4.Header{
		Version:  4,
		Len:      20,
		Flags:    ipv4.DontFragment,
		TTL:      64,
		Protocol: 17,
		Src:      net.ParseIP(addr),
	}, srcIPandPort, int(fd), make([]byte, 2), setEnCrypto}
}

type PacketConn interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
	WriteTo(p []byte, addr net.Addr) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}
