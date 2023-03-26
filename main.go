package main

import (
	"fmt"
	"log"

	"github.com/OlivierArgentieri/ligornete/proxy"
)

func main() {
	host := "192.168.1.27:22"
	user := "debian"
	pwd := "toor"

	proxyAddr := "localhost"
	proxyPort := "9091"

	a := proxy.Proxy{}

	s := proxy.Socks{}

	conn, err := a.InitSSH(host, user, pwd)
	if err != nil {
		log.Fatalf("failed to tunnel init ssh conneciton : %q", err)
	}

	s.Run(conn, fmt.Sprintf("%v:%v", proxyAddr, proxyPort))
}
