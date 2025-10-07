package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/taham8875/http-from-tcp/internal/request"
	"github.com/taham8875/http-from-tcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

func writeHandlerError(w io.Writer, handlerError *HandlerError) error {
	if handlerError == nil {
		return nil
	}

	if err := response.WriteStatusLine(w, handlerError.Code); err != nil {
		return err
	}

	body := []byte(handlerError.Message)

	headers := response.GetDefaultHeaders(len(body))

	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}

	if _, err := w.Write(body); err != nil {
		return err
	}

	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)

	if err != nil {
		_ = writeHandlerError(conn, &HandlerError{
			Code:    response.StatusBadRequest,
			Message: "Bad Request\n",
		})
	}

	// prepare buffer for handler body

	var buf bytes.Buffer

	if s.handler != nil {
		if handlerError := s.handler(&buf, req); handlerError != nil {
			_ = writeHandlerError(conn, handlerError)
			return
		}
	}

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		return
	}

	h := response.GetDefaultHeaders(buf.Len())
	if err := response.WriteHeaders(conn, h); err != nil {
		return
	}

	_, _ = io.Copy(conn, &buf)

}
