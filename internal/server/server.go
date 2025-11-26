package server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Closing  atomic.Bool
	Wg       sync.WaitGroup
}

func Serve(port int) (*Server, error) {
	lsn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("Error while creating server: %w", err)
	}

	server := &Server{
		Listener: lsn,
		Closing:  atomic.Bool{},
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
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello World!\n"
	conn.Write([]byte(response))
}
