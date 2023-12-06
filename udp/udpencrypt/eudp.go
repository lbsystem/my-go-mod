package udpencrypt

import (
	"github.com/lbsystem/my-go-mod/encrypt"
	"net"
)

type Addr interface {
	Network() string // name of the network (for example, "tcp", "udp")
	String() string  // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
}

type EncryptUDP struct {
	net.PacketConn
	aesgcm *encrypt.AESGCM
}

func NewEncryptUDP(conn net.PacketConn, key []byte) (*EncryptUDP, error) {
	aesgcm, err := encrypt.NewAESGCM(key)
	if err != nil {
		return nil, err
	}
	return &EncryptUDP{
		PacketConn: conn,
		aesgcm:     aesgcm,
	}, nil
}

func (e *EncryptUDP) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	myN, myAddr, myErr := e.PacketConn.ReadFrom(p)
	if myErr != nil {
		return 0, nil, myErr
	}
	decodeData, err := e.aesgcm.Decrypt(p[:myN])
	if err != nil {
		return 0, nil, err
	}
	copy(p, decodeData)
	return len(decodeData), myAddr, nil
}

func (e *EncryptUDP) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	cipherText, err := e.aesgcm.Encrypt(p)
	if err != nil {
		return 0, err
	}
	return e.PacketConn.WriteTo(cipherText, addr)
}
