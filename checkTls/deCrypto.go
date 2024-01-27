package checkTls

import (
	"fmt"
	"net"
	"regexp"
)

func DeToCrypto(tls, conn net.Conn) {
	fmt.Println("start DeCrpto")
	b := make([]byte, 8*1024)
	n, err := conn.Read(b)
	re := regexp.MustCompile(`gzip,|deflate,|br`)
	replace := re.ReplaceAllString(string(b[:n]), "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("--------------", replace)
	_, err = tls.Write([]byte(replace))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		n, err := conn.Read(b)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if _, err = tls.Write(b[:n]); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func DeReCrypto(tls, conn net.Conn) {
	fmt.Println("start DeCrpto")
	for {
		b := make([]byte, 8*1024)
		read, err := tls.Read(b)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(b[:read]))
		_, err = conn.Write(b[:read])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
