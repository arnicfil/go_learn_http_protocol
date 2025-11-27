package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/arnicfil/go_learn_http_protocol/internal/headers"
)

type StatusCode int

const (
	StatusOK                                   StatusCode = 200
	StatusBadRequest                   StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var ERROR_LEN_MISSMATCH = errors.New("Error writing len mismatch")

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error = nil
	switch statusCode {
	case StatusOK:
		_, err = w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusInternalServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		_, err = w.Write(fmt.Appendf(nil, "HTTP/1.1 %d\r\n", statusCode))
	}

	if err != nil {
		return fmt.Errorf("Error while writing to writer: %w", err)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	hdrs := headers.NewHeaders()
	hdrs.Set("Content-Length", strconv.Itoa(contentLen))
	hdrs.Set("Connection", "close")
	hdrs.Set("Content-Type", "text/plain")

	return hdrs
}

func WriteHeaders(w io.Writer, hdrs headers.Headers) error {
	for headerKey, headerVal := range hdrs {
		data := fmt.Appendf(nil, "%s: %s\r\n", headerKey, headerVal)
		numBytesWritten, err := w.Write(data)
		if err != nil {
			return fmt.Errorf("Error while writing into writer: %w", err)
		}

		if numBytesWritten != len(data) {
			return ERROR_LEN_MISSMATCH
		}
	}

	numBytesWritten, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("Error while writing into writer: %w", err)
	}

	if numBytesWritten != 2 {
		return ERROR_LEN_MISSMATCH
	}

	return nil
}

func WriteBody(w io.Writer, data []byte) error {
	numBytesWritten, err := w.Write(data)
	if err != nil {
		return fmt.Errorf("Error while writing data to writer: %w", err)
	}

	if numBytesWritten != len(data) {
		return ERROR_LEN_MISSMATCH
	}

	return nil
}
