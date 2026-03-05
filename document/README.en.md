# <img src="logo.svg" alt="svg.png" width="50" height="50" /> <img src="font-dark.svg" alt="svg.png" width="200" height="50" />  ![Latest Release](https://img.shields.io/github/v/release/g-brook/brook?label=Latest&style=flat-square)

[中文](/README.md)

**Brook** is a high-performance network tunnel and proxy tool designed for intranet penetration. It's cross-platform (Linux/macOS/Windows) and developed in Go. It supports multiple transmission protocols including TCP, UDP, HTTP(S), and WebSocket, while being compatible with mainstream application protocols such as SSH, HTTP, Redis, and MySQL. It also provides an intuitive visual management interface for easy configuration and real-time monitoring.

---

## 🌟 Features

- ✅ **Multi-Protocol Support**: TCP / UDP / HTTP(S) / WebSocket tunnels
- 🔧 **Wide Compatibility**: Compatible with SSH, HTTP(S), MySQL, Redis, and other common protocols
- 🖥️ **Visual Interface**: Built-in web management panel supporting initialization, configuration, and status monitoring
- ⚙️ **Easy to Use**: Quick setup via `client.json` and `server.json`, with auto-reconnection and logging support
- 🚀 **Lightweight & Efficient**: Low resource consumption, suitable for various application scenarios
- 🌍 **Cross-Platform Deployment**: Supports mainstream operating systems, flexibly adapting to different environments

---

## 🌐 Online Installation

```shell
bash -c "$(curl -fsSL https://www.gbrook.cc/install.sh)"
```

## 📦 Download & Installation (Manual)

Download pre-compiled packages for your system from the [GitHub Releases](https://github.com/g-brook/brook/releases) page:

### Server

| Platform | Architecture | Filename | Type | Download Link |
|----------|--------------|----------|------|---------------|
| Linux | x86_64 (amd64) | `brook-sev_Linux-x86_64(amd64).tar.gz` | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-arm64.tar.gz) |
| Linux | arm64 | `brook-sev_Linux-arm64.tar.gz` | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-arm64.tar.gz) |
| macOS | ARM64 (Apple M) | `brook-sev_macOS-ARM64(Apple-M).tar.gz` | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-ARM64.Apple-M.tar.gz) |
| macOS | Intel | `brook-sev_macOS-Intel.tar.gz` | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-Intel.tar.gz) |
| Windows | x86_64 | `brook-sev_Windows-x86_64.tar.gz` | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Windows-x86_64.tar.gz) |

### Client

| Platform | Architecture | Filename | Type | Download Link |
|----------|--------------|----------|------|---------------|
| Linux | x86_64 (amd64) | `brook-cli_Linux-x86_64(amd64).tar.gz` | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-arm64.tar.gz) |
| Linux | arm64 | `brook-cli_Linux-arm64.tar.gz` | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-arm64.tar.gz) |
| macOS | ARM64 (Apple M) | `brook-cli_macOS-ARM64(Apple-M).tar.gz` | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-ARM64.Apple-M.tar.gz) |
| macOS | Intel | `brook-cli_macOS-Intel.tar.gz` | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_macOS-Intel.tar.gz) |
| Windows | x86_64 | `brook-cli_Windows-x86_64.tar.gz` | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Windows-x86_64.tar.gz) |

> 💡 **Note**: These links automatically redirect to the latest version. For historical versions, please visit the [Releases Page](https://github.com/g-brook/brook/releases).

---

## 🚀 Server Quick Start Guide

### Step 1: Extract and Enter Directory

```bash
mkdir -p ./brook-sev && tar -xzf brook-sev_Linux-arm64.tar.gz -C ./brook-sev
cd brook-sev
```

### Step 2: Edit `server.json` Configuration File

```json
{
  "enableWeb": true,
  "webPort": 8000,
  "serverPort": 8909,
  "tunnelPort": 8919,
  "token": "", // If Web mode is not enabled, a static token can be set here
  "logger": {
    "logLevel": "info",
    "logPath": "./",
    "outs": "file"
  }
}
```

### Step 3: Start Server

```bash
./brook-sev -c ./server.json
```

### Step 4: Access Management Interface

Open your browser and visit `http://localhost:8000/index`. The first login requires initializing account information.

#### Initialization Process:
1. Set administrator account and password
2. Generate Token after login

Related screenshots:

<p align="center">
  <img src="document/img_1.png" alt="Initialization" width="45%" />
  <img src="document/img_2.png" alt="Login" width="45%" />
  <br/>
  <img src="document/img_3.png" alt="Initialize Token" width="45%" />
  <img src="document/img_4.png" alt="Set Channel Information" width="45%" />
</p>

---

## 🧩 Client Quick Start Guide

### Step 1: Extract and Enter Directory

```bash
mkdir -p ./brook-cli && tar -xzf brook-cli_Linux-arm64.tar.gz -C ./brook-cli
cd brook-cli
```

### Step 2: Prepare `client.json` Configuration File

You can download the template from the server management interface and modify it according to your actual needs:

```json
{
  "serverPort": 8909,
  "serverHost": "127.0.0.1",
  "token": "<Token generated in the management backend>",
  "pingTime": 2000,
  "tunnels": [
    {
      "type": "udp",
      "destination": "127.0.0.1:9000",
      "proxyId": "333223"
    },
    {
      "type": "http",
      "destination": "127.0.0.1:8081",
      "proxyId": "HttpLocal-2",
      "httpId": "local"
    }
  ]
}
```

### Step 3: Start Client

```bash
./brook-cli -c ./client.json
```

### Step 4: Get Template and Identifiers

- Get `ProxyId` template:
  <p align="center"><img src="document/img_8.png" alt="Get ProxyId" width="60%" /></p>

- Get `httpId` (HTTP/HTTPS tunnels only):
  <p align="center"><img src="document/img_9.png" alt="Get HttpId" width="60%" /></p>

---

## ⚙️ Configuration Details

### Server Configuration (`server.json`) Key Fields:

| Field | Description |
|-------|-------------|
| `enableWeb` | Whether to enable the web management interface |
| `webPort` | Web management interface listening port, default is `8000` |
| `serverPort` | Master control communication port, default is `8909` |
| `tunnelPort` | Data tunnel listening port, default is `serverPort + 10` |
| `token` | Static authentication token in non-Web mode |
| `logger` | Log configuration object |

### Client Configuration (`client.json`) Key Fields:

| Field | Description |
|-------|-------------|
| `serverHost` | Server host address |
| `serverPort` | Server control port |
| `token` | Authentication token from server management interface |
| `pingTime` | Heartbeat detection interval (milliseconds), recommended not less than 2000ms |
| `tunnels` | Tunnel list |
| `type` | Tunnel type (tcp / udp / http / websocket) |
| `destination` | Local target address |
| `proxyId` | Unique identifier assigned by server |
| `httpId` | HTTP/HTTPS tunnel dedicated ID, must match server configuration |

### CLI Startup Commands

#### Server Startup:

```bash
./brook-sev -c ./server.json
# Or using long parameter form
./brook-sev --configs ./server.json
```

#### Client Startup:

```bash
./brook-cli -c ./client.json
# Or using long parameter form
./brook-cli --configs ./client.json
```

##### Background Running on Linux (requires systemd support):

```bash
sudo ./brook-cli start
# Check help for more command options
./brook-cli help
```

---

## ❓ FAQ (Frequently Asked Questions)

- **Extraction failed?**
  > Use the correct command: `tar -xzf <file>.tar.gz`, avoid misusing the packaging command `tar -czvf`.

- **Cannot connect to server?**
  > Please check if `serverHost`, `serverPort`, and `token` are correct; confirm the server is running normally and the firewall allows access to the corresponding ports.

- **HTTP/HTTPS tunnel abnormal?**
  > The `httpId` in the client must exactly match the server's web settings.

- **Port conflict?**
  > If default ports (such as 8000/8909/8919) are occupied, you can change to other free ports in the configuration.

- **How to debug issues?**
  > Check the log files in `logger.logPath`. If necessary, temporarily adjust `logLevel` to `debug` to get more detailed information.

---

## 🗂️ Project Structure Overview

```
├── server/               # Server core logic and web management interface
├── client/               # Client core logic and CLI interface
├── common/               # Common modules (configuration parsing, logging system, transmission encapsulation, etc.)
├── portal/server/        # Frontend management page (built with Vite)
└── document/             # Documentation, screenshots, and other auxiliary resources
```

For further details on functionality or extended protocol support, please refer to the source code and comments in each directory.
