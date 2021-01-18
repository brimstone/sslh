package sslh

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gliderlabs/ssh"
)

type TCPHandler interface {
	Handle(net.Conn)
}

type Listener struct {
	HTTP  *http.Server
	HTTPS *http.Server
	Raw   TCPHandler
	SSH   *ssh.Server
}

func (l *Listener) Listen(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go l.handleConnection(conn)
	}
}

func (l *Listener) handleConnection(conn net.Conn) error {
	defer conn.Close()

	buffConn := newBufferedConn(conn)
	peekChan := make(chan []byte)
	errChan := make(chan error)
	go func() {
		hello, err := buffConn.Peek(1)
		if err != nil {
			errChan <- err
			return
		}
		peekChan <- hello
	}()

	var clientBanner []byte
	// Take a look at the first 4 bytes
	select {
	case clientBanner = <-peekChan:
	case <-time.After(time.Second):
	case err := <-errChan:
		return err
	}

	sln := &singleListener{
		conn: buffConn,
	}

	switch string(clientBanner) {
	case "GET ":
		if l.HTTP == nil {
			return errors.New("No handler for HTTP")
		}
		l.HTTP.Serve(sln)
		return nil
	case "\x16\x03\x01\x02":
	case "\x16":
		if l.HTTPS == nil {
			return errors.New("No handler for HTTPS")
		}
		l.HTTPS.ServeTLS(sln, "", "")
		return nil
	case "SSH-":
		if l.SSH == nil {
			return errors.New("No handler for SSH")
		}
		l.SSH.Serve(sln)
		return nil
	case "":
		if l.Raw == nil {
			return errors.New("No handler for Raw")
		}
		l.Raw.Handle(buffConn)
		return nil
	default:
		return fmt.Errorf("Unknown banner: %q", clientBanner)
	}
	return nil
}
