package proxy

import (
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

type Proxy struct{}

func (p *Proxy) InitSSH(host, user, password string) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	clientConn, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		log.Fatalf("failed to connect to the ssh server: %q", err)
	}
	return clientConn, nil
}
