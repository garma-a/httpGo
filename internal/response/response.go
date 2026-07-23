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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	hexLen := strconv.FormatInt(int64(len(p)), 16)
	if _, err := w.w.Write([]byte(hexLen + "\r\n")); err != nil {
		return 0, err
	}
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}
	if _, err := w.w.Write([]byte("\r\n")); err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.w.Write([]byte("0\r\n\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if _, err := w.w.Write([]byte("0\r\n")); err != nil {
		return err
	}
	return WriteHeaders(w.w, h)
}
