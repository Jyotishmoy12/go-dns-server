
# Go-DNS-Server — Recursive & Authoritative DNS Resolver

A **high‑performance, recursive, and authoritative DNS server built from scratch in Go**.  
This project demonstrates how the Internet’s DNS hierarchy really works by implementing the **full resolution chain**:

**Root → TLD → Authoritative → Final Answer**

Unlike typical DNS forwarders, this server **does not depend on Google (8.8.8.8) or Cloudflare (1.1.1.1)** — it performs **true iterative resolution** just like a real DNS resolver.

---

## 🌐 What This Project Proves

This project shows that you can build a **real DNS resolver** by understanding:

- How DNS packets are structured
- How referrals and glue records work
- How recursion is actually iterative
- How caching and TTL are enforced
- How resolvers avoid infinite loops
- How concurrency is handled safely

This is not a toy — it is a **working DNS implementation**.

---

## 🚀 Features

### 1. True Recursive Iterative Resolution
Instead of forwarding queries, the resolver starts at the **Root Servers** and walks the hierarchy:

```
Root Server → TLD Server → Authoritative Server → Final Answer
```

The resolver sets:
```
RecursionDesired = false
```
so that upstream servers treat it as a **peer resolver**, not a client.

---

### 2. Glue Record & Sub‑Resolution Handling
When a DNS server refers to another nameserver **without providing its IP**, this resolver:

1. Pauses the main query
2. Resolves the nameserver hostname
3. Continues the original resolution

This mimics how real recursive resolvers work.

---

### 3. High‑Performance DNS Cache
- Implemented using `sync.Map`
- Thread‑safe
- TTL‑aware
- Returns cached answers in **~0ms**
- Prevents repeated upstream queries

---

### 4. Authoritative Overrides (Local DNS)
You can override any domain using `config.json`:

```json
{
  "dev.local.": "127.0.0.1",
  "api.internal.": "192.168.1.50",
  "ads.tracker.com.": "0.0.0.0"
}
```

This allows:
- Local development domains
- Network‑wide ad‑blocking
- Internal service routing

---

### 5. Hot‑Safe Concurrency
The server handles **thousands of concurrent DNS requests** using goroutines.

Thread safety is guaranteed using:

| Purpose | Tool |
|-------|------|
| Config map | `sync.RWMutex` |
| DNS cache | `sync.Map` |
| Network IO | Goroutines |
| Resolution flow | Channels + blocking waits |

Config reloads can happen while queries are being served without crashes.

---

### 6. Zero External Resolver Dependency
No forwarding to:
- Google
- Cloudflare
- ISP DNS

The server performs **100% independent resolution** from the root zone.

---

## 🧠 Architecture

```
┌──────────────┐
│   Client     │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│ DNS Handler           │
│ (Cache → Local → Net) │
└─────────┬────────────┘
          │
          ▼
┌────────────────────────────┐
│ Recursive Navigator         │
│ Root → TLD → Authoritative  │
└─────────┬──────────────────┘
          │
          ▼
┌────────────────────┐
│ Cache + TTL Store   │
└────────────────────┘
```

---

## 📂 Project Structure

```
cmd/
 └── server/
      └── main.go        # UDP listener, socket handling

internal/
 └── dns/
      ├── handler.go    # Query flow: Cache → Local → Resolve
      ├── resolver.go   # Iterative recursive engine
      ├── cache.go      # TTL‑aware concurrent cache

config.json             # Local DNS overrides
test_client.go          # DNS query tester
```

---

## 🛠 How Resolution Works

1. Query arrives
2. Check cache
3. Check local records
4. Start at `A.ROOT-SERVERS.NET (198.41.0.4)`
5. Follow referrals
6. Resolve missing nameservers if needed
7. Store result in cache
8. Return final answer

---

## ⚙️ Installation

### Requirements
- Go **1.21+**
- Package:
```
golang.org/x/net/dns/dnsmessage
```

### Setup

```bash
git clone <repo>
cd Go-DNS-Server
go get golang.org/x/net/dns/dnsmessage
go run ./cmd/server/main.go
```

---

## 🧪 Testing

Use the built‑in DNS client:

```bash
go run test_client.go
```

For domains like `.in`, `.co.uk`, or heavily nested zones, increase timeout to **15–20s** to allow sub‑resolutions.

---

## 🔮 Roadmap

- DNSSEC validation
- IPv6 (AAAA) support
- Web‑based hop visualizer
- Live cache hit dashboard
- Auto‑reload config via fsnotify
