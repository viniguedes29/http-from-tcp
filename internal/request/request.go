package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ERROR_MALIFORMED_REQUEST = "malformed request"
var ERROR_UNSUPPORTED_HTTP_VERSION = "unsupported http version"

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestString := string(data)
	requestLines, err := parseRequestLines(requestString)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLines,
	}, nil
}

func parseRequestLines(requestString string) (*RequestLine, error) {
	lines := strings.Split(requestString, "\r\n")
	if len(lines) < 1 {
		return nil, errors.New(ERROR_MALIFORMED_REQUEST)
	}

	requestLineParts := strings.Split(lines[0], " ")
	if len(requestLineParts) != 3 {
		return nil, errors.New(ERROR_MALIFORMED_REQUEST)
	}

	method := requestLineParts[0]
	requestTarget := requestLineParts[1]
	httpVersion := requestLineParts[2]

	if httpVersion != "HTTP/1.1" {
		return nil, errors.New(ERROR_UNSUPPORTED_HTTP_VERSION)
	}

	httpVersion = strings.TrimPrefix(httpVersion, "HTTP/")

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion,
	}, nil
}
