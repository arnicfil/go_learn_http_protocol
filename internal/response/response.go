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
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type StatusWriter int

const (
	WritingStatusLine StatusWriter = iota
	WritingHeaders
	WritingBody
	WritingTrailers
)

type Writer struct {
	writer        io.Writer
	writingStatus StatusWriter
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer:        w,
		writingStatus: WritingStatusLine,
	}
}

const chunkSize = 10

var ERROR_LEN_MISSMATCH = errors.New("Error writing len mismatch")
var ERROR_WRITING_MISMATCH = errors.New("Error writing response in bad order")

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte("HTTP/1.1 ")
	switch statusCode {
	case StatusOK:
		statusLine = fmt.Appendf(statusLine, "%d OK\r\n", statusCode)
	case StatusBadRequest:
		statusLine = fmt.Appendf(statusLine, "%d Bad Request\r\n", statusCode)
	case StatusInternalServerError:
		statusLine = fmt.Appendf(statusLine, "%d Internal Server Error\r\n", statusCode)
	default:
		statusLine = fmt.Appendf(nil, "HTTP/1.1 %d\r\n", statusCode)
	}

	_, err := w.Write(statusLine)
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
	headersData := []byte{}
	for headerKey, headerVal := range hdrs {
		headersData = fmt.Appendf(headersData, "%s: %s\r\n", headerKey, headerVal)
	}
	headersData = fmt.Append(headersData, "\r\n")

	numBytesWritten, err := w.Write(headersData)
	if err != nil {
		return fmt.Errorf("Error while writing into writer: %w", err)
	}

	if numBytesWritten != len(headersData) {
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

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writingStatus != WritingStatusLine {
		return ERROR_WRITING_MISMATCH
	}

	statusLine := []byte("HTTP/1.1 ")
	switch statusCode {
	case StatusOK:
		statusLine = fmt.Appendf(statusLine, "%d OK\r\n", statusCode)
	case StatusBadRequest:
		statusLine = fmt.Appendf(statusLine, "%d Bad Request\r\n", statusCode)
	case StatusInternalServerError:
		statusLine = fmt.Appendf(statusLine, "%d Internal Server Error\r\n", statusCode)
	default:
		statusLine = fmt.Appendf(nil, "HTTP/1.1 %d\r\n", statusCode)
	}

	w.writingStatus = WritingHeaders
	_, err := w.writer.Write(statusLine)
	if err != nil {
		return fmt.Errorf("Error while writing statusLine: %w", err)
	}
	return nil
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writingStatus != WritingHeaders {
		return ERROR_WRITING_MISMATCH
	}

	headersData := []byte{}
	for headerKey, headerVal := range headers {
		headersData = fmt.Appendf(headersData, "%s: %s\r\n", headerKey, headerVal)
	}
	headersData = fmt.Append(headersData, "\r\n")

	w.writingStatus = WritingBody
	numBytesWritten, err := w.writer.Write(headersData)
	if err != nil {
		return fmt.Errorf("Error while writing into writer: %w", err)
	}

	if numBytesWritten != len(headersData) {
		return ERROR_LEN_MISSMATCH
	}

	return nil
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writingStatus != WritingBody {
		return 0, ERROR_WRITING_MISMATCH
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writingStatus != WritingBody {
		return 0, ERROR_WRITING_MISMATCH
	}
	_, err := w.writer.Write(fmt.Appendf(nil, "%X\r\n", len(p)))
	if err != nil {
		return 0, err
	}
	_, err = w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	_, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writingStatus != WritingBody {
		return 0, ERROR_WRITING_MISMATCH
	}

	data := []byte("0\r\n")

	numBytesWritten, err := w.writer.Write(data)
	w.writingStatus = WritingTrailers

	return numBytesWritten, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writingStatus != WritingTrailers {
		return ERROR_WRITING_MISMATCH
	}

	trailersData := []byte{}
	for trailerKey, trailerVal := range h {
		trailersData = fmt.Appendf(trailersData, "%s: %s\r\n", trailerKey, trailerVal)
	}
	trailersData = fmt.Append(trailersData, "\r\n")

	numBytesWritten, err := w.writer.Write(trailersData)
	if err != nil {
		return fmt.Errorf("Error while writing into writer: %w", err)
	}

	if numBytesWritten != len(trailersData) {
		return ERROR_LEN_MISSMATCH
	}

	return nil
}
