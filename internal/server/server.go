package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/arnicfil/go_learn_http_protocol/internal/request"
	"github.com/arnicfil/go_learn_http_protocol/internal/response"
)

type Server struct {
	Listener    net.Listener
	Closing     atomic.Bool
	Wg          sync.WaitGroup
	HandlerFunc Handler
}

type Handler func(w *response.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    bytes.Buffer
}

var ERROR_WRITER = errors.New("Error write didn't accept whole message")

func Serve(port int, handlerFunc Handler) (*Server, error) {
	lsn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("Error while creating server: %w", err)
	}

	server := &Server{
		Listener:    lsn,
		Closing:     atomic.Bool{},
		HandlerFunc: handlerFunc,
	}

	server.Closing.Store(false)
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.Closing.Store(true)
	err := s.Listener.Close()
	s.Wg.Wait()
	return err
}

func (s *Server) listen() error {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Closing.Load() {
				return nil
			}
			return fmt.Errorf("Error while accepting: %w", err)
		}

		s.Wg.Add(1)
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer s.Wg.Done()
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("Error while reading from reader: %v", err)
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	herr := s.HandlerFunc(responseWriter, req)
	if herr != nil {
		err = handleError(conn, herr)
		if err != nil {
			fmt.Printf("Error while returning error: %v", err)
			return
		}
	}

}

func writeResponse(w io.Writer, statusCode response.StatusCode, buffer *bytes.Buffer) error {
	responseHeaders := response.GetDefaultHeaders(buffer.Len())

	err := response.WriteStatusLine(w, statusCode)
	if err != nil {
		return fmt.Errorf("Error while writing status line: %w", err)
	}

	err = response.WriteHeaders(w, responseHeaders)
	if err != nil {
		return fmt.Errorf("Error while writing headers: %w", err)
	}

	err = response.WriteBody(w, buffer.Bytes())
	if err != nil {
		return fmt.Errorf("Error while writing body: %w", err)
	}

	return nil
}

func handleError(w io.Writer, herr *HandlerError) error {
	return writeResponse(w, herr.StatusCode, &herr.Message)
}
