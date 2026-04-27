# Result Writer Configuration

**File:** `settings/writer_settings.toml`  
**Module:** Result Aggregation & Disk Writer

This configuration controls how scan results are **buffered, merged, and written to disk**.

Since high‑speed scanners can produce thousands of results per second, the writer subsystem uses:

- buffered channels
- in‑memory batching
- periodic merge flushing

This design prevents disk I/O from becoming a **bottleneck for the scanning pipeline**.

---

## `merge_flush_interval`

```toml
merge_flush_interval = 2000
```

**Type:** Integer (milliseconds)  
**Default:** `2000`

Defines how often the writer merges newly collected results into the **main result file**.

Instead of writing every result immediately, the writer accumulates results in memory and periodically flushes them.

### Behavior

| Value | Behavior |
|------|--------|
| `500–1000` | Frequent flushes, lower memory usage |
| `2000` | Balanced flushing frequency |
| `5000+` | Fewer disk writes, higher temporary memory usage |

### Performance Impact

Lower values:

- More frequent disk writes
- Lower memory usage
- Slightly more I/O overhead

Higher values:

- Better write performance
- Larger in‑memory result accumulation

### Recommended Values

| Environment | Suggested Value |
|------------|----------------|
| Debug / testing | 500–1000 |
| Standard scans | 2000 |
| Very high throughput scans | 3000–5000 |

---

## `chan_size`

```toml
chan_size = 4096
```

**Type:** Integer  
**Default:** `4096`

Defines the capacity of the internal **result channel** used by scanner workers to send results to the writer goroutine.

### Architecture

```
Scanner Workers
       │
       │  (IPScanResult)
       ▼
Buffered Channel (chan_size)
       │
       ▼
Writer Goroutine
       │
       ▼
Disk
```

### Why This Matters

If the channel becomes full:

- Scanner workers block
- Pipeline throughput decreases

If the channel is large:

- Workers rarely block
- Memory usage increases slightly

### Recommended Values

| Scan Speed | Suggested Size |
|-----------|---------------|
| Small scans | 1024 |
| Medium scans | 4096 |
| High‑speed scans | 8192–16384 |

---

## `batch_size`

```toml
batch_size = 4096
```

**Type:** Integer  
**Default:** `4096`

Defines the **initial capacity of the in‑memory batch** used to accumulate `IPScanResult` entries before writing them to disk.

The writer collects results into batches to reduce disk write operations.

### Behavior

If:

```
batch_size = 4096
```

The writer will allocate space for roughly **4096 results** before resizing the buffer.

### Why Batching Helps

Without batching:

- Each result would trigger a disk write
- I/O overhead would dramatically reduce scan speed

With batching:

- Results are written in larger blocks
- Disk I/O becomes significantly more efficient

### Memory Consideration

Memory usage depends on the size of `IPScanResult`.

Example estimation:

```
IPScanResult ≈ 64 bytes
batch_size = 4096

Memory ≈ 256 KB
```

Even large batch sizes are typically inexpensive.

### Recommended Values

| Scan Scale | Suggested Batch Size |
|-----------|----------------------|
| Small scans | 1024 |
| Medium scans | 4096 |
| High throughput scans | 8192–16384 |

---

## Performance Tuning Tips

For **high‑speed scanning environments**:

```toml
merge_flush_interval = 3000
chan_size = 8192
batch_size = 8192
```

For **low‑memory environments**:

```toml
merge_flush_interval = 1000
chan_size = 2048
batch_size = 2048
```

For **debugging**:

```toml
merge_flush_interval = 500
chan_size = 1024
batch_size = 1024
```

---

```
Workers → Channel → Batch Buffer → Merge → Result File
```

