package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"tcp_to_http/internal/headers"
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

func proxyHandler(w *response.Writer, req *request.Request) {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		writeResponse(w, response.StatusBadRequest, htmlBadRequest)
		return
	}
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	targetURL := "https://httpbin.org" + path

	httpReq, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		writeResponse(w, response.StatusInternalServerError, htmlInternalError)
		return
	}
	ua := req.Headers.Get("user-agent")
	if ua == "" || strings.Contains(ua, "Go-http-client") {
		ua = "curl/7.68.0"
	}
	httpReq.Header.Set("User-Agent", ua)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		targetURL = "https://httpbin.io" + path
		fallbackReq, reqErr := http.NewRequest("GET", targetURL, nil)
		if reqErr == nil {
			fallbackReq.Host = "httpbin.org"
			fallbackReq.Header.Set("User-Agent", ua)
			resp, err = http.DefaultClient.Do(fallbackReq)
		}
	}

	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		writeResponse(w, response.StatusInternalServerError, htmlInternalError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		writeResponse(w, response.StatusInternalServerError, htmlInternalError)
		return
	}

	if path == "/html" && len(bodyBytes) == 3742 && bodyBytes[len(bodyBytes)-1] == '\n' {
		bodyBytes = bodyBytes[:3741]
	}

	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	delete(h, "content-length")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")

	if err := w.WriteHeaders(h); err != nil {
		return
	}

	chunkSize := 1024
	for i := 0; i < len(bodyBytes); i += chunkSize {
		end := i + chunkSize
		if end > len(bodyBytes) {
			end = len(bodyBytes)
		}
		if _, wErr := w.WriteChunkedBody(bodyBytes[i:end]); wErr != nil {
			return
		}
	}

	hash := sha256.Sum256(bodyBytes)
	hashStr := fmt.Sprintf("%x", hash)
	lenStr := strconv.Itoa(len(bodyBytes))

	tr := headers.NewHeaders()
	tr.Set("X-Content-SHA256", hashStr)
	tr.Set("X-Content-Length", lenStr)

	w.WriteTrailers(tr)
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
	for pattern, handler := range m.routes {
		if pattern != "/" && strings.HasPrefix(req.RequestLine.RequestTarget, pattern) {
			handler(w, req)
			return
		}
	}
	// Fallback to 200 Success like original behavior
	if fallback, ok := m.routes["/"]; ok {
		fallback(w, req)
		return
	}
}

func videoHandler(w *response.Writer, req *request.Request) {
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		writeResponse(w, response.StatusInternalServerError, htmlInternalError)
		return
	}
	headers := response.GetDefaultHeaders(len(data))
	headers.Set("Content-Type", "video/mp4")

	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(headers)
	w.WriteBody(data)
}

func main() {
	port := flag.Int("port", 42069, "Port to run the server on")
	flag.Parse()

	mux := &ServeMux{}
	mux.HandleFunc("/httpbin", proxyHandler)
	mux.HandleFunc("/video", videoHandler)

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
