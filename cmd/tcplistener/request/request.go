package request

import (
	"bytes"
	"fmt"
	"io"
)

type parseState string
const (
	StateInit parseState = "init"
	StateDone parseState = "done"
	StateError parseState = "error"
)

type RequestLine struct {
	HttpVersion  string
	RequestTarget string
	Method       string
}

type Request struct {
	RequestLine RequestLine
	state parseState
}

func newRequest() *Request {
	return &Request {
		state: StateInit,
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformeed request-line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(line []byte) (*RequestLine, int, error) {
	Len := bytes.Index(line, SEPARATOR)
	if Len == -1 {
		return nil, 0 , nil
	}

	startLine := line[:Len]
	read := Len + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2  || string(httpParts[0]) != "HTTP"  || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method: string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion: string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	 for {
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateDone
		case StateDone:
			break outer
		}
	 }
	 return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// buffer could get over run a header above 1k could do this
	// or even a large body

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		// TODO: what to do here?
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
