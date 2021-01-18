package sslh

import (
	"io"
	"net"
	"sync"
)

// http://grokbase.com/t/gg/golang-nuts/142k5varhg/go-nuts-net-http-serve-method-that-accepts-an-existing-connection#201402205xcwi4ogh5d76tpxz2bbctt3pi

// A singleListener is a net.Listener that returns a single connection, then
// gives the error io.EOF.
type singleListener struct {
	conn net.Conn
	once sync.Once
}

func (s *singleListener) Accept() (net.Conn, error) {
	var c net.Conn
	s.once.Do(func() {
		c = s.conn
	})
	if c != nil {
		return c, nil
	}
	return nil, io.EOF
}

func (s *singleListener) Close() error {
	s.once.Do(func() {
		s.conn.Close()
	})
	return nil
}

func (s *singleListener) Addr() net.Addr {
	return s.conn.LocalAddr()
}
