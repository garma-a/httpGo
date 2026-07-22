package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"tcp_to_http/internal/request"
	"tcp_to_http/internal/response"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

func (hErr *HandlerError) Error() string {
	return fmt.Sprintf("handler error: %s (status code: %d)", hErr.Message, hErr.StatusCode)
}

type Handler func(w *response.Writer, req *request.Request)

func (hErr *HandlerError) Write(w io.Writer) error {
	err := response.WriteStatusLine(w, response.StatusCode(hErr.StatusCode))
	if err != nil {
		return fmt.Errorf("failed to write status line: %w", err)
	}
	headers := response.GetDefaultHeaders(len(hErr.Message))
	err = response.WriteHeaders(w, headers)
	if err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}
	_, err = w.Write([]byte(hErr.Message))
	if err != nil {
		return fmt.Errorf("failed to write error message: %w", err)
	}
	return nil
}

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s := &Server{
		listener: listener,
		handler:  handler,
	}
	go s.listen()
	return s, nil

}
func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}
	writer := response.NewWriter(conn)
	s.handler(writer, req)

}
