
#  <img src="document/svg.png" alt="svg.png" style="zoom:80%;" />   Brook

**Brook** 是一款跨平台（Linux / macOS / Windows）的高性能网络隧道与代理工具，使用 Go 语言编写。  
它支持 **TCP、UDP、HTTP(S)** 等多种隧道传输方式，兼容 **SSH、HTTP、REDIS、MySQL、WebSocket** 等主流协议。  
Brook 提供直观的 **可视化管理界面**，让用户能够轻松配置和监控连接，实现安全、高效的网络通信。



## 🚀 功能特性

- ✅ 支持 **TCP / UDP / HTTP(S)** 隧道
- ✅ 支持多种协议：**SSH、HTTP(S)、MySQL、Redis、WebSocket等**
- ✅ 提供 **可视化界面**，支持一键配置与状态监控
- ✅ 配置简单，配置文件（`client.json`,`server.json`）
- ✅ 支持超时配置、自动重连与日志输出
- ✅ 轻量高效，跨平台运行

---

## ⚙️ 快速开始

### 🧩下载与安装

你可以从 [GitHub Releases](https://github.com/g-brook/brook/releases) 页面下载适合你系统的二进制包。

| 平台 | 架构 | 文件名                           | 类型                        | 下载地址                                                                                                    |
|------|------|------------------------------------|------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
| 🐧 Linux | amd64(X86_64) | `brook-sev_Linux-x86_64(amd64).tar.gz` | Server | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-x86_64(amd64).tar.gz)    |
| 🐧 Linux | arm64 | `brook-sev_Linux-arm64.tar.gz` | Server | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Linux-arm64.tar.gz)            |
| 🍎 macOS | arm64 (Apple M) | `brook-sev_macOS-ARM64(Apple M).tar.gz` | Server | [下载](`https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-ARM64(Apple-M).tar.gz`) |
| 🍎 macOS | Intel | ` brook-sev_macOS-Intel.tar.gz` | Server | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-sev_macOS-Intel.tar.gz)            |
| 🪟 Windows | amd64(X86_64) | `brook-sev_Windows-x86_64.tar.gz` | Server | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-sev_Windows-x86_64.tar.gz)         |


| 平台 | 架构 | 文件名                            | 类型     | 下载地址                                                                                 |
|------|------|--------------------------------|--------|------|
| 🐧 Linux | amd64(X86_64) | `brook-cli_Linux-x86_64(amd64).tar.gz` | Client | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-x86_64(amd64).tar.gz) |
| 🐧 Linux | arm64 | `brook-cli_Linux-arm64.tar.gz` | Client | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Linux-arm64.tar.gz) |
| 🍎 macOS | arm64 (Apple M) | `brook-cli_macOS-ARM64(Apple M).tar.gz` | Client | [下载](`https://github.com/g-brook/brook/releases/latest/download/brook-cli_macOS-ARM64(Apple-M).tar.gz`) |
| 🍎 macOS | Intel | ` brook-cli_macOS-Intel.tar.gz`   | Client | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-cli_macOS-Intel.tar.gz) |
| 🪟 Windows | amd64(X86_64) | `brook-cli_Windows-x86_64.tar.gz` | Client | [下载](https://github.com/g-brook/brook/releases/latest/download/brook-cli_Windows-x86_64.tar.gz) |


> 🔄 以上链接会自动指向最新版本（`/latest/download/`）。  
> 你也可以进入 [Releases 页面](https://github.com/g-brook/brook/releases) 查看历史版本。


### 🖥️ 服务端运行示例

**1、解压下载的服务器运行包**

```sh
tar -czvf /path/to/archive.tar.gz /path/to/brook
```

**2、更新服务器配置**

* 更新server.json文件

```json
{
  "enableWeb": true, //是否启用web管理界面
  "webPort": 8000, //管理界面的端口,默认8000 端口4000~9000之间
  "serverPort": 8909, //服务管理端口，默认:8909，端口4000~9000之间 
  "tunnelPort": 8919, //隧道端口, 默认：serverPort+10
  "logger": {
    "logLevel": "info",
    "logPath": "./",
    "outs": "file"
  }
}
```

* 更多配置声明,参考：

**3、运行服务**

```sh
./brook-srv
```

**4、界面配置**

访问Web管理界面：https://localhost:8000/index

* 首次登录需要进行初始化：

<img src="document/img_1.png" alt="初始化　" style="zoom:60%;" />

* 使用初始化设置的账号与密码进行登录

<img src="document/img_2.png" alt="初始化　" style="zoom:60%;" />

* 初始化Token

  <img src="document/img_3.png" alt="初始化　" style="zoom:60%;" />

* 设置通道信息(如果是Http隧道，还需要进行Web设置)

  <img src="document/img_4.png" alt="初始化　" style="zoom:60%;" />

### 🖥️ 客户端运行示例

**1、解压下载的客户端运行包**



**客户端配置：**

```sh
tar -czvf /path/to/archive.tar.gz /path/to/brook
./brook-cli
```