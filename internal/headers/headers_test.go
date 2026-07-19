package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header (should result in lowercase key)
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"]) // lowercase check
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Case-insensitive lowercasing of keys
	headers = NewHeaders()
	data = []byte("User-Agent: curl/7.81.0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "curl/7.81.0", headers["user-agent"]) // lowercase check
	assert.Equal(t, 25, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
	// Test: Multiple values for a single header key
	headers = NewHeaders()
	headers["set-person"] = "lane-loves-go"
	data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.Equal(t, 29, n)
	assert.False(t, done)
}
