package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

const Success uint8 = iota
const (
	Socks5 = uint8(5)
	NoAuth = uint8(0)
)

type CNX struct {
	DOMAIN string
	IP     net.IP
	Port   int
}

type Socks struct {
	version         string
	securityMetohds string
}

func (s *Socks) header(bufConn *bufio.Reader) (uint8, error) {
	/*
		+----+----------+----------+
		|VER | NMETHODS | METHODS  |
		+----+----------+----------+
		| 1  |    1     | 1 to 255 |
		+----+----------+----------+

		The VER field is set to X'05' for this version of the protocol.  The
		NMETHODS field contains the number of method identifier octets that
		appear in the METHODS field.
	*/

	// socks version
	vBuffer := []byte{0}
	if _, err := bufConn.Read(vBuffer); err != nil {
		log.Printf("failed to read socks version: %q", err)
		return 0, err
	}

	// num methods
	nmBuffer := []byte{0}
	bufConn.Read(nmBuffer)
	if _, err := bufConn.Read(nmBuffer); err != nil {
		log.Printf("failed to read num security methods byte: %q", err)
		return 0, err
	}

	// methods
	numMethods := int(nmBuffer[0])
	mBuffer := make([]byte, int(numMethods))
	if _, err := io.ReadAtLeast(bufConn, mBuffer, numMethods); err != nil {
		log.Printf("failed to read num security methods byte: %q", err)
		return 0, err
	}

	return vBuffer[0], nil
}

func (s *Socks) sendAuthMethods(conn net.Conn) {
	/*
		+----+--------+
		|VER | METHOD |
		+----+--------+
		| 1  |   1    |
		+----+--------+

		If the selected METHOD is X'FF', none of the methods listed by the
		client are acceptable, and the client MUST close the connection.

		The values currently defined for METHOD are:

			o  X'00' NO AUTHENTICATION REQUIRED
			o  X'01' GSSAPI
			o  X'02' USERNAME/PASSWORD
			o  X'03' to X'7F' IANA ASSIGNED
			o  X'80' to X'FE' RESERVED FOR PRIVATE METHODS
			o  X'FF' NO ACCEPTABLE METHODS
	*/
	conn.Write([]byte{5, 0})
}

func (s *Socks) requestDetails(bufConn *bufio.Reader) (string, int, error) {
	/*
		Once the method-dependent subnegotiation has completed, the client

		sends the request details.  If the negotiated method includes
		encapsulation for purposes of integrity checking and/or
		confidentiality, these requests MUST be encapsulated in the method-
		dependent encapsulation.

		The SOCKS request is formed as follows:
			+----+-----+-------+------+----------+----------+
			|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
			+----+-----+-------+------+----------+----------+
			| 1  |  1  | X'00' |  1   | Variable |    2     |
			+----+-----+-------+------+----------+----------+

			Where:
				o  VER    protocol version: X'05'
				o  CMD
					o  CONNECT X'01'
					o  BIND X'02'
					o  UDP ASSOCIATE X'03'
				o  RSV    RESERVED
				o  ATYP   address type of following address
					o  IP V4 address: X'01'
					o  DOMAINNAME: X'03'
					o  IP V6 address: X'04'
				o  DST.ADDR       desired destination address
				o  DST.PORT desired destination port in network octet
					order
	*/

	// Read the header byte request (ignored)
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(bufConn, header, 3); err != nil {
		log.Printf("failed to read the header bytes: %q", err)
		return "", 0, err
	}

	// Get the address type
	addrType := []byte{0}
	if _, err := io.ReadAtLeast(bufConn, addrType, len(addrType)); err != nil {
		log.Printf("failed to read addr type byte: %q", err)
		return "", 0, err
	}

	// TODO: manage domain name and ip addr only
	if addrType[0] != uint8(3) {
		log.Printf("Unsupported address type: %v", addrType)
		return "", 0, nil
	}

	// Read dest address
	if _, err := bufConn.Read(addrType); err != nil {
		log.Printf("failed to read addr type byte buffer: %q", err)
	}

	addrLen := int(addrType[0])
	domainName := make([]byte, addrLen)
	if _, err := io.ReadAtLeast(bufConn, domainName, addrLen); err != nil {
		log.Printf("failed to get domain name: %q", err)
		return "", 0, err
	}

	fmt.Printf("domain name: %v \v", string(domainName))
	ip, err := net.ResolveIPAddr("ip", string(domainName))
	if err != nil {
		log.Printf("failed to resolve domain name: %q", err)
		return "", 0, err
	}

	// get port
	portBuffer := []byte{0, 0}
	if _, err := io.ReadAtLeast(bufConn, portBuffer, 2); err != nil {
		log.Printf("failed to port: %q", err)
		return "", 0, err
	}
	port := (int(portBuffer[0]) << 8) | int(portBuffer[1])
	ips := fmt.Sprintf("%v", ip)

	return ips, port, nil
}

func (s *Socks) Run(conn *ssh.Client, local string) error {
	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}
	for {
		socksServer, err := listener.Accept()
		if err != nil {
			return err
		}

		// Main reader
		bufConn := bufio.NewReader(socksServer)

		// Get version and method data
		version, err := s.header(bufConn)
		if err != nil {
			continue
		}

		if version != Socks5 {
			continue
		}

		// Send auth methods and version
		s.sendAuthMethods(socksServer)

		// Get target ip:port
		ip, port, err := s.requestDetails(bufConn)
		if err != nil {
			continue
		}

		there, err := conn.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%v", port)))
		if err != nil {
			log.Printf("failed to dial to remote: %q", err)
			continue
		}

		addrTypea := uint8(3) // domain name, ip ->
		addrBody := append([]byte{byte(len(ip))}, ip...)
		addrPort := uint16(port)

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
		socksServer.Write(msg)

		pipe := func(writer, reader net.Conn) {
			defer writer.Close()
			defer reader.Close()
			_, err := io.Copy(writer, reader)
			if err != nil {
				log.Printf("failed to copy: %s", err)
			}
		}
		go pipe(socksServer, there)
		go pipe(there, socksServer)
	}
}
