package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Error while reading from reader: %v", err)
	}

	first_line := strings.Split(string(data), "\r\n")[0]
	rLine, err := parseRequestLine(first_line)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing request line: %v", err)
	}

	req := Request{
		RequestLine: rLine,
	}
	return &req, nil
}

func parseRequestLine(data string) (RequestLine, error) {
	parts := strings.Split(data, " ")
	if len(parts) != 3 {
		return RequestLine{}, errors.New("Request line doesn't have 3 parts")
	}

	for _, c := range parts[0] {
		if !unicode.IsUpper(c) || !unicode.IsLetter(c) {
			return RequestLine{}, errors.New("Method contains invalid characters")
		}
	}

	version_parts := strings.Split(parts[2], "/")
	if len(version_parts) != 2 || version_parts[1] != "1.1" {
		return RequestLine{}, errors.New("Version part is invalid")
	}

	return RequestLine{
		HttpVersion:   version_parts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}, nil
}
