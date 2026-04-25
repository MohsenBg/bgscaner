# 🚀 bgscan

> Ultra‑fast multi‑protocol scanner with a modular chained engine and interactive BubbleTea terminal UI.  
> Designed for developers and researchers who need speed, flexibility, and a modern scanning experience.

**bgscan** is a high‑performance scanning engine written in Go. It can run **multiple protocols in parallel** and **chain protocol stages** to build advanced detection workflows.  
It features an interactive **BubbleTea UI**, a modular execution pipeline, and an asset‑driven architecture that makes it easy to extend, customize, and integrate with external tools such as **Xray**, **DNSTT**, and **Slipstream**.

Built for performance, clarity, and extensibility.

---

![Go](https://img.shields.io/badge/Go-1.25.4+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20windows%20%7C%20macOS-lightgrey)
![License](https://img.shields.io/badge/license-MIT-blue)
![UI](https://img.shields.io/badge/UI-BubbleTea-ff69b4)
![Status](https://img.shields.io/badge/status-production--ready-brightgreen)

---

![bgscan](./image/bgscan.png)

# 📌 What is bgscan?

**bgscan** is a modular, multi‑protocol scanning engine with:

- ⚡ Concurrent pipeline architecture
- 🧠 Smart protocol detection & fallback
- 🌐 DNS tunneling & advanced DNS querying
- 🔐 Xray core integration
- 🧩 Asset‑driven extensibility
- 🎛 Interactive terminal UI (BubbleTea)
- 📦 Clean, validated config system
- 🗂 Robust result management & crash‑safe merging
- 🖥 Cross‑platform support

Ideal for developers, researchers, and power users who need a flexible scanner with a modern TUI.

---

# ✨ Features

## 🔹 Core Engine

- Worker‑pool–based scanning
- Stage / chain execution model
- Context‑aware cancellation
- Bounded, controllable concurrency
- Retry & fallback strategies
- IPv4 / IPv6 CIDR streaming
- Crash‑safe, atomic result merging

## 🔹 Supported Protocols

| Protocol   | Description                         |
| ---------- | ----------------------------------- |
| ICMP       | Ping‑based detection                |
| TCP        | TCP handshake scan                  |
| HTTP       | HTTP response validation            |
| DNS        | Advanced DNS querying with fallback |
| DNSTT      | DNS tunnel transport                |
| Slipstream | Slipstream probing                  |
| Xray       | Xray outbound verification          |

---

## 🔹 BubbleTea UI

Modern interactive terminal interface:

- Multi‑menu navigation
- Live config viewer
- Real‑time logs
- Status indicators
- Progress bars
- Styled components (Lipgloss)
- Keyboard‑driven operation

---

## 📦 Installation

### Download from Releases

Go to:

```text
https://github.com/MohsenBg/bgscaner/releases
```

Download the archive matching your platform:

| OS               | Binary                                        |
| ---------------- | --------------------------------------------- |
| Linux            | `bgscan-linux-amd64` / `bgscan-linux-arm64`   |
| Windows          | `bgscan-windows-amd64.exe`                    |
| macOS            | `bgscan-darwin-amd64` / `bgscan-darwin-arm64` |
| Android (Termux) | `bgscan-android-amd64`                        |

---

## ▶️ Running bgscan

After downloading a release archive:

### 1️⃣ Extract the archive

Linux / macOS / Termux:

```bash
unzip bgscan-*.zip
```

Windows:

> Right click → **Extract All**

---

### 2️⃣ Enter the extracted directory

```bash
cd bgscan-*
```

---

### 3️⃣ Run the binary

Linux / macOS / Termux:

```bash
./bgscan
```

Windows (PowerShell):

```powershell
.\bgscan.exe
```

---

## ⚙️ Settings

All configuration files are located in the `settings/` directory:

```text
settings/
├── dns_settings.toml
├── general_settings.toml
├── http_settings.toml
├── icmp_settings.toml
├── tcp_settings.toml
├── writer_settings.toml
└── xray_settings.toml
```

Edit these files to adjust **bgscan**’s behavior.

Example:

```bash
nano settings/general_settings.toml
# or
vim settings/tcp_settings.toml
```

Changes are applied the next time `bgscan` runs.

---

## 📂 Asset System

`bgscan` uses external runtime assets stored under `assets/`:

```text
assets/
  xray/
    outbounds/
      *.example
  dnstt-client/
  slipstream-client/
```

### Xray Outbounds

Inside `assets/xray/outbounds/` you will find only `.example` templates packaged with the project.

To add your own Xray outbound:

1. Go to the outbound directory:

   ```bash
   cd assets/xray/outbounds
   ```

2. Copy an `.example` file to a `.json` file:

   ```bash
   cp vmess.example vmess.json
   ```

3. Edit the new `.json` file with your real outbound configuration:

   ```bash
   nano vmess.json
   # or
   vim vmess.json
   ```

4. `bgscan` automatically picks up `.json` outbound files at runtime.

`*.example` files are **templates** –  
you should always create and edit your own `.json` files instead of modifying the templates directly.

---

# 🧠 Engine Architecture

```text
CIDR Streamer
      ↓
   Stage Chain
      ↓
   Worker Pool
      ↓
 Protocol Runner
      ↓
  Result Writer
      ↓
  Atomic Merge
```

---

## 🔹 Concurrency Model

- Bounded worker pool
- Context‑based cancellation
- Graceful shutdown
- Channel‑based streaming between stages
- Safe, batched result flushing

---

# 📊 Result System

- Asynchronous writer
- Buffered batching
- Atomic file replace
- `fsync`‑safe writes
- Duplicate filtering
- Structured output by protocol

Supported result types include:

- ICMP
- TCP
- HTTP
- DNS
- DNSTT
- Slipstream
- Xray
- Custom (reserved for future extensions)

---

# 🛠 Build (For Developers)

## Requirements

- Go **1.25.4+**
- Git
- Bash (Linux / macOS / WSL)

---

## Clone the Repository

```bash
git clone https://github.com/MohsenBg/bgscaner.git
cd bgscaner
```

---

## Download Required Assets

`bgscan` depends on external runtime binaries:

- `xray`
- `dnstt-client`
- `slipstream-client`

Download them from the official releases page or their upstream projects, then place them under `assets/` as described below.

---

## 📂 Asset Directory Structure

Place the binaries inside the `assets/` directory with the following structure:

```text
assets/
  xray/
    xray-linux-amd64/
      xray
    xray-darwin-arm64/
      xray
    ...
  dnstt-client/
    dnstt-client-linux-amd64/
      dnstt-client
    ...
  slipstream-client/
    slipstream-client-linux-amd64/
      slipstream-client
    ...
```

Folder names must follow:

```text
<name>-<GOOS>-<GOARCH>
```

Examples:

```text
xray-linux-amd64
dnstt-client-windows-amd64
slipstream-client-darwin-arm64
```

On Windows, the binary filename must end with `.exe`.

---

## 🔨 Build

### Build for Current Platform

```bash
./build.sh
```

The script will:

- Detect your OS/ARCH
- Verify required assets
- Inject version info from Git
- Copy settings and assets
- Generate a SHA256 checksum
- Write releases into:

```text
dist/<version>/
```

---

### Build for All Supported Targets

```bash
./build.sh --all
```

⚠️ When using `--all`, you must provide assets for **every supported platform**:

```text
linux/amd64
linux/arm64
windows/amd64
darwin/amd64
darwin/arm64
android/arm64
```

If any required binary is missing, the build fails with an error.

---

## ✅ Output Example

```text
dist/
  v1.0.0/
    bgscan-linux-amd64/
    bgscan-darwin-arm64/
    bgscan-windows-amd64/
    ...
```

Each platform folder contains:

```text
bgscan (or bgscan.exe)
settings/
assets/
ips/
<checksum file>
```

---

# 🏗 Project Structure

```text
.
├── assets/                     # External runtime binaries + Xray configs
│   ├── dnstt-client/
│   ├── slipstream-client/
│   └── xray/
│       ├── configs/
│       └── outbounds/
│           ├── *.json          # User outbounds
│           └── *.example       # Templates
│
├── cmd/
│   └── bgscan/
│       └── main.go             # Entry point
│
├── internal/
│   ├── core/
│   │   ├── config/             # TOML config loader + defaults
│   │   ├── dns/                # DNS logic (DNSTT, Slipstream, resolvers)
│   │   ├── filemanager/        # CSV/JSON/TXT loaders
│   │   ├── ip/                 # IP utilities (parse/expand)
│   │   ├── iplist/             # IP list registry + loaders
│   │   ├── process/            # Cross‑platform process manager
│   │   ├── result/             # Result writer / merger
│   │   ├── scanner/            # Multi‑stage scanning engine
│   │   └── xray/               # Xray outbound runner / wrapper
│   │
│   ├── logger/                 # Core logger + debug + UI logging
│   ├── startup/                # Startup / health checks
│   └── ui/                     # BubbleTea TUI
│       ├── components/
│       ├── menus/
│       ├── scanner/
│       ├── main/
│       ├── shared/
│       └── theme/
│
├── ips/                        # IP ranges + default templates
│
├── settings/                   # User‑editable TOML configs
│   ├── dns_settings.toml
│   ├── general_settings.toml
│   ├── http_settings.toml
│   ├── icmp_settings.toml
│   ├── tcp_settings.toml
│   ├── writer_settings.toml
│   └── xray_settings.toml
│   # *.default templates exist in source but are not shipped in releases
│
├── build.sh                    # Build script (single or multi‑platform)
├── go.mod
├── go.sum
└── README.md
```

---

# 🔐 Safety & Legal Notice

**bgscan** is intended for:

- Security research
- Educational use
- Authorized infrastructure testing

You are solely responsible for how you use this tool.  
Do **not** scan systems you do not own or do not have explicit permission to test.

---

# 💰 Support & Donation

If this project is useful to you, consider supporting its development.

### 🟡 Bitcoin (BTC)

```text
bc1q3c7cu36faxddjwc3h99k0vt82nj2m9t6u7tdfj
```

### 🟢 USDT (BEP20 – BNB Smart Chain)

```text
0x2ea5A8558B4250cCBF147b2E2501B086700f184A
```

### 🟡 BNB (BEP20)

```text
0x2ea5A8558B4250cCBF147b2E2501B086700f184A
```

### 🔵 Ethereum (ERC20)

```text
0x2ea5A8558B4250cCBF147b2E2501B086700f184A
```

### 🔴 TRON (TRX)

```text
TVxmGjLfyDL3ArbdWk9F8Za24EDm1CHMF4
```

### 🟣 TON

```text
UQDpsu6VBCbl31-LLKcAX8CUCD6BHzzVoHoM2clFJBsct8rq
```

---

# 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Open a pull request

Guidelines:

- Write idiomatic Go
- Add GoDoc for exported types and functions
- Avoid decorative comments / banners
- Keep concurrency safe and clearly documented

---

# 📜 License

Released under the **MIT License**.

---

# 👨‍💻 Author

eveloped and maintained by **MohsenBg**.
