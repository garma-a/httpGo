<h1 align="center">httpGo 🚀</h1>

<p align="center">
  <img src="imgs/logo.png" alt="httpGo Logo" width="500" />
</p>

`httpGo` is a low-level, high-performance TCP-to-HTTP request parser and server built from scratch in Go. Rather than relying on Go's built-in `net/http` package, `httpGo` reads raw TCP streams, manages dynamic memory buffers, and implements an RFC-compliant state machine parser to dissect HTTP request lines and headers.

---

## 🛠️ Features

* **Concurrently Accepts TCP Connections**: Spawns independent goroutines to handle multiplexed connections on port `:42069`.
* **Zero-Dependency HTTP Parser**: Reads raw socket bytes directly.
* **HTTP Request Line Parsing**: Extracts the `Method`, `RequestTarget`, and `HttpVersion`.
* **Advanced HTTP Header Parser**:
  * **Incremental Streaming Support**: Safely parses headers line-by-line across arbitrary TCP chunk boundaries.
  * **RFC-Compliant Token Validation**: Detects invalid characters (like `©` or leading whitespace) in header keys and rejects bad requests.
  * **Case-Insensitive Normalization**: Canonicalizes header keys to lowercase (e.g., `User-Agent` becomes `user-agent`).
  * **Multi-Value Header Merging**: Automatically combines duplicate header keys into a comma-separated list (e.g., merging multiple `Set-Cookie` headers into a single value, per RFC 9110).

---

## 🏗️ Code Concepts & Parser Architecture

### 1. The Parser State Machine
To handle network streams that arrive in unpredictable chunk sizes, `httpGo` uses an internal state machine defined in the `request` package:

```go
const (
	requestStateInitialized state = iota // Waiting for/parsing Request Line
	requestStateParsingHeaders          // Parsing individual HTTP headers
	requestStateDone                    // Parsing complete
)
```

The parser cycles through these states in a loop, ensuring that it only parses what has been fully received.

### 2. Dynamic Buffering & Byte Shifting
Since TCP sends data in packets of variable sizes, the server starts with a small **8-byte buffer** to save memory. 
If the buffer fills up before a full HTTP boundary (like `\r\n`) is reached, `httpGo`:
1. **Doubles the buffer size** dynamically.
2. Reads the next chunk from the socket.
3. After parsing is successful, **shifts** the unparsed bytes back to the beginning of the buffer to free up space.

```
Buffer:
+-------------------+-----------------------------+
|  Parsed Request   |      Unparsed Headers       |
+-------------------+-----------------------------+
\______  ___________/
       \/ (consumed bytes)
       
Shifting:
+-----------------------------+-------------------+
|      Unparsed Headers       |    Empty Space    |
+-----------------------------+-------------------+
```

### 3. Case Insensitivity & Merging (RFC 9110)
Standard HTTP allows header keys to be case-insensitive, and multiple lines of the same header key are merged into a comma-separated value:

```http
Set-Cookie: session=123\r\n
SET-COOKIE: theme=dark\r\n
```

Is parsed and stored internally as:
```go
map[string]string{
    "set-cookie": "session=123, theme=dark",
}
```

---

## 📂 Project Structure

* **[cmd/tcplistener/main.go](file:///home/girgis/repo/learning/boot_dev/httpGo/cmd/tcplistener/main.go)**: Starts the TCP listener, accepts TCP connections, and hands them off to the request parser.
* **[internal/request/request.go](file:///home/girgis/repo/learning/boot_dev/httpGo/internal/request/request.go)**: Manages buffering, byte shifting, and orchestrates request line and header parsing.
* **[internal/headers/headers.go](file:///home/girgis/repo/learning/boot_dev/httpGo/internal/headers/headers.go)**: Handles individual header line token validation, parsing, lowercasing, and duplicate key merging.

---

## 🚀 How to Run the Project

To start the TCP-to-HTTP server:

```bash
go run cmd/tcplistener/main.go
```

The server will start listening on port `:42069`:
```
Listening on :42069...
```

You can send standard HTTP requests using `curl` or `netcat` in another terminal:

```bash
curl http://localhost:42069/coffee -H "User-Agent: test-agent"
```

The server console will output:
```
connection accepted
Request line:
- Method: GET
- Target: /coffee
- Version: 1.1
```

---

## 🧪 How to Run Tests

The project has robust unit tests covering edge cases like malformed headers, invalid characters, multi-value headers, and different chunk sizes.

Run all tests:
```bash
go test -v ./...
```
