package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}
func isValidHeaderKeyChar(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '!' || b == '#' || b == '$' || b == '%' || b == '&' ||
		b == '\'' || b == '*' || b == '+' || b == '-' || b == '.' ||
		b == '^' || b == '_' || b == '`' || b == '|' || b == '~'
}

func (h Headers) Get(key string) string {
	val, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return val
}
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if len(data) == 0 {
		return 0, false, nil
	}
	if bytes.HasPrefix(data, []byte("\r\n")) {
		return 2, true, nil
	}

	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	line := data[:idx]
	if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
		return 0, false, errors.New("invalid header format: leading whitespace")
	}
	colonIdx := bytes.IndexByte(line, ':')
	if colonIdx == -1 {
		return 0, false, errors.New("invalid header format: missing colon")
	}

	key := strings.TrimSpace(string(line[:colonIdx]))
	val := strings.TrimSpace(string(line[colonIdx+1:]))
	for i := 0; i < len(key); i++ {
		if !isValidHeaderKeyChar(key[i]) {
			return 0, false, errors.New("invalid character in header key")
		}
	}
	lowerKey := strings.ToLower(key)

	if curVal, exists := h[lowerKey]; exists {
		h[lowerKey] = curVal + ", " + val
	} else {
		h[lowerKey] = val
	}

	return idx + 2, false, nil
}
