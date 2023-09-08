package main

import (
	// "bytes"
	"fmt"
	"net"

	// "os"
	"bytes"

	"time"

	"github.com/lbsystem/my-go-mod/udp/rawConn"
)

func main() {

	rConn := rawConn.NewRawConn("192.168.1.23", 33311, false, false)
	// go handleConn1(rConn)
	dst, err := net.ResolveUDPAddr("udp", "192.168.9.23:35315")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Start")
	i := 0
	b := bytes.Repeat([]byte("f"), 28)
	go func() {
		k := make([]byte, 1500)
		for {
			n, a, err := rConn.ReadFrom(k)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(string(k[:n]))
			rConn.WriteTo(append(k[:n], []byte(" server reply")...), a)
		}
	}()
	for {
		i++
		if i > 12 {
			break
		}
		_, err := rConn.WriteTo(b, dst)
		if err != nil {
			fmt.Println("dfasdfdas", err.Error())
			time.Sleep(time.Millisecond * 10)
			break
		}
	}
	select {}

	// select {}
}
