# bgscan — Fast IP Scanner

`bgscan` is a fast, lightweight, and interactive tool for scanning lists of IPs using **TCP**, **ICMP**, or **HTTP**.  
It’s distributed as a precompiled executable—**no Go install or build required**.

Just download and run.

[bgscan](./image/bgscan.png)

---

### Features

- Supports 3 scan modes:
  - **TCP** (Recommended – no special permissions required)
  - **ICMP** (real ping – requires root privileges or `CAP_NET_RAW` capability)
  - **HTTP** (scan web servers with custom `Host` header)
- Fully interactive CLI interface
- Complete control over:
  - Threads
  - Timeout
  - Port
  - Shuffle IPs
  - Stop after N successes
  - Verbose mode
- Memory‑safe execution for large IP lists
- Save output with custom prefix


## Download & Run (No Build Needed)

Download the executable for your operating system from the Releases page:

Releases:  
https://github.com/MohsenBg/bgscaner/releases

### Linux / macOS

```bash
chmod +x bgscan
./bgscan
```

### Windows

```powershell
bgscan.exe
```

No Go installation or `go build` required.


## Usage

After starting, the app interactively asks for your configuration:

1. **Choose scan type**
   - **TCP**: Fast, no special permissions
   - **ICMP**: Needs root
     - Run as root
     - Or set capability:
       ```bash
       sudo setcap cap_net_raw+ep bgscan
       ```
     - If permission is missing, falls back to TCP automatically.
   - **HTTP**: For scanning web servers

2. **Configure scan settings**
   - Number of threads
   - Timeout per IP
   - Port number
   - IP list file path
   - Shuffle on/off
   - Limit scanned IP count
   - Output file prefix

3. **Automatic scan start**  
   Scan starts immediately once configuration is complete.


## IP List File

By default, the tool reads from:

```
ips.txt
```

Example file format:

```
8.8.8.8
1.1.1.1
192.168.1.1
```

You can specify a different path at runtime.


## HTTP Mode

- Sends HTTP requests
- Supports custom `Host` header
- Useful for CDN and Virtual Host checking


## Output

- Successful IPs are printed to terminal.
- Results are saved to a file in the format:

```
<output_prefix>_success.txt
```

Example:
```
results_success.txt
```


## Internal Structure (for Developers)

```
bgscan/
 ┣ input/     # Interactive input handling
 ┣ pinger/    # TCP / ICMP / HTTP ping logic
 ┣ scanner/   # Scan management and concurrency
 ┣ main.go    # CLI and main program flow
```


## Legal Disclaimer

This tool is intended **only** for use on:
- Your own systems
- Your own networks
- Or with proper authorization

Unauthorized use may violate network laws.

---

## Author

Developed by MohsenBg  
GitHub: https://github.com/MohsenBg


## License

MIT License — free to use, modify, and distribute.
