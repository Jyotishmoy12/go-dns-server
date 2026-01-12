# GoDNS-from-Scratch

A high-performance, concurrent DNS Recursive Resolver and Authoritative Server built in Go using industry best practices.

## 🚀 Features

- **Recursive Resolution**: Forwards unknown queries to upstream providers (Google 8.8.8.8).
- **Authoritative Support**: Serves local custom records for internal domains.
- **High-Concurrency**: Uses a Goroutine-per-packet model for high throughput.
- **Intelligent Caching**: Thread-safe `sync.Map` cache with support for Dynamic TTL (Time To Live).
- **Performance Tracking**: Real-time logging of latency (ms vs µs).
- **Robustness**: Defensive programming with nil-pointer checks and binary packet validation.

## 📂 Project Structure

```plaintext
go-dns-server/
├── cmd/
│   └── server/
│       └── main.go       # UDP Socket listener and Signal handling
├── internal/
│   ├── dns/
│   │   ├── handler.go    # Main logic (The "Brain")
│   │   ├── resolver.go   # Upstream communication (UDP Client)
│   │   └── cache.go      # Thread-safe in-memory storage
├── go.mod                # Module definition
└── test_client.go        # Diagnostic tool to verify responses
```

## 🛠️ Implementation Details

### 1. The Decision Tree

The server follows a specific priority order for every incoming packet:

1. **Cache Check**: If valid data exists in RAM, respond in `<1ms`.
2. **Local Check**: If it's a "known" local domain, respond with the hardcoded IP.
3. **Recursion**: If unknown, dial the upstream DNS, wait for the response, cache it, and return.

### 2. Best Practices Used

- **Binary Safety**: Used `golang.org/x/net/dns/dnsmessage` to handle complex DNS pointer compression safely.
- **Timeouts**: Implemented `net.DialTimeout` to prevent server "hanging" on network failures.
- **Graceful Shutdown**: Listens for `OS.Interrupt` signals to close UDP ports cleanly.
- **ID Synchronization**: Ensures the Transaction ID of the response matches the query (prevents client-side spoofing errors).

## 🚦 Getting Started

### Prerequisites

- Go 1.21 or higher.

### Running the Server

```powershell
# From the project root
go run ./cmd/server/main.go
```

### Testing the Resolver

```powershell
# In a separate terminal
go run test_client.go
```

## 📈 Performance Benchmarks

| Query Type               | Latency (Approx) | Resource                     |
| ------------------------ | ---------------- | ---------------------------- |
| First Request (Miss)     | 70ms - 100ms     | Network I/O (Google 8.8.8.8) |
| Subsequent Request (Hit) | < 1ms (0.0001s)  | RAM (sync.Map)               |

## 📜 License

MIT License — Feel free to use this for learning or as a base for your own networking tools.
