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
	udp := codec.NewUDPHeaderPtr(12345, 12345)
	b := bytes.Repeat([]byte("a"), 1450)

	now := time.Now()
	for i := 0; i < 14000000; i++ {
		codec.AddPayload(ip, udp, b)
	}
	fmt.Printf("time.Since(now).Milliseconds(): %v\n", time.Since(now).Milliseconds())

}
