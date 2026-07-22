package response

import (
	"fmt"
	"io"
	"strconv"
	"tcp_to_http/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func statusText(code StatusCode) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	text := statusText(statusCode)
	if text == "" {
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", int(statusCode))
		return err
	}
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", int(statusCode), text)
	return err
}
func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-type":   "text/plain",
		"content-length": strconv.Itoa(contentLen),
		"connection":     "close",
	}
}
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	return WriteStatusLine(w.w, statusCode)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	return WriteHeaders(w.w, headers)

}

func (w *Writer) WriteBody(body []byte) (int, error) {
	return w.w.Write(body)
}
