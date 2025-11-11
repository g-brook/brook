 
#  ![svg.png](docment/svg.png)   Brook

**Brook** 是一款跨平台（Linux / macOS / Windows）的高性能网络隧道与代理工具，使用 Go 语言编写。  
它支持 **TCP、UDP、HTTP(S)** 等多种隧道传输方式，兼容 **SSH、HTTP、REDIS、MYSQL、WebSocket** 等主流协议。  
Brook 提供直观的 **可视化管理界面**，让用户能够轻松配置和监控连接，实现安全、高效的网络通信。

---

## 🚀 功能特性

- ✅ 支持 **TCP / UDP / HTTP(S)** 隧道
- ✅ 支持多种协议：**SSH、HTTP、MYSQL、Redis、WebSocket**
- ✅ 提供 **可视化界面**，支持一键配置与状态监控
- ✅ 配置简单，配置文件（`client.json`,`server.json`）
- ✅ 支持超时配置、自动重连与日志输出
- ✅ 轻量高效，跨平台运行

---

## ⚙️ 快速开始

### 1️⃣ 下载与安装

🧩 下载与安装

你可以从 [GitHub Releases](https://github.com/yourname/brook/releases) 页面下载适合你系统的二进制包。

| 平台 | 架构 | 文件名 | 下载地址                                                                                     |
|------|------|---------|------------------------------------------------------------------------------------------|
| 🐧 Linux | amd64 | `brook-linux-amd64.tar.gz` | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-linux-amd64.tar.gz) |
| 🍎 macOS | arm64 | `brook-darwin-arm64.zip` | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-darwin-arm64.zip)  |
| 🪟 Windows | amd64 | `brook-windows-amd64.zip` | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-windows-amd64.zip) |

> 🔄 以上链接会自动指向最新版本（`/latest/download/`）。  
> 你也可以进入 [Releases 页面](https://github.com/yourname/brook/releases) 查看历史版本。