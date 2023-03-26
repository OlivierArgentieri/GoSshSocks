package proxy

import (
	"bufio"
	"fmt"
	"io"
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

func (p *Proxy) Tunnel(conn *ssh.Client, local string) error {

	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}
	for {
		here, err := listener.Accept()
		if err != nil {
			return err
		}
		bufConn := bufio.NewReader(here)

		// client ip
		clientIP, _, err := net.SplitHostPort(here.RemoteAddr().String())
		client := net.ParseIP(clientIP)
		fmt.Printf("client: %v \n", client)

		version := []byte{0}
		bufConn.Read(version)
		fmt.Printf("version: %v \n", version)

		if version[0] != Socks5 {
			continue
		}

		sec := []byte{0}
		bufConn.Read(sec)
		fmt.Printf("sec: %v \n", sec)
		numMethods := int(sec[0])
		methods := make([]byte, numMethods)
		io.ReadAtLeast(bufConn, methods, numMethods)
		fmt.Printf("methods : %v \n", methods)

		// send auth method
		here.Write([]byte{5, 0})

		// Read the header byte requestA
		header := []byte{0, 0, 0}
		io.ReadAtLeast(bufConn, header, 3)
		fmt.Printf("header buffer : %v \n", header)

		// Get the address type
		addrType := []byte{0}
		io.ReadAtLeast(bufConn, addrType, len(addrType))
		fmt.Printf("addrtype: %v \v", addrType)

		// Read in the destination address
		if addrType[0] != uint8(3) {
			continue
		}

		bufConn.Read(addrType)
		addrLen := int(addrType[0])
		fqdn := make([]byte, addrLen)
		io.ReadAtLeast(bufConn, fqdn, addrLen)
		fmt.Printf("domain name: %v \v", string(fqdn))
		addr, _ := net.ResolveIPAddr("ip", string(fqdn))
		fmt.Printf("resolved domain name: %v \v", addr.IP)

		// Read in the destination address
		d := &AddrSpec{}
		fmt.Printf("dest: %v\n", fqdn)

		// // port
		port := []byte{0, 0}
		io.ReadAtLeast(bufConn, port, 2)
		fmt.Printf("port: %v \n", port)
		d.Port = (int(port[0]) << 8) | int(port[1])
		// d.IP = net.IP(addr)

		fmt.Printf("aaaa %v\n", d.Port)
		// Read in the destination address

		rp := fmt.Sprintf("%v", d.Port)
		rip := fmt.Sprintf("%v", addr.IP)
		a := net.JoinHostPort(rip, rp)
		fmt.Printf("aaasdad %v\n", addr.IP)
		fmt.Printf("aaasdad %v\n", a)
		there, err := conn.Dial("tcp", a)
		if err != nil {
			log.Printf("failed to dial to remote: %q", err)
			continue
		}

		addrTypea := uint8(3) // domain name, ip ->
		addrBody := append([]byte{byte(len(fqdn))}, fqdn...)
		addrPort := uint16(d.Port)

		// Format the message
		msg := make([]byte, 6+len(addrBody))
		msg[0] = 5
		msg[1] = Success
		msg[2] = 0 // Reserved
		msg[3] = addrTypea
		copy(msg[4:], addrBody)
		msg[4+len(addrBody)] = byte(addrPort >> 8)
		msg[4+len(addrBody)+1] = byte(addrPort & 0xff)
		// Send the message
		here.Write(msg)
		pipe := func(writer, reader net.Conn) {
			defer writer.Close()
			defer reader.Close()
			_, err := io.Copy(writer, reader)
			if err != nil {
				log.Printf("failed to copy: %s", err)
			}
		}
		go pipe(here, there)
		go pipe(there, here)
	}
}
