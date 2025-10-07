package response

import (
	"fmt"
	"io"

	"github.com/taham8875/http-from-tcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type WriterState int

const (
	StateInitial WriterState = iota
	StateStatusLine
	StateHeaders
	StateBody
)

type Writer struct {
	conn  io.Writer
	state WriterState
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		conn:  conn,
		state: StateInitial,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StateInitial {
		return fmt.Errorf("invalid state: expected StateInitial, got %d", w.state)
	}

	var reason string
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}

	_, err := io.WriteString(w.conn, fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason))

	if err != nil {
		return err
	}

	w.state = StateStatusLine

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Content-Type", "text/plain")
	h.Set("Connection", "close")
	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != StateStatusLine {
		return fmt.Errorf("invalid state: expected StateStatusLine, got %d", w.state)
	}

	b := []byte{}
	for key, value := range h {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	}
	// append the rb
	b = append(b, []byte("\r\n")...)
	_, err := w.conn.Write(b)
	if err != nil {
		return err
	}

	w.state = StateHeaders

	return nil
}

func (w *Writer) WriteBody(body []byte) error {
	if w.state != StateHeaders {
		return fmt.Errorf("invalid state: expected StateHeaders, got %d", w.state)
	}

	_, err := w.conn.Write(body)
	if err != nil {
		return err
	}

	w.state = StateBody

	return nil
}
