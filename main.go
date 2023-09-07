package main

import (
	// "bytes"
	"fmt"
	"net"

	// "os"
	"bytes"

	"github.com/lbsystem/my-go-mod/udp/rawConn"
	"time"
)

func main() {

	rConn := rawConn.NewRawConn("192.168.1.23", 35315, false, true)
	// go handleConn1(rConn)
	dst, err := net.ResolveUDPAddr("udp", "192.168.1.23:35315")
	if err != nil {
		fmt.Println(err.Error())
	}
	now := time.Now()
	fmt.Println("Start")
	i := 0
	c := 0
	b := bytes.Repeat([]byte("fdasf454543----2fdgbakuio"), 1)
	e := make([]byte, 280)
	for {
		i++
		if i > 3 {
			break
		}
		_, err := rConn.WriteTo(b, dst)
		if err != nil {
			fmt.Println("dfasdfdas", err.Error())
			time.Sleep(time.Millisecond * 10)
			break

		}
		n, _, err := rConn.ReadFrom(e)
		if err != nil {
			fmt.Println("dfasdfdas", err.Error())
			time.Sleep(time.Millisecond * 10)
			break
		}
		fmt.Println(string(e[:n]))

	}
	rConn.Close()
	defer fmt.Println(c)
	fmt.Println("--------------", time.Since(now).Milliseconds())
	// select {}
}
