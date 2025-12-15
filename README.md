# <img src="document/svg.png" alt="svg.png" style="zoom:80%;" /> Brook ![Latest Release](https://img.shields.io/github/v/release/g-brook/brook?label=Latest&style=flat-square)

[English](document/README.en.md)

**Brook** 是一款跨平台（Linux/macOS/Windows）的高性能网络隧道与代理工具，专为内网穿透设计，采用 Go 语言开发。它支持多种传输协议，包括 TCP、UDP、HTTP(S) 和 WebSocket，并兼容 SSH、HTTP、Redis、MySQL 等主流应用协议，同时提供直观的可视化管理界面，方便进行配置和实时监控。

---

## 🌟 功能特性

- ✅ **多协议支持**：TCP / UDP / HTTP(S) / WebSocket 隧道
- 🔧 **广泛兼容性**：兼容 SSH、HTTP(S)、MySQL、Redis 等常见协议
- 🖥️ **可视化界面**：内置 Web 管理面板，支持初始化、配置与状态监控
- ⚙️ **简单易用**：通过 `client.json` 和 `server.json` 快速完成配置，支持自动重连和日志记录
- 🚀 **轻量高效**：资源占用低，适用于各种规模的应用场景
- 🌍 **跨平台部署**：支持主流操作系统，灵活适应不同环境需求

---

## 📦 下载与安装

前往 [GitHub Releases](https://github.com/g-brook/brook/releases) 页面下载适合您系统的预编译包：

### 服务端（Server）

| 平台     | 架构            | 文件名                                         | 类型   | 下载链接                                                                                           |
|----------|------------------|------------------------------------------------|--------|----------------------------------------------------------------------------------------------------|
| Linux    | x86_64 (amd64)   | `brook-sev_Linux-x86_64(amd64).tar.gz`        | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-arm64.tar.gz) |
| Linux    | arm64            | `brook-sev_Linux-arm64.tar.gz`                | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-arm64.tar.gz) |
| macOS    | ARM64 (Apple M)  | `brook-sev_macOS-ARM64(Apple-M).tar.gz`       | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-ARM64.Apple-M.tar.gz) |
| macOS    | Intel            | `brook-sev_macOS-Intel.tar.gz`                | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-Intel.tar.gz) |
| Windows  | x86_64           | `brook-sev_Windows-x86_64.tar.gz`             | Server | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Windows-x86_64.tar.gz) |

### 客户端（Client）

| 平台     | 架构            | 文件名                                         | 类型   | 下载链接                                                                                          |
|----------|------------------|------------------------------------------------|--------|---------------------------------------------------------------------------------------------------|
| Linux    | x86_64 (amd64)   | `brook-cli_Linux-x86_64(amd64).tar.gz`        | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-arm64.tar.gz) |
| Linux    | arm64            | `brook-cli_Linux-arm64.tar.gz`                | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-arm64.tar.gz) |
| macOS    | ARM64 (Apple M)  | `brook-cli_macOS-ARM64(Apple-M).tar.gz`       | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-ARM64.Apple-M.tar.gz) |
| macOS    | Intel            | `brook-cli_macOS-Intel.tar.gz`                | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_macOS-Intel.tar.gz) |
| Windows  | x86_64           | `brook-cli_Windows-x86_64.tar.gz`             | Client | [Download](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Windows-x86_64.tar.gz) |

> 💡 **提示**：上述链接将自动跳转到最新版本。若需查看历史版本，请访问 [Releases 页面](https://github.com/g-brook/brook/releases)。

---

## 🚀 服务端快速上手指南

### 步骤一：解压并进入目录

mkdir -p ./brook-sev && tar -xzf brook-sev_Linux-arm64.tar.gz -C ./brook-sev
cd brook-sev

### 步骤二：编辑 `server.json` 配置文件

{
  "enableWeb": true,
  "webPort": 8000,
  "serverPort": 8909,
  "tunnelPort": 8919,
  "token": "", // 若不启用 Web 模式，可在此设置静态 token
  "logger": {
    "logLevel": "info",
    "logPath": "./",
    "outs": "file"
  }
}

### 步骤三：启动服务端

./brook-sev -c ./server.json

### 步骤四：访问管理界面

打开浏览器访问 `http://localhost:8000/index`，首次登录需要初始化账户信息。

#### 初始化流程：
1. 设置管理员账号与密码  
2. 登录后生成 Token  

相关截图说明：

<p align="center">
  <img src="document/img_1.png" alt="初始化" width="45%" />
  <img src="document/img_2.png" alt="登录" width="45%" />
  <br/>
  <img src="document/img_3.png" alt="初始化Token" width="45%" />
  <img src="document/img_4.png" alt="设置通道信息" width="45%" />
</p>

---

## 🧩 客户端快速上手指南

### 步骤一：解压并进入目录

mkdir -p ./brook-cli && tar -xzf brook-cli_Linux-arm64.tar.gz -C ./brook-cli
cd brook-cli

### 步骤二：准备 `client.json` 配置文件

您可以从服务端管理界面下载模板并根据实际需求修改：

{
  "serverPort": 8909,
  "serverHost": "127.0.0.1",
  "token": "<在管理后台生成的Token>",
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

### 步骤三：启动客户端

./brook-cli -c ./client.json

### 步骤四：获取模板与标识符

- 获取 `ProxyId` 模板：
  <p align="center"><img src="document/img_8.png" alt="获取ProxyId" width="60%" /></p>

- 获取 `httpId`（仅限 HTTP/HTTPS 隧道）：
  <p align="center"><img src="document/img_9.png" alt="获取HttpId" width="60%" /></p>

---

## ⚙️ 配置详解

### 服务端配置 (`server.json`) 关键字段说明：

| 字段         | 描述                                       |
|--------------|--------------------------------------------|
| `enableWeb`  | 是否启用 Web 管理界面                      |
| `webPort`    | Web 管理界面监听端口，默认为 `8000`        |
| `serverPort` | 主控通信端口，默认为 `8909`                |
| `tunnelPort` | 数据隧道监听端口，默认为 `serverPort + 10` |
| `token`      | 非 Web 模式下的静态身份验证令牌            |
| `logger`     | 日志配置对象                               |

### 客户端配置 (`client.json`) 关键字段说明：

| 字段          | 描述                                               |
|---------------|----------------------------------------------------|
| `serverHost`  | 服务端主机地址                                     |
| `serverPort`  | 服务端控制端口                                     |
| `token`       | 来自服务端管理界面的身份验证令牌                   |
| `pingTime`    | 心跳检测时间间隔（单位：毫秒），建议不少于 2000ms  |
| `tunnels`     | 隧道列表                                           |
| `type`        | 隧道类型（tcp / udp / http / websocket）           |
| `destination` | 本地目标地址                                       |
| `proxyId`     | 服务端分配的唯一标识符                             |
| `httpId`      | HTTP/HTTPS 隧道专用 ID，必须和服务端一致           |

### CLI 启动命令

#### 服务端启动方式：

./brook-sev -c ./server.json
# 或者使用长参数形式
./brook-sev --configs ./server.json

#### 客户端启动方式：

./brook-cli -c ./client.json
# 或者使用长参数形式
./brook-cli --configs ./client.json

##### Linux 系统后台运行（需要 systemd 支持）：

sudo ./brook-cli start
# 查看帮助了解更多命令选项
./brook-cli help

---

## 🔨 从源码构建

### 前端构建（用于服务端 UI 页面）

cd portal/server/
npm install
npm run build

### 服务端构建

cd server/
bash gobuild.sh
# 根据提示选择平台和架构
# 输出路径：server/dist/brook-sev_*.tar.gz

### 客户端构建

cd client/
bash gobuild.sh
# 输出路径：client/dist/brook-cli_*.tar.gz

---

## ❓ 常见问题解答（FAQ）

- **解压失败？**
  > 使用正确的命令：`tar -xzf <file>.tar.gz`，避免误用打包命令 `tar -czvf`。

- **连接不上服务端？**
  > 请检查 `serverHost`、`serverPort` 和 `token` 是否正确；确认服务端是否正常运行且防火墙允许对应端口访问。

- **HTTP/HTTPS 隧道异常？**
  > 客户端中的 `httpId` 必须与服务端 Web 设置完全匹配。

- **端口冲突？**
  > 如果默认端口（如 8000/8909/8919）已被占用，可以在配置中更换其他空闲端口。

- **如何调试问题？**
  > 查阅 `logger.logPath` 中的日志文件，如有必要，可临时将 `logLevel` 调整为 `debug` 获取更详细的信息。

---

## 🗂️ 项目结构概览

├── server/               # 服务端核心逻辑与 Web 管理界面
├── client/               # 客户端核心逻辑与 CLI 接口
├── common/               # 公共模块（配置解析、日志系统、传输封装等）
├── portal/server/        # 前端管理页面（基于 Vite 构建）
└── document/             # 文档资料、截图和其他辅助资源

如需进一步了解功能细节或扩展协议支持，请参阅各目录下的源代码及注释。
