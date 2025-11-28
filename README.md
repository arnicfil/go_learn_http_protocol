# go_learn_http_protocol

A small, educational HTTP server implemented in Go to learn about the HTTP protocol internals: parsing requests, writing status lines/headers/bodies, chunked transfer encoding, trailers, and simple proxying. The project intentionally implements many pieces that standard libraries normally hide so you can see how HTTP works on the wire.

This repository contains:
- a tiny server that accepts TCP connections and parses raw HTTP requests
- a `response` package that provides an explicit `Writer` which can write status-line, headers, body, chunked body, and trailers
- a `headers` package that models header parsing/manipulation with normalized keys
- a sample `httpserver` binary (cmd/httpserver) that demonstrates handlers, including an `/httpbin/*` proxy that streams chunked responses from https://httpbin.org

Goals
- Understand the HTTP/1.1 message format (status line, header block, CRLF, body)
- Practice implementing chunked transfer encoding and trailers
- Explore how to implement a minimal router/handler model without net/http abstractions
- Learn how to proxy and stream responses from an upstream server

Quick build & run

1. Build or run the example server:
- Run directly:
  go run ./cmd/httpserver

- Build a binary:
  go build -o httpserver ./cmd/httpserver
  ./httpserver

2. The server listens on port 42069 by default (see cmd/httpserver/main.go). Use port 0 in tests if you want the OS to pick a free port.

Usage examples

Raw request via netcat (good for inspecting raw chunk framing):
- Show a simple request
  echo -e "GET / HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069

- Request the proxy stream endpoint (shows raw hexadecimal chunk sizes and CRLFs):
  echo -e "GET /httpbin/stream/20 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069

Using curl
- curl will usually de-chunk the response for you. To see headers:
  curl -i --http1.1 "http://localhost:42069/httpbin/stream/20"

- To ask curl to try to show raw chunk frames (behavior can vary by platform/build):
  curl --raw -i -N --http1.1 "http://localhost:42069/httpbin/stream/20"

Project layout (high level)
- cmd/httpserver: example server file that registers handlers and demonstrates HTML responses and the `/httpbin/*` streaming proxy
- internal/request: request parsing utilities (parses request-line, headers, etc.)
- internal/headers: header container and parser (normalizes keys, parse, set, remove)
- internal/response: Writer type and helper functions to compose HTTP responses (status line, headers, body, chunked writes, trailers)
- internal/server: server loop, listener, and handler dispatch glue

Key APIs and conventions

Handler
- type Handler func(w *response.Writer, req *request.Request) *HandlerError
- Handlers receive a response.Writer (not plain io.Writer) so they can set headers, write raw []byte body, write chunked bodies, and call WriteTrailers when needed.
- Handlers return a *HandlerError when they want the server to write an error status (but once headers or body are flushed, you cannot write a fresh status line; handlers should prefer writing an error body and status via the Writer before flush).

response.Writer (high-level)
- NewWriter(w io.Writer) *Writer
- (w *Writer) WriteStatusLine(statusCode response.StatusCode) error
- (w *Writer) WriteHeaders(headers.Headers) error
- (w *Writer) WriteBody(p []byte) (int, error)
- (w *Writer) WriteChunkedBody(p []byte) (int, error)
- (w *Writer) WriteChunkedBodyDone() (int, error)
- (w *Writer) WriteTrailers(h headers.Headers) error
- Writer implements writing order/state checks (status -> headers -> body/chunks -> trailers) to help you spot protocol misuse.

Chunked proxy behavior
- The `/httpbin/*` handler fetches an upstream URL at https://httpbin.org/<path>, strips Content-Length, sets Transfer-Encoding: chunked, writes status+headers, then reads the upstream body in a loop and forwards each read as a chunk immediately using WriteChunkedBody. After EOF it writes the zero chunk and then writes any trailers computed (e.g. SHA256 and total length) via WriteTrailers.
```
