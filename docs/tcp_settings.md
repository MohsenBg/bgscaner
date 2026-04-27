# TCP Scanner Configuration

**File:** `settings/tcp_settings.toml`  
**Module:** TCP Connect Scanner

This module performs **TCP connect() probes** to determine whether a target IP has a specific port open.  
It works by attempting a standard TCP three‑way handshake with a configurable timeout and retry logic.

---

## `port`

```toml
port = 80
```

**Type:** Integer  
**Default:** `80`

Defines which **TCP port** the scanner attempts to connect to for every IP address.

### Common Examples

| Service | Port |
|---------|------|
| HTTP | 80 |
| HTTPS | 443 |
| SSH | 22 |
| RDP | 3389 |
| FTP | 21 |
| MySQL | 3306 |

### Notes

- Only one port can be scanned per run.
- For multi‑port scanning, invoke the scanner multiple times with different configs or implement a port‑list mode.

---

## `timeout`

```toml
timeout = 3000
```

**Type:** Integer (milliseconds)  
**Default:** `3000`

Defines how long to wait for a TCP handshake to succeed.

### Behavior

- `3000` → Wait up to 3 seconds for connection establishment  
- Too low → May misclassify slow/latency‑heavy servers as offline  
- Too high → Increases overall scan duration

### Recommended Values

| Environment | Suggested Timeout |
|------------|-------------------|
| Local/LAN networks | 300–800 |
| Regional scans | 1000–2000 |
| Global internet scanning | 2000–4000 |

---

## `workers`

```toml
workers = 200
```

**Type:** Integer  
**Default:** `200`

Number of concurrent goroutines performing TCP connect attempts.

### Behavior

Higher values:

- Increase scanning speed  
- Increase CPU load  
- Increase number of simultaneous TCP sockets  
- Risk hitting system limits: `ulimit -n`, ephemeral port exhaustion, SYN backlog, etc.

### OS Notes

Linux may require tuning if `workers > 1000`:

Useful sysctl parameters:

```
net.ipv4.ip_local_port_range
net.ipv4.tcp_fin_timeout
net.core.somaxconn
net.ipv4.tcp_tw_reuse
ulimit -n
```

### Recommended Worker Counts

| Scan Type | Worker Count |
|-----------|--------------|
| Safe / conservative | 100–200 |
| High‑speed / optimized | 500–1500 |
| Very large ranges | 1500–3000 (requires OS tuning) |

---

## `prefix_output`

```toml
prefix_output = "tcp_"
```

**Type:** String  
**Default:** `"tcp_"`

Specifies the prefix added to the result file.  
Useful when multiple scanners write to the same directory.

### Example Outputs

```
tcp_20260427_210734.csv.txt
```

---

## `shuffle_ips`

```toml
shuffle_ips = true
```

**Type:** Boolean  
**Default:** `true`

Randomizes the order of scan targets.

### Benefits

- Avoids linear scanning patterns
- Reduces triggering firewall/rate‑limit rules
- Spreads load across the target ranges
- Provides more uniform traffic distribution

### When to Disable

- Debugging  
- Reproducible testing / deterministic scans  

---

## `tries`

```toml
tries = 1
```

**Type:** Integer  
**Default:** `1`

Number of retry attempts if the TCP connection fails.

### Behavior

| Value | Meaning |
|-------|---------|
| `1` | Fastest mode, lowest reliability |
| `2–3` | Better reliability on unstable networks |

### Impact

Higher values:

- Increase chance of detecting open ports under packet loss  
- Increase overall scan duration  
- Increase socket churn

### Recommended Values

| Environment | Suggested Tries |
|------------|-----------------|
| Standard scans | 1 |
| Unstable networks | 2 |
| Critical accuracy | 3 |

