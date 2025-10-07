package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/taham8875/http-from-tcp/internal/headers"
)

const (
	StateInit = iota
	StateParsingHeaders
	StateParsingBody
	StateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string
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
				// if we are expecting a body, ensure we got it all
				if request.State == StateParsingBody {
					if v, ok := request.Headers.Get("content-length"); ok && len(v) > 0 {
						if cl, err := strconv.Atoi(v); err == nil && cl >= 0 {
							if len(request.Body) < cl {
								return nil, ERROR_MALFORMED_REQUEST
							}
						} else {
							return nil, ERROR_MALFORMED_REQUEST
						}
					}
				}
				request.State = StateDone
				break
			}

			return nil, err
		}

		readToIndex += n

		bytesConsumed, err := request.parse(buf[:readToIndex])

		if err != nil {
			return nil, err
		}

		if bytesConsumed > 0 {
			copy(buf, buf[bytesConsumed:])
			readToIndex -= bytesConsumed
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}

		if n == 0 {
			// need more data
			return totalBytesParsed, nil
		}

		totalBytesParsed += n

	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case StateInit:
		requestLine, bytesConsumed, err := parseRequestLines(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = StateParsingHeaders

		return bytesConsumed, nil
	case StateParsingHeaders:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			// need more date
			return 0, nil
		}

		if done {
			r.State = StateParsingBody
		}

		return n, nil

	case StateParsingBody:
		contentLengthValue, ok := r.Headers.Get("content-length")

		// if no content-length, assume no body and finish
		if !ok || len(contentLengthValue) == 0 || contentLengthValue == "0" {
			r.State = StateDone
			return 0, nil
		}

		// convert contentLength to int from contentLengthValue
		contentLength, err := strconv.Atoi(contentLengthValue)
		if err != nil || contentLength < 0 {
			return 0, ERROR_MALFORMED_REQUEST
		}

		r.Body += string(data)

		if len(r.Body) > contentLength {
			return len(data), ERROR_MALFORMED_REQUEST
		}

		if len(r.Body) == contentLength {
			r.State = StateDone
			return len(data), nil
		}

		return len(data), nil

	case StateDone:
		return 0, ERROR_PARSE_WHEN_DONE
	default:
		return 0, errors.New("invalid state")
	}
}

func parseRequestLines(requestBytes []byte) (*RequestLine, int, error) {
	rnIdx := bytes.Index(requestBytes, []byte(SEPARATOR))
	if rnIdx == -1 {
		// need more data for a full request line
		return nil, 0, nil
	}

	line := requestBytes[:rnIdx]
	parts := bytes.Split(line, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST
	}

	httpVersion := parts[2]
	if string(httpVersion) != "HTTP/1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}

	httpVersion = bytes.TrimPrefix(httpVersion, []byte("HTTP/"))

	return &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpVersion),
	}, rnIdx + len(SEPARATOR), nil
}
