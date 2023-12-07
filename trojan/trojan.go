package trojan

import (
	"context"
	"fmt"
	"net"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
	"github.com/p4gefau1t/trojan-go/tunnel/mux"
	"github.com/p4gefau1t/trojan-go/tunnel/simplesocks"
	"github.com/p4gefau1t/trojan-go/tunnel/tls"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
)

type dialer interface {
	Close() error
	DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error)
	DialPacket(t tunnel.Tunnel) (tunnel.PacketConn, error)
}

type MyTrojan struct {
	Ctx          context.Context
	TransportCfg *transport.Config
	MuxCfg       *mux.Config
	TlsCfg       *tls.Config
	MemCfg       *memory.Config
	TcpDial      dialer
	Udp          tunnel.PacketConn
	Cancel       context.CancelFunc
	TrojanCfg    *trojan.Config
}

type myPacketConn struct {
	tunnel.PacketConn
	addr *tunnel.Metadata
}

func (p *myPacketConn) WriteTo(data []byte, addr *net.UDPAddr) (int, error) {
		
	
	return p.PacketConn.WriteWithMetadata(data, &tunnel.Metadata{
		Address: &tunnel.Address{
			DomainName:  addr.IP.String(),
			AddressType: tunnel.DomainName,
			Port:        addr.Port,
		},
	})
	
}
func (p *myPacketConn) ReadFrom(data []byte) (int, *net.UDPAddr, error) {
	n, a, err := p.PacketConn.ReadFrom(data)
	u, err2 := net.ResolveUDPAddr("udp", a.String())
	if err2 != nil {
		return 0, nil, nil
	}
	return n, u, err
}

//	func (t *MyTrojan) DialPacket() (myPacketConn, error) {
//		pc, err := t.Udp.DialPacket(nil)
//		return myPacketConn{pc}, err
//	}
func (t *MyTrojan) DialConn(target *net.TCPAddr) (tunnel.Conn, error) {

	conn1, err := t.TcpDial.DialConn(&tunnel.Address{
		DomainName:  target.IP.String(),
		NetworkType: "tcp",
		Port:        target.Port,
		AddressType: tunnel.DomainName,
	}, nil)
	if err != nil {
		fmt.Println("DialConn", err.Error())
	}

	return conn1, err

}

type MyTrojanCfg struct {
	ServerAddr       string
	ServerPort       int
	ServerSNI        string
	DisableHTTPCheck bool
	MuxOpen          bool
	MuxLimit         int
	ServerPassword   string
}

func NewTrojan(option MyTrojanCfg) *MyTrojan {
	var myTrojan MyTrojan
	myTrojan.Ctx, myTrojan.Cancel = context.WithCancel(context.Background())
	myTrojan.TransportCfg = &transport.Config{
		RemoteHost: option.ServerAddr,
		RemotePort: option.ServerPort,
	}
	myTrojan.MuxCfg = &mux.Config{
		Mux: mux.MuxConfig{
			Enabled:     option.MuxOpen,
			Concurrency: option.MuxLimit,
			IdleTimeout: 60,
		},
	}
	myTrojan.MemCfg = &memory.Config{Passwords: []string{option.ServerPassword}}
	myTrojan.TlsCfg = &tls.Config{
		RemoteHost: option.ServerAddr,
		RemotePort: option.ServerPort,
		TLS: tls.TLSConfig{
			Verify: false,
			SNI:    option.ServerSNI,
		},
	}
	myTrojan.TrojanCfg = &trojan.Config{
		RemoteHost:       option.ServerAddr,
		RemotePort:       option.ServerPort,
		DisableHTTPCheck: option.DisableHTTPCheck,
	}
	myTrojan.Ctx = config.WithConfig(myTrojan.Ctx, transport.Name, myTrojan.TransportCfg)
	myTrojan.Ctx = config.WithConfig(myTrojan.Ctx, tls.Name, myTrojan.TlsCfg)
	myTrojan.Ctx = config.WithConfig(myTrojan.Ctx, mux.Name, myTrojan.MuxCfg)
	myTrojan.Ctx = config.WithConfig(myTrojan.Ctx, freedom.Name, &freedom.Config{})
	myTrojan.Ctx = config.WithConfig(myTrojan.Ctx, memory.Name, myTrojan.MemCfg)
	myTrojan.Ctx = config.WithConfig(myTrojan.Ctx, trojan.Name, myTrojan.TrojanCfg)
	transportClient, err := transport.NewClient(myTrojan.Ctx, nil)
	if err != nil {
		fmt.Println("init err: ", err.Error())
	}
	tlsClient, err := tls.NewClient(myTrojan.Ctx, transportClient)
	if err != nil {
		fmt.Println("init err: ", err.Error())
	}
	c, err := trojan.NewClient(myTrojan.Ctx, tlsClient)
	if err != nil {
		fmt.Println("init err: ", err.Error())

	}

	if !option.MuxOpen {
		myTrojan.Udp, _ = c.DialPacket(nil)
		myTrojan.TcpDial = c
		return &myTrojan
	}
	muxTunnel := mux.Tunnel{}
	muxClient, err := muxTunnel.NewClient(myTrojan.Ctx, c)
	if err != nil {
		fmt.Println("init err: ", err.Error())
	}
	myTrojan.TcpDial, err = simplesocks.NewClient(myTrojan.Ctx, muxClient)
	if err != nil {
		fmt.Println("init err: ", err.Error())
	}
	myTrojan.Udp, _ = myTrojan.TcpDial.DialPacket(nil)
	return &myTrojan
}

// func runTrojan() {

// 	//port := common.PickPort("tcp", "127.0.0.1")
// 	//fmt.Println("port------------", port)
// 	transportConfig := &transport.Config{
// 		RemoteHost: "shdata1.fc-streaming.xyz",
// 		RemotePort: 10102,
// 	}
// 	muxCfg := &mux.Config{
// 		Mux: mux.MuxConfig{
// 			Enabled:     true,
// 			Concurrency: 16,
// 			IdleTimeout: 60,
// 		},
// 	}

// 	clientCfg1 := &tls.Config{
// 		RemoteHost: "shdata1.fc-streaming.xyz",
// 		RemotePort: 10102,
// 		TLS: tls.TLSConfig{
// 			Verify: false,
// 			SNI:    "oss-cn-hangzhou.aliyuncs.com",
// 		},
// 	}
// 	authConfig := &memory.Config{Passwords: []string{"10681B6D-A7AB-BA83-9D93-D8FA17554CE1"}}
// 	clientConfig := &trojan.Config{
// 		RemoteHost:       "shdata1.fc-streaming.xyz",
// 		RemotePort:       10102,
// 		DisableHTTPCheck: true,
// 	}
// 	ctx, cancel := context.WithCancel(context.Background())
// 	ctx = config.WithConfig(ctx, transport.Name, transportConfig)
// 	ctx = config.WithConfig(ctx, tls.Name, clientCfg1)
// 	ctx = config.WithConfig(ctx, mux.Name, muxCfg)
// 	ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
// 	ctx = config.WithConfig(ctx, memory.Name, authConfig)

// 	transportClient, err := transport.NewClient(ctx, nil)
// 	if err != nil {
// 		fmt.Println("init err: ", err.Error())
// 	}

// 	tlsClient, err := tls.NewClient(ctx, transportClient)
// 	if err != nil {
// 		fmt.Println("init err: ", err.Error())
// 	}

// 	clientCtx := config.WithConfig(ctx, trojan.Name, clientConfig)
// 	c, err := trojan.NewClient(clientCtx, tlsClient)
// 	if err != nil {
// 		fmt.Println("dfasdfas", err.Error())
// 	}
// 	muxTunnel := mux.Tunnel{}
// 	muxClient, err := muxTunnel.NewClient(ctx, c)
// 	if err != nil {
// 		fmt.Println("init err: ", err.Error())
// 	}
// 	fmt.Println("opokoko")
// 	simpleClient, err := simplesocks.NewClient(ctx, muxClient)
// 	if err != nil {
// 		fmt.Println("init err: ", err.Error())
// 	}
// 	//发送TCP 报文

// conn1, err := simpleClient.DialConn(&tunnel.Address{
// 	DomainName:  "lbtest.top",
// 	NetworkType: "tcp",
// 	Port:        33326,
// 	AddressType: tunnel.DomainName,
// }, nil)
// 	if err != nil {
// 		fmt.Println("dfasdfas", err.Error())
// 	}

// 	common.Must(err)
// 	// bb := make([]byte, 1024)

// 	_, err = conn1.Write([]byte("asdfdasf"))
// 	fmt.Printf("conn1.Metadata().Address: %v\n", conn1.Metadata())
// 	common.Must(err)
// 	// 发送UDP

// pc, err := c.DialPacket(nil)
// 	common.Must(err)
// pc.WriteWithMetadata([]byte("1111111"), &tunnel.Metadata{
// 	Address: &tunnel.Address{
// 		DomainName:  "lbtest.top",
// 		AddressType: tunnel.DomainName,
// 		Port:        33326,
// 	},
// })
// 	common.Must(err)
// 	// fmt.Printf("n: %v\n", n)

// 	// redirecting

// 	conn1.Close()

// 	c.Close()

// 	cancel()
// }

func test() {
	trcfg := MyTrojanCfg{
		ServerAddr:       "shdata1.fc-streaming.xyz",
		ServerPort:       10102,
		ServerSNI:        "oss-cn-hangzhou.aliyuncs.com",
		DisableHTTPCheck: false,
		MuxOpen:          true,
		MuxLimit:         16,
		ServerPassword:   "10681B6D-A7AB-BA83-9D93-D8FA17554CE1",
	}
	mt := NewTrojan(trcfg)
	t, _ := net.ResolveTCPAddr("tcp", "lbtest.top:33326")
	tcp, err := mt.DialConn(t)
	if err != nil {
		fmt.Println(err.Error())
	}

	tcp.Write([]byte("fasdfaasdfasd"))
	fmt.Printf("tcp.RemoteAddr(): %v\n", tcp.Metadata().Address)

	// mt.Udp.WriteWithMetadata([]byte("222222"), &tunnel.Metadata{
	// 	Address: &tunnel.Address{
	// 		DomainName:  "lbtest.top",
	// 		AddressType: tunnel.DomainName,
	// 		Port:        33326,
	// 	},
	// })
	uIP, _ := net.ResolveUDPAddr("udp", "lbtest.top:33326")
	mt.Udp.WriteTo([]byte("asdfsdaf"), uIP)
	bb := make([]byte, 1500)
	n, a, err := mt.Udp.ReadFrom(bb)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(a, bb[:n])
	// pc.WriteWithMetadata([]byte("1111111"), &tunnel.Metadata{
	// 	Address: &tunnel.Address{
	// 		DomainName:  "8.210.34.161",
	// 		AddressType: tunnel.DomainName,
	// 		Port:        33326,
	// 	},
	// })
}
