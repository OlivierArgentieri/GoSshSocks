package cmd

import (
	"github.com/spf13/cobra"
)

var Host string
var User string
var Password string
var LocalProxyPort string

var rootCmd = &cobra.Command{
	Use:   "GoSshSocks",
	Short: "GoSshSocks is used to create a local socks5 Proxy server to redirect your connection through an ssh tunnel.",
}

func Execute() error {
	return rootCmd.Execute()
}
