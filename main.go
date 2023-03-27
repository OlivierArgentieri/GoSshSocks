package main

import (
	"fmt"
	"log"

	"github.com/OlivierArgentieri/ligornete/proxy"
	"github.com/spf13/cobra"
)

var Host string
var User string
var Password string
var LocalProxyPort string

var rootCmd = &cobra.Command{
	Use:   "ligornette",
	Short: "Ligornette is used to create an SSH tunnel using a local socks5 Proxy server",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&Host, "sshhost", "H", "", "ssh host (required)")
	rootCmd.Flags().StringVarP(&User, "sshport", "U", "", "ssh user (required)")
	rootCmd.Flags().StringVarP(&Password, "sshpassword", "P", "", "ssh password (required)")
	rootCmd.Flags().StringVarP(&LocalProxyPort, "port", "T", "9090", "local proxy port")
}

func main() {
	proxyAddr := "localhost"
	rootCmd.Execute()
	a := proxy.Proxy{}

	s := proxy.Socks{}

	conn, err := a.InitSSH(Host, User, Password)
	if err != nil {
		log.Fatalf("failed to tunnel init ssh conneciton : %q", err)
	}

	s.Run(conn, fmt.Sprintf("%v:%v", proxyAddr, LocalProxyPort))
}
