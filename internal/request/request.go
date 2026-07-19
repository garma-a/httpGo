package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"tcp_to_http/internal/headers"
)

// Internal enum for parser state
type state int

const (
	requestStateInitialized state = iota
	requestStateParsingHeaders
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       state
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// parseRequestLine returns the parsed RequestLine, the number of bytes consumed, and an error.
// If it can't find an \r\n, it returns 0 and no error.
func parseRequestLine(data []byte) (RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		// Not enough data yet. Return 0 consumed bytes.
		return RequestLine{}, 0, nil
	}

	line := string(data[:idx])
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("invalid request line")
	}

	version := strings.TrimPrefix(parts[2], "HTTP/")

	reqLine := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version,
	}

	// Consume the line length plus the 2 bytes for \r\n
	return reqLine, idx + 2, nil
}

// parseSingle parses a single component of the request based on the current state.
func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		reqLine, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		// If we successfully consumed data, update the state
		if consumed > 0 {
			r.RequestLine = reqLine
			r.state = requestStateParsingHeaders
		}
		return consumed, nil

	case requestStateParsingHeaders:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}
		consumed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if consumed > 0 {
			if done {
				r.state = requestStateDone
			}
		}
		return consumed, nil

	case requestStateDone:
		return 0, nil
	}

	return 0, nil
}

// parse accepts unparsed bytes, updates the state, and returns consumed bytes in a loop
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}

	// Start with a small buffer as requested
	buffer := make([]byte, 8)
	totalRead := 0

	// Loop until the parser is completely done
	for req.state != requestStateDone {
		// If our buffer is full but we still need more data, double its size
		if totalRead == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		// Read into the remaining empty space in the buffer
		n, err := reader.Read(buffer[totalRead:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		// If the stream ended but we aren't done parsing, it's an incomplete request
		if n == 0 && err == io.EOF {
			if req.state != requestStateDone {
				return nil, errors.New("unexpected EOF")
			}
			break
		}

		totalRead += n

		// Pass the currently accumulated data to the parser
		consumed, parseErr := req.parse(buffer[:totalRead])
		if parseErr != nil {
			return nil, parseErr
		}

		// If the parser consumed bytes, shift the remaining unparsed data
		// to the front of the buffer so we don't hold onto old data.
		if consumed > 0 {
			copy(buffer, buffer[consumed:totalRead])
			totalRead -= consumed
		}
	}

	return req, nil
}
