package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp_to_http/internal/request"
	"tcp_to_http/internal/response"
	"tcp_to_http/internal/server"
)

const (
	htmlBadRequest = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	htmlInternalError = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

	htmlSuccess = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
)

func writeResponse(w *response.Writer, status response.StatusCode, body string) {
	headers := response.GetDefaultHeaders(len(body))
	headers.Set("Content-Type", "text/html")

	w.WriteStatusLine(status)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(body))
}

type ServeMux struct {
	routes map[string]server.Handler
}

func (m *ServeMux) HandleFunc(pattern string, handler server.Handler) {
	if m.routes == nil {
		m.routes = make(map[string]server.Handler)
	}
	m.routes[pattern] = handler
}

func (m *ServeMux) Serve(w *response.Writer, req *request.Request) {
	if handler, ok := m.routes[req.RequestLine.RequestTarget]; ok {
		handler(w, req)
		return
	}
	// Fallback to 200 Success like original behavior
	if fallback, ok := m.routes["/"]; ok {
		fallback(w, req)
		return
	}
}

func main() {
	port := flag.Int("port", 42069, "Port to run the server on")
	flag.Parse()

	mux := &ServeMux{}
	mux.HandleFunc("/yourproblem", func(w *response.Writer, req *request.Request) {
		writeResponse(w, response.StatusBadRequest, htmlBadRequest)
	})
	mux.HandleFunc("/myproblem", func(w *response.Writer, req *request.Request) {
		writeResponse(w, response.StatusInternalServerError, htmlInternalError)
	})
	mux.HandleFunc("/", func(w *response.Writer, req *request.Request) {
		writeResponse(w, response.StatusOK, htmlSuccess)
	})

	srv, err := server.Serve(*port, mux.Serve)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()

	log.Printf("Server started on port %d\n", *port)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
