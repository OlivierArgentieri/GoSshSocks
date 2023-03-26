package proxy

import (
	"bufio"
	"io"
	"log"
	"net"
)

const Success uint8 = iota
const (
	Socks5 = uint8(5)
	NoAuth = uint8(0)
)

type Socks struct {
	version         string
	securityMetohds string
}

func (s *Socks) Header(bufConn *bufio.Reader) (uint8, uint8, error) {
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
		return 0, 0, err
	}

	// num methods
	nmBuffer := []byte{0}
	bufConn.Read(nmBuffer)
	if _, err := bufConn.Read(nmBuffer); err != nil {
		log.Printf("failed to num security methods byte: %q", err)
		return 0, 0, err
	}

	// methods
	numMethods := int(nmBuffer[0])
	mBuffer := make([]byte, int(numMethods))
	if _, err := io.ReadAtLeast(bufConn, mBuffer, numMethods); err != nil {
		log.Printf("failed to num security methods byte: %q", err)
		return 0, 0, err
	}

	return vBuffer[0], mBuffer[0], nil
}

func (s *Socks) SendAuthMethods(conn net.Conn) {
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
