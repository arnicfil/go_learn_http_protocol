package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/arnicfil/go_learn_http_protocol/internal/headers"
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
	target := req.RequestLine.RequestTarget

	if strings.Contains(target, "yourproblem") {
		body = respond400()
		status = response.StatusBadRequest
	} else if strings.Contains(target, "myproblem") {
		body = respond500()
		status = response.StatusInternalServerError
	} else if strings.Contains(target, "httpbin") {
		err := handleChunks(w, req)
		return err
	} else if strings.Contains(target, "video") {
		err := handleVideo(w, req)
		return err
	}

	hdrs.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	w.WriteStatusLine(status)
	w.WriteHeaders(hdrs)
	w.WriteBody(body)

	return nil
}

func handleChunks(w *response.Writer, req *request.Request) *server.HandlerError {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		return nil
	}

	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	target := "https://httpbin.org" + path

	resp, err := http.Get(target)
	if err != nil {
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    *bytes.NewBufferString(err.Error()),
		}
	}
	defer resp.Body.Close()

	hdrs := response.GetDefaultHeaders(0)
	for k, val := range resp.Header {
		if len(val) > 0 {
			hdrs.Set(k, val[len(val)-1])
		}
	}

	hdrs.Remove("Content-Length")
	hdrs.Set("Transfer-Encoding", "chunked")
	hdrs.Set("Trailer", "X-Content-SHA256")
	hdrs.Set("Trailer", "X-Content-Length")

	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(hdrs)
	body := bytes.NewBuffer([]byte{})

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body.Write(buf[:n])
			if _, err := w.WriteChunkedBody(buf[:n]); err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    *bytes.NewBufferString(err.Error()),
				}
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    *bytes.NewBufferString(err.Error()),
			}
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    *bytes.NewBufferString(err.Error()),
		}
	}

	hash := sha256.Sum256(body.Bytes())
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%X", hash))
	trailers.Set("X-Content-Length", strconv.Itoa(body.Len()))
	w.WriteTrailers(trailers)

	return nil
}

func handleVideo(w *response.Writer, req *request.Request) *server.HandlerError {
	hdrs := response.GetDefaultHeaders(0)

	hdrs.Remove("Content-Length")
	hdrs.Set("Transfer-Encoding", "chunked")
	hdrs.Set("Trailer", "X-Content-SHA256")
	hdrs.Set("Trailer", "X-Content-Length")

}
