# DNS Scanner Configuration

**File:** `settings/dns_settings.toml`  
**Modules Covered:**

- Bulk DNS Resolver Scanner
- DNSTT Tunnel Probe
- Slipstream Probe

This configuration file controls all DNS‑based scanning and tunneling behavior in the system.

---

# Architecture Overview

```
Resolver Scan
   ↓
Filtering (Rcode / DPI / Behavior)
   ↓
Surviving Resolvers
   ↓
DNSTT Probe (optional)
   ↓
Slipstream Probe (optional)
   ↓
Output Writer
```

---

# [resolver] — Bulk DNS Resolver Scanner

Scans public DNS resolvers to determine:

- Reachability
- Protocol behavior
- Recursion capability
- Filtering characteristics

Results are used by DNSTT and Slipstream probes.

---

## Core Settings

### `workers`

```toml
workers = 500
```

**Type:** Integer  
**Recommended Range:** 50–500

Number of concurrent goroutines sending DNS queries.

Higher values increase:

- Scan speed
- Bandwidth usage
- Packet rate

---

### `protocol`

```toml
protocol = "udp"
```

**Type:** String  
**Supported:**

| Value | Description |
|-------|------------|
| `"udp"` | Standard DNS over UDP |
| `"tcp"` | DNS over TCP |
| `"dot"` | DNS over TLS (port 853) |

⚠️ `"doh"` (DNS-over-HTTPS) is not supported.

---

### `domain`

```toml
domain = ""
```

Domain used for testing resolvers.

- Should be a stable domain
- For tunneling compatibility, must be under your control

---

### `port`

```toml
port = 53
```

DNS port used for resolver scanning.

Typical values:

| Protocol | Port |
|----------|------|
| UDP/TCP | 53 |
| DoT | 853 |

---

## Query Configuration

### `check_types`

```toml
check_types = ["txt"]
```

Record types to query.

Supported examples:

- `"a"`
- `"aaaa"`
- `"ns"`
- `"txt"`
- `"mx"`

✅ For DNSTT compatibility → `"txt"` is strongly recommended.

---

### `ends_buffer_size`

```toml
ends_buffer_size = 0
```

EDNS buffer size in bytes.

- `0` → Disable EDNS OPT record
- Typical EDNS sizes → 512–4096

Larger values allow bigger UDP payloads.

---

## Timing & Retry

### `timeout`

```toml
timeout = 2000
```

Milliseconds to wait for response.

Lower → faster scan  
Higher → more tolerant of latency

---

### `tries`

```toml
tries = 2
```

Network retries for resolver probe.

Important behavior:

- Retries occur only on **network errors or timeout**
- If ANY DNS response is received (even SERVFAIL or REFUSED), retry stops

---

## Behavioral Controls

### `shuffle_ips`

```toml
shuffle_ips = true
```

Randomizes resolver list to distribute load.

---

### `random_subdomain`

```toml
random_subdomain = true
```

If enabled:

```
x7a91k.example.com
```

is generated to bypass resolver caches.

Purpose:

- Forces recursive lookup
- Prevents cached responses
- Reveals real resolver behavior

---

## Response Filtering

### `accepted_rcodes`

```toml
accepted_rcodes = ["noerror", "nxdomain"]
```

Defines which DNS Rcodes are considered “Alive”.

### Supported Rcodes

| Name | Code |
|------|------|
| noerror / success | 0 |
| formerr | 1 |
| servfail | 2 |
| nxdomain | 3 |
| notimp | 4 |
| refused | 5 |

### Recommended

```toml
["noerror", "nxdomain"]
```

Why?

- `NOERROR` → Valid successful resolution
- `NXDOMAIN` → Valid recursive behavior

---

## DPI / Anti-Hijacking Detection

Resolvers sometimes hijack NXDOMAIN responses.

### `check_dpi`

```toml
check_dpi = false
```

If enabled:

- Scanner queries a `.invalid` domain
- If resolver returns `NOERROR`
- It is flagged as hijacked/tampered and discarded

---

### `dpi_timeout`

```toml
dpi_timeout = 500
```

Shorter timeout for DPI pre-check.

---

### `dpi_tries`

```toml
dpi_tries = 2
```

Retries only on timeout/network errors.

---

## Output

```toml
prefix_output = "dns_"
```

Example:

```
dns_20260427_210734.csv
```

---

# [dnstt] — DNSTT Tunnel Probe

Tests whether surviving resolvers can carry a full DNS tunnel.

Requires:

- Authoritative NS under your control
- DNSTT server running

---

## Core

### `enabled`

```toml
enabled = false
```

Enable/disable DNSTT probing phase.

---

### `workers`

```toml
workers = 20
```

Recommended: 5–50

Handshake is heavier than resolver scan.

---

### `domain`

```toml
domain = ""
```

Authoritative zone delegated to DNSTT server.

---

### `public_key`

```toml
public_key = ""
```

Base64‑encoded Ed25519 public key.

Used to verify server identity.

---

## Timing

```toml
timeout = 10000
```

Time allowed for full handshake.

---

## Output

```toml
prefix_output = "dnstt_"
```

---

# [slip_stream] — Slipstream DNS Probe

Alternative DNS tunneling technique.

Exploits DNS behaviors differently than DNSTT.

---

## Core

### `enabled`

```toml
enabled = true
```

Enable/disable Slipstream phase.

---

### `workers`

```toml
workers = 20
```

Recommended: 5–50

---

### `domain`

```toml
domain = ""
```

Authoritative DNS zone for Slipstream server.

---

### `cert_path`

```toml
cert_path = ""
```

Path to TLS certificate used by Slipstream server.

Required for TLS validation.

---

## Timing

```toml
timeout = 8000
```

Maximum duration for full probe attempt.

---

## Output

```toml
prefix_output = "slipstream_"
```
