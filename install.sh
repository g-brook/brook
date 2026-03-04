#!/bin/bash
#
# Copyright © sixh sixh@apache.org
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#===============================================================================
# Brook Client 安装脚本
#
# 根据 README.md 提供的说明自动下载、安装和配置 Brook Client
# GitHub: https://github.com/g-brook/brook
#===============================================================================

set -e

# ─────────────────────────────────────────────
#  颜色 & 样式定义
# ─────────────────────────────────────────────
RESET='\033[0m'
BOLD='\033[1m'
DIM='\033[2m'

BLACK='\033[0;30m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[0;37m'

BG_GREEN='\033[42m'
BG_RED='\033[41m'
BG_YELLOW='\033[43m'
BG_BLUE='\033[44m'
BG_CYAN='\033[46m'

# ─────────────────────────────────────────────
#  全局变量
# ─────────────────────────────────────────────
INSTALL_TYPE=""  # "server" 或 "client"
APP_NAME=""
INSTALL_BASE_DIR=""
REPO_URL="https://github.com/g-brook/brook/releases/latest/download"
SYSTEMD_SERVICE_DIR="/etc/systemd/system"

# 步骤计数
TOTAL_STEPS=6
CURRENT_STEP=0

# ─────────────────────────────────────────────
#  日志函数
# ─────────────────────────────────────────────

# 获取当前时间戳
_ts() {
    date '+%H:%M:%S'
}

# INFO  ✔ 绿色
log_info() {
    echo -e "${DIM}$(_ts)${RESET}  ${BOLD}${GREEN}  INFO ${RESET}  $1"
}

# WARN  ⚠ 黄色
log_warn() {
    echo -e "${DIM}$(_ts)${RESET}  ${BOLD}${YELLOW}  WARN ${RESET}  ${YELLOW}$1${RESET}"
}

# ERROR ✘ 红色
log_error() {
    echo -e "${DIM}$(_ts)${RESET}  ${BOLD}${RED} ERROR ${RESET}  ${RED}$1${RESET}" >&2
}

# DEBUG 灰色（仅 DEBUG=1 时输出）
log_debug() {
    [ "${DEBUG:-0}" = "1" ] && \
    echo -e "${DIM}$(_ts)  DEBUG   $1${RESET}"
}

# 成功完成行
log_done() {
    echo -e "${DIM}$(_ts)${RESET}  ${BOLD}${GREEN}  DONE ${RESET}  ${GREEN}✔  $1${RESET}"
}

# ─────────────────────────────────────────────
#  STEP 步骤标题
# ─────────────────────────────────────────────
log_step() {
    CURRENT_STEP=$(( CURRENT_STEP + 1 ))
    local title="$1"
    local bar=""
    local i=1
    # 进度块：已完成用实心，未完成用空心
    while [ $i -le $TOTAL_STEPS ]; do
        if [ $i -lt $CURRENT_STEP ]; then
            bar="${bar}${GREEN}█${RESET}"
        elif [ $i -eq $CURRENT_STEP ]; then
            bar="${bar}${CYAN}█${RESET}"
        else
            bar="${bar}${DIM}░${RESET}"
        fi
        i=$(( i + 1 ))
    done

    echo ""
    echo -e "${BOLD}${BG_BLUE}${WHITE}  STEP ${CURRENT_STEP}/${TOTAL_STEPS}  ${RESET}${BOLD}  ${title}${RESET}"
    echo -e "  ${bar}  ${DIM}Step ${CURRENT_STEP} of ${TOTAL_STEPS}${RESET}"
    echo -e "  ${DIM}────────────────────────────────────────${RESET}"
}

# ─────────────────────────────────────────────
#  Banner
# ─────────────────────────────────────────────
show_banner() {
    echo -e "${BOLD}${CYAN}      __                     __${RESET}"
    echo -e "${BOLD}${CYAN}      / /__________________  / /__${RESET}"
    echo -e "${BOLD}${CYAN}     / __  / ___/ __ \/ __ \/ //_/ ${RESET}"
    echo -e "${BOLD}${CYAN}    / /_/ / /  / /_/ / /_/ / ,<   ${RESET}"
    echo -e "${BOLD}${CYAN}   /_.___/_/   \____/\____/_/|_|  ${RESET}"
    echo -e "${DIM}──────────────────────────────────────────${RESET}"
    echo -e "${BOLD}          Brook  安装脚本${RESET}"
    echo -e "${DIM}          https://github.com/g-brook/brook${RESET}"
    echo ""
}

# ─────────────────────────────────────────────
#  选择安装类型
# ─────────────────────────────────────────────
select_install_type() {
    echo ""
    echo -e "${BOLD}请选择安装类型：${RESET}"
    echo ""
    echo -e "  ${CYAN}1${RESET}) 安装服务端 ${DIM}(brook-sev)${RESET}"
    echo -e "  ${CYAN}2${RESET}) 安装客户端 ${DIM}(brook-cli)${RESET}"
    echo ""
    
    while true; do
        printf "请输入选项 [1-2]: "
        read -r choice
        
        case $choice in
            1)
                INSTALL_TYPE="server"
                APP_NAME="brook-sev"
                INSTALL_BASE_DIR="$HOME/$APP_NAME"
                log_info "已选择：安装服务端 (brook-sev)"
                break
                ;;
            2)
                INSTALL_TYPE="client"
                APP_NAME="brook-cli"
                INSTALL_BASE_DIR="$HOME/$APP_NAME"
                log_info "已选择：安装客户端 (brook-cli)"
                break
                ;;
            *)
                log_warn "无效选项，请输入 1 或 2"
                ;;
        esac
    done
    
    echo ""
}

# ─────────────────────────────────────────────
#  帮助信息
# ─────────────────────────────────────────────
show_help() {
    cat << EOF

${BOLD}Brook 安装脚本${RESET} — 根据 README.md 自动下载安装

${BOLD}用法：${RESET}
  $0 [选项]

${BOLD}选项：${RESET}
  ${CYAN}-h, --help${RESET}           显示帮助信息
  ${CYAN}-u, --uninstall${RESET}      卸载 Brook
  ${CYAN}-r, --run${RESET}            安装完成后立即启动服务
  ${CYAN}-p, --path PATH${RESET}      指定安装目录 (默认：\$HOME/brook-*)
  ${CYAN}-s, --server HOST${RESET}    设置服务端地址 (仅客户端)
  ${CYAN}-t, --token TOKEN${RESET}    设置认证 Token
  
${BOLD}安装类型选择：${RESET}
  ${CYAN}--server${RESET}             安装服务端 (brook-sev)
  ${CYAN}--client${RESET}             安装客户端 (brook-cli)
  ${DIM}(不指定则交互式选择)${RESET}

${BOLD}示例：${RESET}
  $0 --server                        # 安装服务端
  $0 --client                        # 安装客户端
  $0 --client -s example.com -t token123  # 安装客户端并配置
  $0 -r --server                     # 安装服务端并启动
  $0 -u                              # 卸载

EOF
}

# ─────────────────────────────────────────────
#  STEP 1：检测操作系统
# ─────────────────────────────────────────────
detect_os() {
    log_step "检测操作系统"

    if [ "$(uname)" = "Darwin" ]; then
        OS="macOS"
        OS_VERSION=$(sw_vers -productVersion)
        log_info "操作系统  : macOS ${OS_VERSION}"

        if command -v sysctl > /dev/null 2>&1 && \
           [ "$(sysctl -n machdep.cpu.brand_string 2>/dev/null | grep -c 'Apple')" -gt 0 ]; then
            ARCH="ARM64"
            FILE_DESC="macOS-ARM64(Apple-M)"
            log_info "处理器架构: Apple Silicon (M 系列)"
        else
            ARCH="Intel"
            FILE_DESC="macOS-Intel"
            log_info "处理器架构: Intel x86_64"
        fi

    elif [ "$(uname)" = "Linux" ]; then
        OS="Linux"
        if [ -f /etc/os-release ]; then
            # shellcheck source=/dev/null
            . /etc/os-release
            OS_NAME="${NAME:-Linux}"
            OS_VERSION="${VERSION_ID:-unknown}"
            log_info "操作系统  : ${OS_NAME} ${OS_VERSION}"
        else
            OS_NAME="Linux"
            log_info "操作系统  : Linux (未知发行版)"
        fi

        ARCH_RAW=$(uname -m)
        case $ARCH_RAW in
            x86_64)
                ARCH="x86_64"
                FILE_DESC="Linux-x86_64(amd64)"
                log_info "处理器架构: x86_64 (amd64)"
                ;;
            aarch64|arm64)
                ARCH="arm64"
                FILE_DESC="Linux-arm64"
                log_info "处理器架构: ARM64"
                ;;
            *)
                log_error "不支持的 CPU 架构：${ARCH_RAW}"
                exit 1
                ;;
        esac

        if command -v systemctl > /dev/null 2>&1; then
            HAS_SYSTEMD=true
            log_info "服务管理  : systemd ✔"
        else
            HAS_SYSTEMD=false
            log_warn "服务管理  : 未检测到 systemd，将跳过服务注册"
        fi

    elif [ "$OSTYPE" = "msys" ] || [ "$OSTYPE" = "cygwin" ]; then
        OS="Windows"
        ARCH="x86_64"
        FILE_DESC="Windows-x86_64"
        INSTALL_BASE_DIR="$HOME/$APP_NAME"
        log_info "操作系统  : Windows (Git Bash / WSL)"
        log_info "处理器架构: x86_64"
    else
        log_error "不支持的操作系统：$(uname)"
        exit 1
    fi
    
    # 根据安装类型设置下载文件名
    set_download_filename

    log_done "操作系统检测完成  [${BOLD}${OS}${RESET}${GREEN} / ${BOLD}${ARCH}${RESET}${GREEN}]"
}

# ─────────────────────────────────────────────
#  设置下载文件名
# ─────────────────────────────────────────────
set_download_filename() {
    if [ "$INSTALL_TYPE" = "server" ]; then
        # 服务端文件名
        case "${OS}-${ARCH}" in
            "Linux-x86_64")
                DOWNLOAD_FILE="brook-sev_Linux-x86_64.amd64.tar.gz"
                ;;
            "Linux-arm64")
                DOWNLOAD_FILE="brook-sev_Linux-arm64.tar.gz"
                ;;
            "macOS-ARM64")
                DOWNLOAD_FILE="brook-sev_macOS-ARM64.Apple-M.tar.gz"
                ;;
            "macOS-Intel")
                DOWNLOAD_FILE="brook-sev_macOS-Intel.tar.gz"
                ;;
            "Windows-x86_64")
                DOWNLOAD_FILE="brook-sev_Windows-x86_64.tar.gz"
                ;;
            *)
                log_error "不支持的平台组合：${OS}-${ARCH}"
                exit 1
                ;;
        esac
    else
        # 客户端文件名
        case "${OS}-${ARCH}" in
            "Linux-x86_64")
                DOWNLOAD_FILE="brook-cli_Linux-x86_64.amd64.tar.gz"
                ;;
            "Linux-arm64")
                DOWNLOAD_FILE="brook-cli_Linux-arm64.tar.gz"
                ;;
            "macOS-ARM64")
                DOWNLOAD_FILE="brook-cli_macOS-ARM64.Apple-M.tar.gz"
                ;;
            "macOS-Intel")
                DOWNLOAD_FILE="brook-cli_macOS-Intel.tar.gz"
                ;;
            "Windows-x86_64")
                DOWNLOAD_FILE="brook-cli_Windows-x86_64.tar.gz"
                ;;
            *)
                log_error "不支持的平台组合：${OS}-${ARCH}"
                exit 1
                ;;
        esac
    fi
    
    log_info "安装包    : ${DOWNLOAD_FILE}"
}

# ─────────────────────────────────────────────
#  STEP 2：检查系统依赖
# ─────────────────────────────────────────────
check_dependencies() {
    log_step "检查系统依赖"

    local missing_deps=""

    for dep in curl tar; do
        if command -v "$dep" > /dev/null 2>&1; then
            log_info "依赖检查  : ${dep} $(command -v "$dep") ✔"
        else
            log_warn "依赖缺失  : ${dep}"
            missing_deps="${missing_deps} ${dep}"
        fi
    done

    if [ -n "$missing_deps" ]; then
        log_error "缺少必要依赖：${missing_deps}"
        log_info "请使用以下命令安装后重试："
        if [ "$OS" = "macOS" ]; then
            echo -e "  ${CYAN}brew install${missing_deps}${RESET}"
        elif [ "$OS" = "Linux" ]; then
            if command -v apt-get > /dev/null 2>&1; then
                echo -e "  ${CYAN}sudo apt-get update && sudo apt-get install -y${missing_deps}${RESET}"
            elif command -v yum > /dev/null 2>&1; then
                echo -e "  ${CYAN}sudo yum install -y${missing_deps}${RESET}"
            elif command -v apk > /dev/null 2>&1; then
                echo -e "  ${CYAN}sudo apk add${missing_deps}${RESET}"
            fi
        fi
        exit 1
    fi

    log_done "所有依赖检查通过"
}

# ─────────────────────────────────────────────
#  STEP 3：创建安装目录
# ─────────────────────────────────────────────
create_directories() {
    log_step "创建安装目录"

    if [ -d "$INSTALL_BASE_DIR" ]; then
        log_warn "目录已存在，将覆盖安装：${INSTALL_BASE_DIR}"
    fi

    mkdir -p "$INSTALL_BASE_DIR"
    log_info "安装目录  : ${INSTALL_BASE_DIR}"
    log_done "目录创建完成"
}

# ─────────────────────────────────────────────
#  STEP 4：下载 Brook Client
# ─────────────────────────────────────────────
download_file() {
    log_step "下载 Brook Client"

    local DOWNLOAD_URL="${REPO_URL}/${DOWNLOAD_FILE}"
    local TEMP_FILE="/tmp/${DOWNLOAD_FILE}"

    log_info "下载地址  : ${DOWNLOAD_URL}"
    log_info "目标文件  : ${DOWNLOAD_FILE}"
    log_info "临时路径  : ${TEMP_FILE}"
    echo ""

    if curl --fail --location --progress-bar --output "$TEMP_FILE" "$DOWNLOAD_URL"; then
        echo ""
        DOWNLOADED_FILE="$TEMP_FILE"
        log_done "文件下载完成"
    else
        echo ""
        log_error "下载失败，请检查网络连接或访问以下地址手动下载："
        echo -e "  ${CYAN}https://github.com/g-brook/brook/releases${RESET}"
        exit 1
    fi
}

# ─────────────────────────────────────────────
#  STEP 5：解压文件
# ─────────────────────────────────────────────
extract_file() {
    log_step "解压安装包"

    mkdir -p "$INSTALL_BASE_DIR"
    log_info "解压至    : ${INSTALL_BASE_DIR}"

    if tar -xzf "$DOWNLOADED_FILE" -C "$INSTALL_BASE_DIR"; then
        log_done "解压完成"
    else
        log_error "解压失败，安装包可能已损坏"
        exit 1
    fi
}

# ─────────────────────────────────────────────
#  STEP 6：验证安装
# ─────────────────────────────────────────────
verify_installation() {
    log_step "验证安装结果"

    local BINARY_FILE="$INSTALL_BASE_DIR/$APP_NAME"
    [ "$OS" = "Windows" ] && BINARY_FILE="${BINARY_FILE}.exe"

    if [ -x "$BINARY_FILE" ]; then
        log_info "二进制文件: ${BINARY_FILE} ✔"

        local version
        if version=$("$BINARY_FILE" --version 2>/dev/null); then
            log_info "版本信息  : ${version}"
        else
            log_info "版本信息  : (无法获取，程序可能需要配置后运行)"
        fi
    else
        log_error "二进制文件不存在或不可执行：${BINARY_FILE}"
        exit 1
    fi

    log_done "安装验证通过"
    
    # 创建默认配置文件
    create_config
}

# ─────────────────────────────────────────────
#  创建配置文件
# ─────────────────────────────────────────────
create_config() {
    log_step "创建配置文件"
    
    local CONFIG_FILE="$INSTALL_BASE_DIR/client.json"
    [ "$INSTALL_TYPE" = "server" ] && CONFIG_FILE="$INSTALL_BASE_DIR/server.json"
    
    if [ -f "$CONFIG_FILE" ]; then
        log_warn "配置文件已存在：$CONFIG_FILE"
        return
    fi
    
    if [ "$INSTALL_TYPE" = "server" ]; then
        # 服务端配置
        cat > "$CONFIG_FILE" << EOF
{
  "enableWeb": true,
  "webPort": 8000,
  "serverPort": 8909,
  "tunnelPort": 8919,
  "token": "",
  "logger": {
    "logLevel": "info",
    "logPath": "./",
    "outs": "file"
  }
}
EOF
        log_info "配置文件  : ${CONFIG_FILE}"
        log_warn "请编辑配置文件设置 token（如果不使用 Web 模式）"
    else
        # 客户端配置
        local server_host="${SERVER_HOST:-127.0.0.1}"
        local token="${TOKEN:-}"
        
        cat > "$CONFIG_FILE" << EOF
{
  "serverPort": 8909,
  "serverHost": "$server_host",
  "managerPort": 0,
  "token": "$token",
  "pingTime": 10000,
  "tunnels": [
    {
      "type": "tcp",
      "destination": "127.0.0.1:3306",
      "proxyId": "mysql"
    },
    {
      "type": "tcp",
      "destination": "127.0.0.1:5432",
      "proxyId": "postgres"
    },
    {
      "type": "tcp",
      "destination": "127.0.0.1:6379",
      "proxyId": "redis"
    }
  ]
}
EOF
        log_info "配置文件  : ${CONFIG_FILE}"
        if [ -z "$TOKEN" ]; then
            log_warn "请编辑配置文件设置 serverHost 和 token"
        fi
    fi
    
    log_done "配置文件创建完成"
}

# ─────────────────────────────────────────────
#  安装完成摘要
# ─────────────────────────────────────────────
show_usage() {
    echo ""
    echo -e "${BOLD}${BG_GREEN}${BLACK}  ✔  安装完成  ${RESET}"
    echo ""
    
    if [ "$INSTALL_TYPE" = "server" ]; then
        echo -e "${BOLD}  服务端安装信息${RESET}"
    else
        echo -e "${BOLD}  客户端安装信息${RESET}"
    fi
    echo -e "  ${DIM}────────────────────────────────────────${RESET}"
    echo -e "  安装类型  : ${CYAN}${INSTALL_TYPE}${RESET}"
    echo -e "  程序名称  : ${CYAN}${APP_NAME}${RESET}"
    echo -e "  安装目录  : ${CYAN}${INSTALL_BASE_DIR}${RESET}"
    
    if [ "$INSTALL_TYPE" = "server" ]; then
        echo -e "  配置文件  : ${CYAN}${INSTALL_BASE_DIR}/server.json${RESET}"
    else
        echo -e "  配置文件  : ${CYAN}${INSTALL_BASE_DIR}/client.json${RESET}"
    fi
    echo ""

    if [ "$OS" = "Linux" ] && [ "${HAS_SYSTEMD:-false}" = "true" ]; then
        local service_name="brook"
        [ "$INSTALL_TYPE" = "client" ] && service_name="brook-cli"
        
        echo -e "${BOLD}  服务管理命令 (systemd)${RESET}"
        echo -e "  ${DIM}────────────────────────────────────────${RESET}"
        echo -e "  启动服务  : ${CYAN}sudo systemctl start  ${service_name}${RESET}"
        echo -e "  停止服务  : ${CYAN}sudo systemctl stop   ${service_name}${RESET}"
        echo -e "  开机自启  : ${CYAN}sudo systemctl enable ${service_name}${RESET}"
        echo -e "  查看状态  : ${CYAN}sudo systemctl status ${service_name}${RESET}"
        echo -e "  实时日志  : ${CYAN}sudo journalctl -u ${service_name} -f${RESET}"
        echo ""
        echo -e "${BOLD}  手动启动 (备用)${RESET}"
        echo -e "  ${DIM}────────────────────────────────────────${RESET}"
        if [ "$INSTALL_TYPE" = "server" ]; then
            echo -e "  ${CYAN}cd ${INSTALL_BASE_DIR} && ./${APP_NAME} -c ./server.json${RESET}"
        else
            echo -e "  ${CYAN}cd ${INSTALL_BASE_DIR} && ./${APP_NAME} -c ./client.json${RESET}"
        fi
    else
        echo -e "${BOLD}  启动服务${RESET}"
        echo -e "  ${DIM}────────────────────────────────────────${RESET}"
        if [ "$INSTALL_TYPE" = "server" ]; then
            echo -e "  ${CYAN}cd ${INSTALL_BASE_DIR} && ./${APP_NAME} -c ./server.json${RESET}"
        else
            echo -e "  ${CYAN}cd ${INSTALL_BASE_DIR} && ./${APP_NAME} -c ./client.json${RESET}"
        fi
    fi

    echo ""
    echo -e "${BOLD}  下一步操作${RESET}"
    echo -e "  ${DIM}────────────────────────────────────────${RESET}"
    
    if [ "$INSTALL_TYPE" = "server" ]; then
        echo -e "  ${BOLD}1.${RESET} 编辑配置文件："
        echo -e "     ${CYAN}vim ${INSTALL_BASE_DIR}/server.json${RESET}"
        echo -e "     ${DIM}· 设置 token 或通过 Web 界面生成${RESET}"
        echo -e "     ${DIM}· 配置端口（webPort, serverPort, tunnelPort）${RESET}"
        echo -e "  ${BOLD}2.${RESET} 启动服务端"
        echo -e "  ${BOLD}3.${RESET} 访问 Web 管理界面：${CYAN}http://localhost:8000/index${RESET}"
        echo -e "  ${BOLD}4.${RESET} 在管理界面创建隧道配置"
    else
        echo -e "  ${BOLD}1.${RESET} 编辑配置文件："
        echo -e "     ${CYAN}vim ${INSTALL_BASE_DIR}/client.json${RESET}"
        echo -e "     ${DIM}· 修改 serverHost 为服务端地址${RESET}"
        echo -e "     ${DIM}· 修改 token 为服务端生成的 Token${RESET}"
        echo -e "     ${DIM}· 修改 proxyId 与服务端保持一致${RESET}"
        echo -e "  ${BOLD}2.${RESET} 启动客户端"
        echo -e "  ${BOLD}3.${RESET} 测试隧道连接是否正常"
    fi
    
    echo ""
    echo -e "  ${DIM}更多信息：${CYAN}https://github.com/g-brook/brook${RESET}"
    echo ""
}

# ─────────────────────────────────────────────
#  启动服务
# ─────────────────────────────────────────────
start_service() {
    echo ""
    echo -e "${BOLD}${BG_CYAN}${BLACK}  ▶  启动服务  ${RESET}"
    echo ""

    if [ "$OS" = "Linux" ] && [ "${HAS_SYSTEMD:-false}" = "true" ]; then
        log_info "正在通过 systemd 启动 brook-cli ..."
        sudo systemctl start brook-cli
        sleep 2
        if sudo systemctl is-active --quiet brook-cli; then
            log_done "服务已成功启动"
            sudo systemctl status brook-cli --no-pager | sed 's/^/  /'
        else
            log_warn "服务启动失败，请检查："
            echo -e "  ${CYAN}sudo systemctl status brook-cli${RESET}"
        fi
    else
        log_info "请手动启动服务："
        echo -e "  ${CYAN}cd ${INSTALL_BASE_DIR} && ./${APP_NAME} -c ./client.json${RESET}"
    fi
}

# ─────────────────────────────────────────────
#  卸载
# ─────────────────────────────────────────────
uninstall() {
    echo ""
    echo -e "${BOLD}${BG_YELLOW}${BLACK}  ⚠  开始卸载 Brook  ${RESET}"
    echo ""
    
    # 先检测操作系统和类型
    if [ "$OS" = "Linux" ] && [ "${HAS_SYSTEMD:-false}" = "true" ]; then
        local service_name="brook"
        [ -f "$HOME/brook-cli/client.json" ] && service_name="brook-cli"
        
        log_info "停止并禁用 systemd 服务 ..."
        sudo systemctl stop    "$service_name" 2>/dev/null || true
        sudo systemctl disable "$service_name" 2>/dev/null || true
        sudo rm -f "${SYSTEMD_SERVICE_DIR}/${service_name}.service"
        sudo systemctl daemon-reload
        log_done "systemd 服务已移除 (${service_name})"
    fi
    
    # 删除服务端目录
    if [ -d "$HOME/brook-sev" ]; then
        rm -rf "$HOME/brook-sev"
        log_done "已删除服务端目录：$HOME/brook-sev"
    fi
    
    # 删除客户端目录
    if [ -d "$HOME/brook-cli" ]; then
        rm -rf "$HOME/brook-cli"
        log_done "已删除客户端目录：$HOME/brook-cli"
    fi
    
    # 如果指定了自定义路径，也删除
    if [ -n "$INSTALL_BASE_DIR" ] && [ "$INSTALL_BASE_DIR" != "$HOME/brook-sev" ] && \
       [ "$INSTALL_BASE_DIR" != "$HOME/brook-cli" ] && [ -d "$INSTALL_BASE_DIR" ]; then
        rm -rf "$INSTALL_BASE_DIR"
        log_done "已删除自定义安装目录：${INSTALL_BASE_DIR}"
    fi

    echo ""
    log_done "卸载完成"
    log_info "如需彻底清理，可手动删除残留的配置文件和日志文件"
    echo ""
}

clean() {
      rm -rf "$DOWNLOADED_FILE"
      log_debug "已清理临时文件：${DOWNLOADED_FILE}"
}
# ─────────────────────────────────────────────
#  主函数
# ─────────────────────────────────────────────
main() {
    show_banner

    local RUN_AFTER_INSTALL=false
    local SERVER_HOST=""
    local TOKEN=""
    local INSTALL_TYPE_ARG=""

    # 解析命令行参数
    while [ $# -gt 0 ]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -u|--uninstall)
                detect_os
                uninstall
                exit 0
                ;;
            -r|--run)
                RUN_AFTER_INSTALL=true
                shift
                ;;
            -p|--path)
                if [ -z "${2:-}" ]; then
                    log_error "--path 选项需要指定目录路径"
                    exit 1
                fi
                INSTALL_BASE_DIR="$2"
                shift 2
                ;;
            --server)
                INSTALL_TYPE_ARG="server"
                shift
                ;;
            --client)
                INSTALL_TYPE_ARG="client"
                shift
                ;;
            -s|--server-host)
                if [ -z "${2:-}" ]; then
                    log_error "--server-host 选项需要指定服务端地址"
                    exit 1
                fi
                SERVER_HOST="$2"
                shift 2
                ;;
            -t|--token)
                if [ -z "${2:-}" ]; then
                    log_error "--token 选项需要指定 Token 值"
                    exit 1
                fi
                TOKEN="$2"
                shift 2
                ;;
            *)
                log_error "未知选项：$1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定安装类型，则交互式选择
    if [ -z "$INSTALL_TYPE_ARG" ]; then
        select_install_type
    else
        INSTALL_TYPE="$INSTALL_TYPE_ARG"
        APP_NAME="brook-${INSTALL_TYPE}"
        INSTALL_BASE_DIR="${INSTALL_BASE_DIR:-$HOME/$APP_NAME}"
        log_info "已选择：安装${INSTALL_TYPE} (brook-${INSTALL_TYPE})"
    fi

    export SERVER_HOST
    export TOKEN

    # 执行安装步骤
    detect_os
    check_dependencies
    create_directories
    download_file
    extract_file
    verify_installation
    show_usage
    
    if [ "$RUN_AFTER_INSTALL" = "true" ]; then
        start_service
    fi
    clean
}

# 执行主函数
main "$@"
