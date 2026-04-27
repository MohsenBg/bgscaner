# Settings and Configuration

All scanner configuration files are stored inside the `settings` directory.  
Each subsystem has its own dedicated `.toml` file, making the scanner modular, maintainable, and easy to extend.

```
settings/
├── dns_settings.toml
├── general_settings.toml
├── http_settings.toml
├── icmp_settings.toml
├── tcp_settings.toml
├── writer_settings.toml
└── xray_settings.toml
```

Every configuration file contains inline comments explaining each parameter.  
For deeper explanations, see the corresponding documentation pages listed below.

---

# Configuration Modules

## General Settings  
**File:** `general_settings.toml`

Controls global scanner behavior, execution flow, and the overall pipeline model.

Core responsibilities:

- Status update intervals  
- Scan limits (max results, max IPs)  
- Verbose debugging  
- Inter‑module pipeline behavior  
- Channel buffering between scan stages  

Execution modes:

- **simple** – sequential, uses disk between stages  
- **parallel** – fully concurrent, channel‑based  
- **pipeline** – hybrid batch execution  

Documentation: `general_settings.md`

---

## DNS Settings  
**File:** `dns_settings.toml`

Controls all DNS‑related scan modules:

### 1. DNS Resolver Scanner
Detects usable public DNS resolvers by testing:

- Reachability  
- Correct protocol behavior  
- Recursive capabilities  
- Tunnel suitability  

Features:

- High‑concurrency DNS probing  
- UDP, TCP, DoT support  
- Random subdomains for cache bypass  
- Response code filtering  
- Anti‑hijack (DPI) checks  
- EDNS buffer tuning  

### 2. DNSTT Probe
Tests whether resolvers support a **full DNSTT tunnel handshake**.  
Requires:

- Delegated DNS zone  
- DNSTT server (Ed25519 public key)  

### 3. Slipstream Probe
Alternative DNS‑based tunneling technique.  
Checks resolver compatibility with Slipstream packet patterns.

Documentation: `dns_settings.md`

---

## HTTP Scanner Settings  
**File:** `http_settings.toml`

Tests HTTP/HTTPS reachability and server behavior.

Capabilities:

- Custom Host / Path  
- TLS version control and validation  
- SNI configuration  
- Concurrency settings  
- Request timeouts  

Useful for detecting:

- Web servers  
- CDN fronting endpoints  
- Reverse proxies  

Documentation: `http_settings.md`

---

## ICMP Scanner Settings  
**File:** `icmp_settings.toml`

Performs fast ICMP Echo Request probing.

Features:

- Configurable timeout  
- Retry attempts  
- Worker concurrency  
- Optional IP randomization  

Typical use case:

- First‑stage alive-host filtering before deeper scans  

Documentation: `icmp_settings.md`

---

## TCP Scanner Settings  
**File:** `tcp_settings.toml`

Attempts plain TCP connection to a target port.

Used for detecting:

- Web services (80/443)  
- SSH (22)  
- Custom ports  

Key options:

- Target port  
- Connection timeout  
- Retries  
- Worker count  
- IP shuffling  

Documentation: `tcp_settings.md`

---

## Writer Settings  
**File:** `writer_settings.toml`

Controls how scan results are buffered and written to disk.

Design highlights:

- Batched asynchronous writes  
- Buffered result channel  
- Periodic flush  
- Avoids writer bottlenecks under heavy load  

Documentation: `writer_settings.md`

---

## Xray Scanner Settings  
**File:** `xray_settings.toml`

Runs advanced connectivity & performance tests using **Xray core**.

Supported modes:

- Connectivity only  
- Download speed  
- Upload speed  
- Full speed test  

Additional features:

- Bandwidth caps  
- Worker concurrency  
- Optional pre‑scan validation (ICMP/TCP/HTTP)  

Documentation: `xray_settings.md`
