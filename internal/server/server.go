package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/taham8875/http-from-tcp/internal/request"
	"github.com/taham8875/http-from-tcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, h Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: ln,
		handler:  h,
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

	req, err := request.RequestFromReader(conn)

	if err != nil {
		w := response.NewWriter(conn)
		w.WriteStatusLine(response.StatusBadRequest)

		errorBody := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

		headers := response.GetDefaultHeaders(0)
		headers.SetOverride("Content-Type", "text/html")
		headers.SetOverride("Content-Length", fmt.Sprintf("%d", len(errorBody)))
		w.WriteHeaders(headers)

		w.WriteBody(errorBody)
		return
	}

	if s.handler != nil {
		w := response.NewWriter(conn)
		s.handler(w, req)
	}
}
