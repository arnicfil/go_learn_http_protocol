package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arnicfil/go_learn_http_protocol/internal/request"
	"github.com/arnicfil/go_learn_http_protocol/internal/response"
	"github.com/arnicfil/go_learn_http_protocol/internal/server"
)

const port = 42069

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	server, err := server.Serve(port, handler)
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

func handler(w *response.Writer, req *request.Request) *server.HandlerError {
	body := respond200()
	hdrs := response.GetDefaultHeaders(0)
	status := response.StatusOK
	hdrs.Set("Content-Type", "text/html")

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body = respond400()
	case "/myproblem":
		body = respond500()
	}

	hdrs.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	w.WriteStatusLine(status)
	w.WriteHeaders(hdrs)
	w.WriteBody(body)

	return nil
}
