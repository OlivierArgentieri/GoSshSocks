package main

import (
	"fmt"
	"log"

	"github.com/OlivierArgentieri/gosshsocks/cmd"
	"github.com/OlivierArgentieri/gosshsocks/proxy"
)

func main() {
	proxyAddr := "localhost"
	cmd.Execute()

	if cmd.Host == "" || cmd.User == "" {
		return
	}
	a := proxy.Proxy{}

	s := proxy.Socks{}

	conn, err := a.InitSSH(cmd.Host, cmd.User, cmd.Password)
	if err != nil {
		log.Fatalf("failed to tunnel init ssh conneciton : %q", err)
	}

	s.Run(conn, fmt.Sprintf("%v:%v", proxyAddr, cmd.LocalProxyPort))
}
