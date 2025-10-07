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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
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

	_, err := io.WriteString(w, fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Content-Type", "text/plain")
	h.Set("Connection", "close")
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	b := []byte{}
	for key, value := range h {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	}
	// append the rb
	b = append(b, []byte("\r\n")...)
	_, err := w.Write(b)
	return err
}
