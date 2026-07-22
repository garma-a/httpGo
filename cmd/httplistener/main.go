package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp_to_http/internal/request"
	"tcp_to_http/internal/response"
	"tcp_to_http/internal/server"
)

const port = 42069

func handler(w *response.Writer, req *request.Request) {
	var body string
	var status response.StatusCode

	if req.RequestLine.RequestTarget == "/yourproblem" {
		status = response.StatusBadRequest
		body = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		status = response.StatusInternalServerError
		body = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
	} else {
		status = response.StatusOK
		body = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Set("Content-Type", "text/html")

	w.WriteStatusLine(status)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(body))
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
