package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: ln,
	}

	go s.listen()

	return s, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) Close() error {
	if s == nil {
		return nil
	}

	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) handle(conn net.Conn) {
	// TODO: Parse the request properly, we now just send a fixed response.
	// Always respond with a 200 Hello World!
	// Content-Length: 13 to include trailing newline.
	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello World!\n"

	_, _ = conn.Write([]byte(resp))
}
