package main

import (
	"fmt"

	"github.com/lbsystem/my-go-mod/gfw"
)

func main() {
	fmt.Printf("ok")
	cd := gfw.NewDomains("https://cdn.jsdelivr.net/gh/Loyalsoldier/v2ray-rules-dat@release/gfw.txt")
	fmt.Printf("cd.IsSubdomainOfAny(\"www.youtube.com\"): %v\n", cd.IsSubdomainOfAny("www.qq.com"))
}
