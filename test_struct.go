package main

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/lbsystem/my-go-mod/udp/codec"
)

func main() {
	src := net.ParseIP("192.168.1.2")
	dst := net.ParseIP("192.168.1.22")
	ip, err := codec.NewIPHeaderPtr(src, dst)
	if err != nil {
		fmt.Println(err.Error())
	}
	fullPacket := codec.FullPacket{}

	udp := codec.NewUDPHeaderPtr(12345, 12345)
	fullPacket.IP = ip
	fullPacket.UDP = udp

	b := bytes.Repeat([]byte("a"), 1450)
	fullPacket.AddPayload(b)
	fullPacket.AddPayload([]byte{1})
	fmt.Println(fullPacket.UDP.Checksum)
	now := time.Now()

	fmt.Println(udp.Checksum)
	fmt.Printf("time.Since(now).Milliseconds(): %v\n", time.Since(now).Milliseconds())

}
