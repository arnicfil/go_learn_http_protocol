package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arnicfil/go_learn_http_protocol/internal/request"
	"github.com/arnicfil/go_learn_http_protocol/internal/response"
	"github.com/arnicfil/go_learn_http_protocol/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, yourProblemHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func yourProblemHandler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    *bytes.NewBufferString("Your problem is not my problem\n"),
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    *bytes.NewBufferString("Woopsie, my bad\n"),
		}
	default:
		w.Write([]byte("All goo frfr\n"))
	}

	return nil
}
