package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const (
	StateInit = iota
	StateDone
)

type Request struct {
	RequestLine RequestLine
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 1024
const SEPARATOR = "\r\n"

var ERROR_MALFORMED_REQUEST = fmt.Errorf("malformed request")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_PARSE_WHEN_DONE = fmt.Errorf("trying to parse when already done")
var ERROR_INVALID_STATE = fmt.Errorf("invalid state")

func newRequest() *Request {
	return &Request{
		State: StateInit,
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := newRequest()

	for request.State != StateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])

		if err != nil {
			if err == io.EOF {
				request.State = StateDone
				break
			}

			return nil, err
		}

		readToIndex += n

		bytesConsumed, err := request.parse(buf[:readToIndex])

		if bytesConsumed > 0 {
			copy(buf, buf[bytesConsumed:])
			readToIndex -= bytesConsumed
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case StateInit:
		requestLine, bytesConsumed, err := parseRequestLines(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.State = StateDone

		r.RequestLine = *requestLine
		return bytesConsumed, nil
	case StateDone:
		return 0, ERROR_PARSE_WHEN_DONE
	default:
		return 0, errors.New("invalid state")
	}
}

func parseRequestLines(requestBytes []byte) (*RequestLine, int, error) {
	lines := bytes.Split(requestBytes, []byte(SEPARATOR))
	if len(lines) < 1 {
		return nil, 0, ERROR_MALFORMED_REQUEST
	}

	requestLineParts := bytes.Split(lines[0], []byte(" "))
	if len(requestLineParts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST
	}

	httpVersion := requestLineParts[2]

	if string(httpVersion) != "HTTP/1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}

	httpVersion = bytes.TrimPrefix(httpVersion, []byte("HTTP/"))

	return &RequestLine{
		Method:        string(requestLineParts[0]),
		RequestTarget: string(requestLineParts[1]),
		HttpVersion:   string(httpVersion),
	}, len(lines[0]) + len(SEPARATOR), nil
}
