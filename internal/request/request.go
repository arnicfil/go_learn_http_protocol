package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type RequestState int

const (
	Initialized RequestState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	State       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, 0, 8)
	buf := make([]byte, 8)

	var req Request
	req.State = Initialized

	flag := false
	for {
		numBytesRead, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				flag = true
			} else {
				return nil, fmt.Errorf("Error while reading from reader: %v", err)
			}
		}

		if len(buffer)+numBytesRead > cap(buffer) {
			newCap := max(cap(buffer)*2, len(buffer)+numBytesRead)

			newBuffer := make([]byte, len(buffer), newCap)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		buffer = append(buffer, buf[:numBytesRead]...)

		bytesParsed, err := req.parse(buffer)
		if err != nil {
			return nil, fmt.Errorf("Error while parsing bytes: %w", err)
		}

		if req.State == Done || flag {
			break
		}

		if bytesParsed == 0 {
			continue
		}

	}

	if flag && req.State != Done {
		return &Request{}, errors.New("Error EOF before a request line")
	}

	return &req, nil
}

func parseRequestLine(data string) (RequestLine, int, error) {
	var line string
	if !strings.Contains(data, "\r\n") {
		return RequestLine{}, 0, nil
	} else {
		line = strings.Split(data, "\r\n")[0]
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("Request line doesn't have 3 parts")
	}

	for _, c := range parts[0] {
		if !unicode.IsUpper(c) || !unicode.IsLetter(c) {
			return RequestLine{}, 0, errors.New("Method contains invalid characters")
		}
	}

	version_parts := strings.Split(parts[2], "/")
	if len(version_parts) != 2 || version_parts[1] != "1.1" {
		return RequestLine{}, 0, errors.New("Version part is invalid")
	}

	return RequestLine{
		HttpVersion:   version_parts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}, len(line) + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case Initialized:
		newRequestLine, numBytesParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if numBytesParsed == 0 {
			return 0, nil
		}
		if numBytesParsed > 0 {
			r.RequestLine = newRequestLine
			r.State = Done
			return numBytesParsed, nil
		}
	case Done:
		return 0, errors.New("Error trying to read data in a done state")
	}

	return 0, errors.New("Error unknown state")
}
