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

	
	// go handleConn1(rConn)
	dst, err := net.ResolveUDPAddr("udp", "192.168.9.23:35315")
	if err != nil {
		fmt.Println(err.Error())
	}
	src, err := net.ResolveUDPAddr("udp", "192.168.1.23:35315")
	if err != nil {
		fmt.Println(err.Error())
	}
	rConn := rawConn.NewRawConn( src,false, false)
	fmt.Println("Start")
	i := 0
	b := bytes.Repeat([]byte("f"), 1400)
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
	//dasdads
	now:=time.Now()
	for {
		i++
		if i > 1400000 {
			break
		}
		_, err := rConn.WriteTo(b, dst)
		if err != nil {
			fmt.Println("dfasdfdas", err.Error())
			time.Sleep(time.Millisecond * 10)
			break
		}
	}
	fmt.Println(time.Since(now).Milliseconds())


	// select {}
}
