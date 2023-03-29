package cmd

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Use serve command, to launch a local socks5 Proxy server to redirect your connection through an ssh tunnel.",
}

func init() {
	serveCmd.Flags().StringVarP(&Host, "sshhost", "H", "", "ssh host (required)")
	serveCmd.Flags().StringVarP(&User, "sshport", "U", "", "ssh user (required)")
	serveCmd.Flags().StringVarP(&Password, "sshpassword", "P", "", "ssh password (required)")
	serveCmd.Flags().StringVarP(&LocalProxyPort, "port", "T", "9090", "local proxy port")
	rootCmd.AddCommand(serveCmd)
}
