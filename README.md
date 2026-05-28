# HTTP-from-TCP: Building HTTP from the Ground Up

A **learning-focused HTTP server implementation** built entirely FROM SCRATCH on top of TCP connections. This project demonstrates how HTTP works at the protocol level by parsing raw TCP streams, implementing HTTP message parsing, and building a complete web server with chunked transfer encoding and binary data serving.

Available in **two languages** for comparison: Go (original) and Python.

## Why This Project?

Understanding HTTP is fundamental to web development, but most developers only interact with high-level HTTP libraries. This project peels back the abstraction layers to show:

- **How HTTP actually works** over TCP connections
- **Raw HTTP message parsing** without libraries
- **Protocol-level implementation** of features like chunked encoding
- **Binary data handling** in text-based protocols

## Quick Start

### Go version

```bash
go run ./cmd/httpserver

curl http://localhost:42069/
```

### Python version

```bash
cd python-version
pip install -r requirements.txt
python app/httpserver/main.py

curl http://localhost:42069/
```

## Project Structure

```
http-from-tcp/
├── cmd/                        # Go entry points
│   ├── httpserver/main.go      #   HTTP server completo
│   └── tcplistener/main.go     #   TCP listener (debug)
├── internal/                   # Go internal packages
│   ├── headers/                #   Header parsing
│   ├── request/                #   Request parsing (state machine)
│   ├── response/               #   Response writing (state machine)
│   └── server/                 #   TCP server
├── python-version/             # Python rewrite
│   ├── app/
│   │   ├── httpserver/main.py  #   HTTP server completo
│   │   └── tcplistener/main.py #   TCP listener (debug)
│   ├── internal/
│   │   ├── headers/            #   Header parsing
│   │   ├── request/            #   Request parsing (state machine)
│   │   ├── response/           #   Response writing (state machine)
│   │   └── server/             #   TCP server
│   ├── tests/                  #   Unit tests (pytest)
│   └── requirements.txt
├── assets/                     # Shared static assets (video)
├── go.mod
└── README.md
```

## Available Endpoints

| Endpoint | Description | Response Type |
|----------|-------------|---------------|
| `GET /` | Default success page | HTML |
| `GET /yourproblem` | 400 Bad Request example | HTML |
| `GET /myproblem` | 500 Internal Server Error | HTML |
| `GET /video` | Streams a video file | Binary (video/mp4) |
| `GET /httpbin/*` | Proxies requests to httpbin.org | Chunked JSON + Trailers |

## Architecture Overview

Both implementations share the same architecture:

```
TCP Connection
    │
    ▼
┌─────────────────────┐
│  Request Parser     │  ← State machine: INIT → HEADERS → BODY → DONE
│  (incremental)      │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Router / Handler   │  ← Matches path, dispatches to handler
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Response Writer    │  ← State machine: INITIAL → STATUS → HEADERS → BODY
│  (status/headers/   │
│   body/chunked)     │
└─────────┬───────────┘
          │
          ▼
    TCP Connection
```

---

## Go vs Python Comparison

### Concurrency Model

| Aspect | Go | Python |
|--------|------|--------|
| Per-connection handling | `go s.handle(conn)` (goroutine) | `threading.Thread(target=..., daemon=True)` |
| Overhead | ~2KB per goroutine | ~8MB stack per thread (OS-level) |
| Scaling | Millions of goroutines | Limited by OS threads |
| Synchronization | `sync/atomic`, channels | `threading.Event` |

### TCP / IO

| Aspect | Go | Python |
|--------|------|--------|
| Listening | `net.Listen("tcp", ":42069")` | `socket.socket(AF_INET, SOCK_STREAM)` + `bind` + `listen` |
| Accept | `listener.Accept()` | `sock.accept()` |
| Reading | `conn.Read(buf)` — returns available bytes | `reader.read1(n)` — returns available bytes without blocking |
| Writing | `conn.Write(b)` / `io.WriteString` | `conn.write(b)` on file-like |
| Close | `defer conn.Close()` | `try/finally: conn.close()` |

### Error Handling

| Aspect | Go | Python |
|--------|------|--------|
| Style | Return `error` as second value | Raise exceptions |
| Custom errors | `fmt.Errorf(...)` / sentinel vars | Exception subclasses |
| Propagation | Explicit `if err != nil` checks | `try/except` blocks |

### Type System

| Aspect | Go | Python |
|--------|------|--------|
| Headers | `type Headers map[string]string` | `class Headers(dict[str, str])` |
| Request | `struct` with fields | `@dataclass` |
| Status codes | `type StatusCode int` + `const` | `class StatusCode(IntEnum)` |
| State machine | `const iota` | `class ParseState(IntEnum)` + `auto()` |
| Interfaces | `io.Reader` / `io.Writer` | `BinaryIO` (typing protocol) |

### Code Comparison — Request Parsing

**Go:**
```go
func (r *Request) parseSingle(data []byte) (int, error) {
    switch r.State {
    case StateInit:
        requestLine, bytesConsumed, err := parseRequestLines(data)
        if err != nil { return 0, err }
        if bytesConsumed == 0 { return 0, nil }
        r.RequestLine = *requestLine
        r.State = StateParsingHeaders
        return bytesConsumed, nil
    case StateParsingHeaders:
        // ...
    }
}
```

**Python:**
```python
def _parse_single(req: Request, data: bytes) -> int:
    if req.state == ParseState.INIT:
        return _parse_request_line(req, data)
    elif req.state == ParseState.PARSING_HEADERS:
        return _parse_headers(req, data)
    elif req.state == ParseState.PARSING_BODY:
        return _parse_body(req, data)
    elif req.state == ParseState.DONE:
        raise MalformedRequestError("trying to parse when already done")
```

### Code Comparison — Response Writing

**Go:**
```go
func (w *Writer) WriteChunkedBody(p []byte) (n int, err error) {
    chunkSize := fmt.Sprintf("%x\r\n", len(p))
    _, err = io.WriteString(w.conn, chunkSize)
    n, err = w.conn.Write(p)
    _, err = io.WriteString(w.conn, "\r\n")
    return n, nil
}
```

**Python:**
```python
def write_chunked_body(self, chunk: bytes) -> int:
    size_line = f"{len(chunk):x}\r\n".encode("ascii")
    self._conn.write(size_line)
    self._conn.write(chunk)
    self._conn.write(b"\r\n")
    return len(chunk)
```

### Key Takeaways

| | Go | Python |
|--|------|--------|
| **Strengths** | Zero-cost goroutines, explicit error handling, fast binary handling | Readable, rapid prototyping, rich stdlib |
| **Weaknesses** | Verbose error checks, no exceptions | GIL limits true parallelism, thread overhead |
| **Best for** | Production network services | Learning/prototyping, scripting |
| **Lines of code** | ~420 | ~380 |

## Running Tests

### Go

```bash
go test ./...
```

### Python

```bash
cd python-version
python -m pytest tests/ -v
```

## Key Concepts Taught

1. **Incremental parsing with state machines** — don't assume full data arrives at once
2. **Response writer with state protection** — enforce correct HTTP message order
3. **Chunked Transfer Encoding** — for streaming responses of unknown size
4. **HTTP Trailers** — metadata sent after the body (SHA256 hash)
5. **Binary data in text protocols** — serving video over HTTP from scratch

## Credits

Original Go implementation by [@taham8875](https://github.com/taham8875/http-from-tcp).
Python rewrite for educational comparison.
