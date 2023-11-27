package trojan

import (
	"fmt"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

func Test() {
	trcfg := MyTrojanCfg{
		ServerAddr:       "kacb97.gkkgrp.xyz",
		ServerPort:       24113,
		ServerSNI:        "0gr4uqmtt8y41hcjsgrzdrc31.ourdvsss.com",
		DisableHTTPCheck: false,
		MuxOpen:          false,
		MuxLimit:         16,
		ServerPassword:   "xxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	mt := NewTrojan(trcfg)
	tcp, err := mt.DialConn(&tunnel.Address{
		DomainName:  "xxxxxxxxxxxxxxxxx",
		AddressType: tunnel.DomainName,
		Port:        33326,
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	tcp.Write([]byte("fasdfaasdfasd"))
	fmt.Printf("tcp.RemoteAddr(): %v\n", tcp.Metadata().Address)
	mt.Udp.WriteWithMetadata([]byte("1111111"), &tunnel.Metadata{
		Address: &tunnel.Address{
			DomainName:  "xxxxxxxxx",
			AddressType: tunnel.DomainName,
			Port:        33326,
		},
	})
}
