package response_test

import (
	"bytes"
	"testing"

	"tcp_to_http/internal/response"
)

func TestWriteChunkedBody(t *testing.T) {
	var buf bytes.Buffer
	w := response.NewWriter(&buf)

	data := []byte("hello world!")
	n, err := w.WriteChunkedBody(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected n=%d, got %d", len(data), n)
	}

	expected := "c\r\nhello world!\r\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestWriteChunkedBodyDone(t *testing.T) {
	var buf bytes.Buffer
	w := response.NewWriter(&buf)

	_, err := w.WriteChunkedBodyDone()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "0\r\n\r\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestWriteChunkedBodyEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := response.NewWriter(&buf)

	n, err := w.WriteChunkedBody([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected n=0, got %d", n)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty buffer, got %q", buf.String())
	}
}

func TestWriteTrailers(t *testing.T) {
	var buf bytes.Buffer
	w := response.NewWriter(&buf)

	h := make(map[string]string)
	h["x-content-sha256"] = "abc123hash"
	h["x-content-length"] = "100"

	err := w.WriteTrailers(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !bytes.HasPrefix([]byte(out), []byte("0\r\n")) {
		t.Errorf("expected output to start with '0\\r\\n', got %q", out)
	}
	if !bytes.HasSuffix([]byte(out), []byte("\r\n")) {
		t.Errorf("expected output to end with '\\r\\n', got %q", out)
	}
}
