 
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

```bash
git clone https://github.com/yourname/brook.git
cd brook
go build -o brook .

![img.png](docment/img.png)

