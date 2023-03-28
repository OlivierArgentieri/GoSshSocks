package cmd

import (
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
