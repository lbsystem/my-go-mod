package trojan

import (
	"context"
	"fmt"

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

func (t *MyTrojan) DialConn(target *tunnel.Address) (tunnel.Conn, error) {

	conn1, err := t.TcpDial.DialConn(target, nil)
	if err != nil {
		fmt.Println("dfasdfas", err.Error())
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
	myTrojan.Udp, err = c.DialPacket(nil)
	if !option.MuxOpen {
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

// 	conn1, err := simpleClient.DialConn(&tunnel.Address{
// 		DomainName:  "lbtest.top",
// 		NetworkType: "tcp",
// 		Port:        33326,
// 		AddressType: tunnel.DomainName,
// 	}, nil)
// 	if err != nil {
// 		fmt.Println("dfasdfas", err.Error())
// 	}

// 	common.Must(err)
// 	// bb := make([]byte, 1024)

// 	_, err = conn1.Write([]byte("asdfdasf"))
// 	fmt.Printf("conn1.Metadata().Address: %v\n", conn1.Metadata())
// 	common.Must(err)
// 	// 发送UDP

// 	pc, err := c.DialPacket(nil)
// 	common.Must(err)
// 	pc.WriteWithMetadata([]byte("1111111"), &tunnel.Metadata{
// 		Address: &tunnel.Address{
// 			DomainName:  "lbtest.top",
// 			AddressType: tunnel.DomainName,
// 			Port:        33326,
// 		},
// 	})
// 	common.Must(err)
// 	// fmt.Printf("n: %v\n", n)

// 	// redirecting

// 	conn1.Close()

// 	c.Close()

// 	cancel()
// }
