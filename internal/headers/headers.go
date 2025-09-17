package headers

import (
	"bytes"
	"fmt"
)


type Headers map[string]string

var rn = []byte("\r\n")

func NewHeaders() Headers {
	return map[string]string{}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	fieldName := parts[0]
	fieldValue := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(fieldName, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	return  string(fieldName), string(fieldValue), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data, rn)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			break
		}

		name, value, err := parseHeader(data[:idx])
		if err != nil {
			return 0, false, err
		}

		read += idx + len(rn)
		h[name] = value
	}

	return read, done, nil
}
