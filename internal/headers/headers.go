package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const rn = "\r\n"

var ERROR_INVALID_HEADER_FORMAT = fmt.Errorf("invalid header format")

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// check for rn
	rnIndex := bytes.Index(data, []byte(rn))

	// if no rn fount, we don't have enough data
	if rnIndex == -1 {
		return 0, false, nil
	}

	// if rn is at the start, we are done
	if rnIndex == 0 {
		return 2, true, nil // 2 = len(rn)
	}

	// get the header headerLine without the rn
	headerLine := data[:rnIndex]

	// find the colon separator
	colonIndex := bytes.IndexByte(headerLine, ':')

	// okay if there is no colon, it is an error
	if colonIndex == -1 {
		return 0, false, ERROR_INVALID_HEADER_FORMAT
	}

	// split the header into key and value
	keyBytes := headerLine[:colonIndex]
	valueBytes := headerLine[colonIndex+1:]

	// check if the key have leading or trailing spaces, this is error if it has
	// keys cannot have leading or trailing spaces, but values can
	if bytes.HasPrefix(keyBytes, []byte(" ")) || bytes.HasSuffix(keyBytes, []byte(" ")) {
		return 0, false, ERROR_INVALID_HEADER_FORMAT
	}

	key := string(keyBytes)
	value := strings.TrimSpace(string(valueBytes))

	// set the header
	h[key] = value

	return rnIndex + 2, false, nil // +2 for the rn

}
