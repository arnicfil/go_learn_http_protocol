package request

import (
	"errors"
	"fmt"
	"github.com/arnicfil/go_learn_http_protocol/internal/headers"
	"io"
	"strings"
	"unicode"
)

type RequestState int

const (
	Initialized RequestState = iota
	ParsingHeaders
	Done
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	state       RequestState
	Headers     headers.Headers
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)

	var req Request
	req.state = Initialized

	readToIndex := 0

	for req.state != Done {
		if readToIndex == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer[:readToIndex])
			buffer = newBuffer
		}

		numBytesRead, err := reader.Read(buffer[readToIndex:])
		readToIndex += numBytesRead

		numBytesParsed, perr := req.parse(buffer[:readToIndex])
		if perr != nil {
			return &Request{}, fmt.Errorf("Error while parsing data: %v", perr)
		}

		if numBytesParsed > 0 {
			copy(buffer, buffer[numBytesParsed:readToIndex])
			readToIndex -= numBytesParsed
		}

		if err != nil {
			if err == io.EOF && req.state != Done {
				return &Request{}, fmt.Errorf("Error final parsing: %v", err)
			}

			return &Request{}, fmt.Errorf("Error while reading from reader: %v", err)
		}
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
	totalBytesParsed := 0
	for r.state != Done {
		numBytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		if numBytesParsed == 0 {
			return totalBytesParsed, nil
		}

		totalBytesParsed += numBytesParsed
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case Initialized:
		newRequestLine, numBytesParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if numBytesParsed == 0 {
			return 0, nil
		}

		r.RequestLine = newRequestLine
		r.state = ParsingHeaders
		return numBytesParsed, nil
	case ParsingHeaders:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		numBytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if numBytesParsed == 0 {
			return 0, nil
		}

		if done {
			r.state = Done
		}

		return numBytesParsed, nil
	case Done:
		return 0, errors.New("Error trying to read data in a done state")
	}

	return 0, errors.New("Error unknown state")
}
