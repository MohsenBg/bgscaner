# General Scanner Configuration

**File:** `settings/general_settings.toml`  
**Module:** Core Execution & Pipeline Control

This configuration file controls the **global behavior of the scanner engine**, including:

- Progress reporting
- Scan limits
- Debug verbosity
- Multi-stage execution strategy
- Inter-stage buffering

These settings directly affect **performance, memory usage, and execution flow** across all scanner modules.

---

## Core Settings

### `status_interval`

```toml
status_interval = 1000
```

**Type:** Integer (milliseconds)  
**Default:** `1000`

Defines how frequently the scanner sends status updates to the UI or progress handler.

#### Behavior

- `1000` → updates every **1 second**
- `5000` → updates every **5 seconds**

Lower values:
- More real‑time progress visibility
- Slightly higher CPU overhead

Higher values:
- Lower overhead
- Slower UI refresh

#### Recommended Values

| Environment | Suggested Value |
|------------|-----------------|
| Local testing | 500–1000 |
| Large-scale scans | 1000–3000 |

---

### `stop_after_found`

```toml
stop_after_found = 0
```

**Type:** Integer  
**Default:** `0` (disabled)

Stops the scan after a defined number of **successful results**.

A successful result means the target passed the scan conditions of the **final stage**.

#### Behavior

| Value | Meaning |
|------|--------|
| `0` | Disabled (scan all IPs) |
| `N` | Stop after **N successful results** |

#### Limitation

In **chain-mode scanning**:

- This limit applies **only to the final stage**
- Intermediate stages do not enforce the limit
- Full chain-limit support is not implemented yet

#### Use Cases

- Sampling valid targets
- Fast result discovery
- Early termination during competitive scans

---

### `max_ips_to_test`

```toml
max_ips_to_test = 0
```

**Type:** Integer  
**Default:** `0` (unlimited)

Limits how many IP addresses are read from the input source.

#### Behavior

| Value | Meaning |
|------|--------|
| `0` | No limit |
| `N` | Process only the first **N IPs** |

#### Common Use Cases

- Testing configuration before running a full scan
- Sampling very large IP lists
- Limiting resource usage in CI pipelines

---

## Chain Execution Mode

Chain mode determines how **multiple scanning stages interact**.

Example pipeline:

```
ICMP → TCP → HTTP → XRAY
```

---

### `chain_mode`

```toml
chain_mode = "parallel"
```

**Type:** String  
**Default:** `"simple"`

Available modes:

---

#### `simple` (Sequential Mode)

Execution model:

1. Stage 1 completes
2. Results written to disk
3. Stage 2 reads the output file
4. Process repeats for the next stage

Pros:

- Lowest RAM usage
- Predictable behavior
- Easy debugging

Cons:

- Slowest mode
- Heavy disk I/O

Recommended for:

- Low-memory systems
- Debugging
- Small datasets

---

#### `parallel` (Streaming Mode)

Execution model:

- All stages run simultaneously
- Results are streamed via buffered channels
- Fully concurrent pipeline

Pros:

- Maximum throughput
- Real-time filtering
- No disk I/O between stages

Cons:

- Higher RAM usage
- More complex concurrency behavior

Recommended for:

- Large-scale scans
- High-performance environments
- Real-time scanning pipelines

---

#### `pipeline` (Batch Hybrid Mode)

Execution model:

- IPs are processed in **batches**
- Each batch flows sequentially through all stages
- Multiple batches may overlap

Pros:

- Balanced RAM usage
- More predictable memory behavior than `parallel`
- Better for uneven stage performance

Recommended for:

- Mixed-cost scanning stages
- Controlled-memory environments
- Medium-scale deployments

---

## Channel Buffer Control

### `channel_buffer_multiple`

```toml
channel_buffer_multiple = 5
```

**Type:** Integer  
**Default:** `5`

Controls how many tasks each stage can buffer from the previous stage.

### Capacity Formula

```
channel_capacity = worker_count × channel_buffer_multiple
```

Example:

```
workers = 30
channel_buffer_multiple = 5
capacity = 150
```

The stage can buffer **150 tasks** before blocking the upstream stage.

---

### Why This Matters

If the buffer becomes full:

- The upstream stage blocks
- Backpressure propagates through the pipeline
- Overall throughput may decrease

If the buffer is too large:

- Memory usage increases
- Large bursts of tasks may accumulate

---

### Recommended Values

The recommended capacity depends on the expected scan size and throughput.

| Scan Size | Recommended Channel Capacity (workers × multiple) | Description |
|-----------|---------------------------------------------------|-------------|
| Small | ~1,000 buffered tasks | Suitable for debugging or small scans |
| Medium | ~10,000 buffered tasks | Balanced throughput and memory usage |
| High Throughput | 50,000–100,000 buffered tasks | Recommended for large-scale parallel scans |

---

### Formula Reminder

```
total_capacity = worker_count × channel_buffer_multiple
```

