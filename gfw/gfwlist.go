package gfw

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func New_gfw_list(url string) []*net.IPNet {
	r, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
	}
	s := string(b)
	gfw_list := strings.Split(s, "\n")
	return praseList_to_cidrs(gfw_list)
}

func praseList_to_cidrs(IPs []string) []*net.IPNet {
	cidrs := make([]*net.IPNet, 0, 35000)
	for _, ip := range IPs {

		if !strings.Contains(ip, "/") {

			ip += "/32"

		}
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			log.Println(err.Error())
		}
		cidrs = append(cidrs, ipnet)
	}
	return cidrs
}
