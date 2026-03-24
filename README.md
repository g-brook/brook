<p align="center">
  <img src="document/logo.svg" alt="Brook Logo" width="120" height="120" />
</p>

<p align="center">
  <img src="document/font-dark.svg" alt="Brook" width="260" height="60" />
</p>

<p align="center">
  <strong>High-performance, Cross-platform, Minimal Configuration Intranet Penetration & Proxy Tool</strong>
</p>

<p align="center">
  <a href="https://github.com/g-brook/brook/releases">
    <img src="https://img.shields.io/github/v/release/g-brook/brook?label=Latest&style=flat-square&color=blue" alt="Latest Release" />
  </a>
  <a href="https://github.com/g-brook/brook/stargazers">
    <img src="https://img.shields.io/github/stars/g-brook/brook?style=flat-square&logo=github" alt="Stars" />
  </a>
  <a href="https://github.com/g-brook/brook/network/members">
    <img src="https://img.shields.io/github/forks/g-brook/brook?style=flat-square&logo=github" alt="Forks" />
  </a>
  <a href="https://github.com/g-brook/brook/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/g-brook/brook?style=flat-square&color=orange" alt="License" />
  </a>
  <img src="https://img.shields.io/github/go-mod/go-version/g-brook/brook?style=flat-square&logo=go" alt="Go Version" />
  <a href="https://github.com/g-brook/brook/issues">
    <img src="https://img.shields.io/github/issues/g-brook/brook?style=flat-square&color=red" alt="Issues" />
  </a>
</p>

<p align="center">
  <a href="README.zh-CN.md">中文文档</a> | 
  <a href="document/README.en.md">English (Docs)</a> |
  <a href="https://www.gbrook.cc">Official Website</a> | 
  <a href="#-quick-start">Quick Start</a> | 
  <a href="#-faq">FAQ</a>
</p>

---

**Brook** is a high-performance network tunnel tool designed specifically for intranet penetration, developed in Go. It not only supports multiple transmission protocols (TCP, UDP, HTTP, WebSocket) but also simplifies complex tunnel configurations through an intuitive Web management interface. Whether for developer debugging, exposing intranet services, or building private network channels, Brook is your ideal choice.

## ✨ Key Highlights

- 🚀 **Blazing Fast Performance**: High-concurrency architecture based on Go routines, with low latency and low resource consumption.
- 🛡️ **All-around Compatibility**: Supports SSH, HTTP/HTTPS, MySQL, Redis, RDP, and almost all mainstream application protocols.
- 🎨 **Visual Management**: Built-in modern Web panel for one-click initialization, real-time traffic monitoring, and connection status.
- 🔗 **Versatile Protocols**: Native support for TCP / UDP / HTTP(S) / WebSocket tunnels, easily handling various network environments (including CDN and firewall restrictions).
- 🛠️ **Minimal Configuration**: Only one JSON file needed, with auto-reconnection for worry-free operation.
- 💻 **Cross-platform Support**: Pre-compiled packages for Linux, macOS (Intel/M-series), and Windows (x64/ARM64).

---

## 📸 Interface Preview

<details>
<summary>Click to expand and view management interface screenshots</summary>

| **Initialization Wizard** | **Secure Login** |
|:---:|:---:|
| <img src="document/img_1.png" width="400" /> | <img src="document/img_2.png" width="400" /> |
| **Token Management** | **Tunnel Configuration** |
| <img src="document/img_7.png" width="400" /> | <img src="document/img_4.png" width="400" /> |

</details>

---

## ⚡ Quick Start

### 1. One-click Online Installation (Recommended)
```shell
bash -c "$(curl -fsSL https://www.gbrook.cc/install.sh)"
```

### 2. Manual Server Deployment
1. **Download and Extract**: Download the `brook-sev` for your platform from [GitHub Releases](https://github.com/g-brook/brook/releases).
2. **Prepare Configuration** (`server.json`):
   ```json
   {
     "enableWeb": true,
     "webPort": 8000,
     "serverPort": 8909,
     "tunnelPort": 8919,
     "logger": { "logLevel": "info", "logPath": "./", "outs": "file" }
   }
   ```
3. **Start Service**:
   ```shell
   ./brook-sev -c ./server.json
   ```
4. **Access Panel**: Open your browser and visit `http://your-ip:8000/index` for initialization.

### 3. Client Configuration
1. **Get Token**: Generate it in the Web management backend.
2. **Prepare Configuration** (`client.json`):
   ```json
   {
     "serverHost": "your-server-ip",
     "serverPort": 8909,
     "token": "YOUR_GENERATED_TOKEN",
     "tunnels": [
       { "type": "tcp", "destination": "127.0.0.1:80", "proxyId": "web-proxy-1" }
     ]
   }
   ```
3. **Start Client**:
   ```shell
   ./brook-cli -c ./client.json
   ```

---

## 📥 Resource Download

| Platform | Architecture | Server | Client |
| :--- | :--- | :---: | :---: |
| **Linux** | x86_64 / arm64 | [⬇️ Download](https://github.com/g-brook/brook/releases/latest) | [⬇️ Download](https://github.com/g-brook/brook/releases/latest) |
| **macOS** | Intel / Apple M | [⬇️ Download](https://github.com/g-brook/brook/releases/latest) | [⬇️ Download](https://github.com/g-brook/brook/releases/latest) |
| **Windows** | x64 / ARM64 | [⬇️ Download](https://github.com/g-brook/brook/releases/latest) | [⬇️ Download](https://github.com/g-brook/brook/releases/latest) |

---

## 🛠️ Advanced Development

### Build from Source
```bash
# Frontend Build
cd portal/server/ && npm install && npm run build

# Server/Client Build
cd server/ && bash build.sh
cd client/ && bash build.sh
```

---

## ❓ FAQ

<details>
<summary>How to solve connection timeouts?</summary>
Please ensure that ports 8909 and 8919 on the server side are open in the firewall/security group.
</details>

<details>
<summary>Does it support CDN forwarding?</summary>
Yes, by using WebSocket protocol tunnels, you can implement CDN forwarding with Nginx or Cloudflare.
</details>

<details>
<summary>How to run in the background?</summary>
Linux users can use `systemd` scripts or directly run `sudo ./brook-cli start`.
</details>

---

## 📄 Open Source License
This project is open-sourced under the [Apache License 2.0](LICENSE) agreement.

---

<p align="center">
  <b>If Brook helps you, please give it a ⭐ Star!</b><br/>
  <img src="https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go" alt="Made with Go" />
</p>
