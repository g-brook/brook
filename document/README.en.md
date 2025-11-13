# <img src="svg.png" alt="logo" style="zoom:80%;" /> Brook

Brook is a cross-platform, high-performance network tunneling and proxy toolkit implemented in Go.
It supports a wide range of transport protocols, including TCP, UDP, HTTP(S), and WebSocket, ensuring compatibility with popular application protocols such as SSH, HTTP, Redis, and MySQL.
A built-in web UI simplifies configuration.

## Features

- Supports TCP / UDP / HTTP(S) / WebSocket tunnels
- Compatible with SSH, HTTP(S), MySQL, Redis, etc.
- Visual management UI with initialization, configuration, and status monitoring
- Simple configs (`client.json`, `server.json`), auto-reconnect, and logging
- Lightweight, efficient, and cross-platform

---

## Download & Install

Download binaries from GitHub Releases that match your OS and architecture:

| Platform | Arch | Filename | Type | Link |
|----------|------|----------|------|------|
| Linux | x86_64 | `brook-sev_Linux-x86_64(amd64).tar.gz` | Server | https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-x86_64(amd64).tar.gz |
| Linux | arm64 | `brook-sev_Linux-arm64.tar.gz` | Server | https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-arm64.tar.gz |
| macOS | ARM64 (Apple M) | `brook-sev_macOS-ARM64(Apple-M).tar.gz` | Server | https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-ARM64(Apple-M).tar.gz |
| macOS | Intel | `brook-sev_macOS-Intel.tar.gz` | Server | https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-Intel.tar.gz |
| Windows | x86_64 | `brook-sev_Windows-x86_64.tar.gz` | Server | https://github.com/g-brook/brook/releases/latest/download/brook-sev_Windows-x86_64.tar.gz |

| Platform | Arch | Filename | Type | Link |
|----------|------|----------|------|------|
| Linux | x86_64 | `brook-cli_Linux-x86_64(amd64).tar.gz` | Client | https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-x86_64(amd64).tar.gz |
| Linux | arm64 | `brook-cli_Linux-arm64.tar.gz` | Client | https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-arm64.tar.gz |
| macOS | ARM64 (Apple M) | `brook-cli_macOS-ARM64(Apple-M).tar.gz` | Client | https://github.com/g-brook/brook/releases/latest/download/brook-cli_macOS-ARM64(Apple-M).tar.gz |
| macOS | Intel | `brook-cli_macOS-Intel.tar.gz` | Client | https://github.com/g-brook/brook/releases/latest/download/brook-cli_macOS-Intel.tar.gz |
| Windows | x86_64 | `brook-cli_Windows-x86_64.tar.gz` | Client | https://github.com/g-brook/brook/releases/latest/download/brook-cli_Windows-x86_64.tar.gz |

Tip: Links point to the latest version via `/latest/download/`. See the Releases page for older versions.

---

## Server Quick Start

1) Extract the package and enter the directory

```sh
mkdir -p ./brook-sev && tar -xzf brook-sev_Linux-arm64.tar.gz -C ./brook-sev
cd brook-sev
```

2) Edit `server.json`

```json
{
  "enableWeb": true,
  "webPort": 8000,
  "serverPort": 8909,
  "tunnelPort": 8919,
  "token": "", // set a static token here when Web is disabled
  "logger": { "logLevel": "info", "logPath": "./", "outs": "file" }
}
```

3) Start the server

```sh
./brook-sev -c ./server.json
```

4) Open the management UI

- Visit `http://localhost:8000/index`
- First login requires initialization:
  - Set admin username and password
  - Log in and initialize the Token

<img src="img_1.png" alt="Init" style="zoom:60%;" />
<img src="img_2.png" alt="Login" style="zoom:60%;" />
<img src="img_3.png" alt="Init Token" style="zoom:60%;" />
<img src="img_4.png" alt="Setup Tunnels" style="zoom:60%;" />

---

## Client Quick Start

1) Extract the package and enter the directory

```sh
mkdir -p ./brook-cli && tar -xzf brook-cli_Linux-arm64.tar.gz -C ./brook-cli
cd brook-cli
```

2) Prepare `client.json` (download a template from the server UI and modify)

```json
{
  "serverPort": 8909,
  "serverHost": "127.0.0.1",
  "token": "<Token generated in server UI>",
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

3) Start the client

```sh
./brook-cli -c ./client.json
```

4) Retrieve templates and identifiers

- Get `ProxyId` from the server:
  <img src="img_8.png" alt="Get ProxyId" style="zoom:60%;" />
- Get `httpId` (required for HTTP/HTTPS tunnels):
  <img src="img_9.png" alt="Get HttpId" style="zoom:60%;" />

---

## Configuration Reference

Server `server.json` keys:

- `enableWeb`: enable/disable web UI
- `webPort`: web UI port (default 8000, recommended 4000–9000)
- `serverPort`: control port (default 8909)
- `tunnelPort`: data tunnel port (default `serverPort + 10`)
- `token`: static authentication token when Web UI is disabled
- `logger`: logging configs (`logLevel`, `logPath`, `outs`)

Client `client.json` keys:

- `serverHost`: server address (IP or domain)
- `serverPort`: control port (match server config)
- `token`: token generated in server UI, used for login and tunnel binding
- `pingTime`: heartbeat interval in ms (recommended ≥ 2000)
- `tunnels`: array of tunnel definitions
  - `type`: tunnel type (`tcp` / `udp` / `http` / `websocket`)
  - `destination`: local forwarding target, e.g., `127.0.0.1:8081`
  - `proxyId`: identifier generated when creating a tunnel on the server
  - `httpId`: required for HTTP/HTTPS tunnels, must match the server Web setting

CLI flags:

- Server: `./brook-sev -c ./server.json` or `./brook-sev --configs ./server.json`
- Client: `./brook-cli -c ./client.json` or `./brook-cli --configs ./client.json`

---

## Build from Source

- Server:
  - Go to `server/` and run: `bash gobuild.sh`
  - Choose platform/arch interactively; packages are created under `server/dist/brook-sev_*.tar.gz`

- Client:
  - Go to `client/` and run: `bash gobuild.sh`
  - Packages are created under `client/dist/brook-cli_*.tar.gz`

---

## FAQ

- Wrong extraction command: use `tar -xzf <file>.tar.gz` (not `tar -czvf`, which creates archives).
- Connection issues: verify `serverHost`, `serverPort`, and `token`; ensure the server is running and firewall allows the ports.
- HTTP/HTTPS tunnel: client `httpId` must match the server’s Web settings.
- Port conflicts: if 8000/8909/8919 are occupied, change them in the configs to available ports.
- Logs: check files under `logger.logPath`; set `logLevel` to `debug` if needed.

---

## Project Structure

- `server/`: server application and web management
- `client/`: client application and CLI UI
- `common/`: shared components (configs, logging, transport, protocols, etc.)
- `portal/server/`: web front-end (Vite)
- `document/`: screenshots and resources

For more examples and protocol details, refer to the source code and comments in each module.