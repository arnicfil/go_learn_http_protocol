package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/arnicfil/go_learn_http_protocol/internal/headers"
)

type StatusCode int

const (
	StatusOK StatusCode = iota
	StatusBadRequest
	StatusInternalServerError
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error = nil
	switch statusCode {
	case StatusOK:
		_, err = w.Write([]byte("HTTP/1.1 200 OK"))
	case StatusBadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request"))
	case StatusInternalServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error"))
	default:
		_, err = w.Write(fmt.Appendf(nil, "HTTP/1.1 %d", statusCode))
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
