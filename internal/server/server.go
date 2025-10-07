package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/taham8875/http-from-tcp/internal/response"
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
	defer conn.Close()

	// return the default empty body response
	_ = response.WriteStatusLine(conn, response.StatusOK)
	h := response.GetDefaultHeaders(0)
	_ = response.WriteHeaders(conn, h)
	_, _ = conn.Write([]byte("\r\n"))
}
