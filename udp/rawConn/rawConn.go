package rawConn

//data 2
import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
	"crypto/sha512"
	"github.com/lbsystem/my-go-mod/udp/codec"

	"golang.org/x/sys/unix"

	"golang.org/x/net/bpf"
	"golang.org/x/net/ipv4"
)

type RawPackConn struct {
	*ipv4.RawConn
	ipHeader *ipv4.Header
	srcIP    *net.UDPAddr
	fd       int

}
var Key []byte
var 	EnCrypto bool	
func (c *RawPackConn) Close() error {
	return unix.Close(c.fd)
}

func (c *RawPackConn) SetEnCrypto(key []byte){
	h := sha512.New()
	h.Write(key)
	h.Write([]byte("hijoadf|~js*(%)io4!!@#"))
	Key=h.Sum(nil)
	EnCrypto=true
}

func (c *RawPackConn) ReadFrom(b []byte) (int, net.Addr, error) {
	h, newByte, _, err := c.RawConn.ReadFrom(b)
	if err != nil {
		return 0, nil, err
	}
	
	n := copy(b, newByte[8:])
	return n, &net.UDPAddr{
		IP:   h.Src,
		Port: int(binary.BigEndian.Uint16(newByte[0:2])),
	}, err
}

func (c *RawPackConn) WriteTo(b []byte, dstaddr net.Addr) (int, error) {
	dstAdrr, ok := dstaddr.(*net.UDPAddr)
	if !ok {
		return 0, errors.New("dstaddr is not a net.UDPAddr")
	}
	packetLen := len(b)
	c.ipHeader.Dst = dstAdrr.IP
	c.ipHeader.ID = codec.GenerateRandomPort()
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
	hb, _ := c.ipHeader.Marshal()
	finalyB := append(hb, b...)
	sockAddr := UDPAddrToSockaddr(dstAdrr)
	if sockAddr == nil {
		return 0, errors.New("Failed to convert net.UDPAddr to unix.Sockaddr")
	}

	return packetLen, unix.Sendto(c.fd, finalyB, 0, sockAddr)
}

func (c *RawPackConn) UnixWriteToEnCryto(b []byte, dstaddr net.Addr) (int, error) {
	dstAdrr, ok := dstaddr.(*net.UDPAddr)
	if !ok {
		return 0, errors.New("dstaddr is not a net.UDPAddr")
	}
	packetLen := len(b)
	c.ipHeader.Dst = dstAdrr.IP
	c.ipHeader.ID = codec.GenerateRandomPort()
	c.ipHeader.TotalLen = 20 + 8 + packetLen
	b=XorCipher(b)
	b = codec.BuildUDPPacket(c.srcIP, dstAdrr, b)
	hb, _ := c.ipHeader.Marshal()
	
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

func NewRawConn(addr string, port int, setNonblocking bool) *RawPackConn {
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
	}, srcIPandPort, int(fd)}
}

func XorCipher(data []byte) []byte {
	encrypted := make([]byte, len(data))
	keyLen := len(Key)
	for i := range data {
		encrypted[i] = data[i] ^ Key[i%keyLen]
	}
	return encrypted
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
