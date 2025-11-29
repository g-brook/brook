#!/bin/sh

set -e

APP_NAME="brook-cli"
APP_PATH="$(cd "$(dirname "$0")" && pwd)/$APP_NAME"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"

# 检测系统语言
detect_language() {
    lang="${LANG:-en_US.UTF-8}"
    case "$lang" in
        zh*) echo "zh" ;;
        *) echo "en" ;;
    esac
}

SYSTEM_LANG=$(detect_language)

# 获取消息
get_msg() {
    key=$1
    case "$key" in
        MSG_INFO_PREFIX)
            [ "$SYSTEM_LANG" = "zh" ] && echo "[信息]" || echo "[INFO]"
            ;;
        MSG_ERROR_PREFIX)
            [ "$SYSTEM_LANG" = "zh" ] && echo "[错误]" || echo "[ERROR]"
            ;;
        MSG_SUCCESS_PREFIX)
            [ "$SYSTEM_LANG" = "zh" ] && echo "[成功]" || echo "[SUCCESS]"
            ;;
        MSG_NO_ROOT)
            [ "$SYSTEM_LANG" = "zh" ] && echo "请不要以root用户直接运行此脚本,脚本会在需要时使用sudo" || echo "Please do not run this script as root user directly, the script will use sudo when needed"
            ;;
        MSG_NO_SYSTEMD)
            [ "$SYSTEM_LANG" = "zh" ] && echo "systemctl命令未找到,正在尝试安装systemd..." || echo "systemctl command not found, trying to install systemd..."
            ;;
        MSG_INSTALLING_SYSTEMD)
            [ "$SYSTEM_LANG" = "zh" ] && echo "正在安装systemd..." || echo "Installing systemd..."
            ;;
        MSG_SYSTEMD_INSTALLED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "systemd安装成功" || echo "systemd installed successfully"
            ;;
        MSG_SYSTEMD_INSTALL_FAILED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "systemd安装失败,请手动安装" || echo "Failed to install systemd, please install manually"
            ;;
        MSG_UNSUPPORTED_OS)
            [ "$SYSTEM_LANG" = "zh" ] && echo "不支持的操作系统,无法自动安装systemd" || echo "Unsupported OS, cannot auto-install systemd"
            ;;
        MSG_NO_EXECUTABLE)
            [ "$SYSTEM_LANG" = "zh" ] && echo "可执行文件 $APP_NAME 未在当前目录找到" || echo "Executable file $APP_NAME not found in current directory"
            ;;
        MSG_EXECUTABLE_PATH)
            [ "$SYSTEM_LANG" = "zh" ] && echo "期望路径: $APP_PATH" || echo "Expected path: $APP_PATH"
            ;;
        MSG_NOT_EXECUTABLE)
            [ "$SYSTEM_LANG" = "zh" ] && echo "文件 $APP_PATH 不可执行,正在添加执行权限..." || echo "File $APP_PATH is not executable, adding execute permission..."
            ;;
        MSG_CREATE_SERVICE)
            [ "$SYSTEM_LANG" = "zh" ] && echo "创建systemd服务配置文件..." || echo "Creating systemd service configuration file..."
            ;;
        MSG_SERVICE_CREATED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "服务配置文件创建成功" || echo "Service configuration file created successfully"
            ;;
        MSG_RELOAD_SYSTEMD)
            [ "$SYSTEM_LANG" = "zh" ] && echo "重载systemd守护进程..." || echo "Reloading systemd daemon..."
            ;;
        MSG_SYSTEMD_RELOADED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "systemd重载完成" || echo "systemd reload completed"
            ;;
        MSG_ENABLE_SERVICE)
            [ "$SYSTEM_LANG" = "zh" ] && echo "启用服务自动启动..." || echo "Enabling service auto-start..."
            ;;
        MSG_SERVICE_ENABLED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "服务已设置为开机自启" || echo "Service has been set to start on boot"
            ;;
        MSG_START_SERVICE)
            [ "$SYSTEM_LANG" = "zh" ] && echo "启动服务..." || echo "Starting service..."
            ;;
        MSG_SERVICE_STARTED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "服务启动成功" || echo "Service started successfully"
            ;;
        MSG_SERVICE_FAILED)
            [ "$SYSTEM_LANG" = "zh" ] && echo "服务启动失败,请检查日志" || echo "Service failed to start, please check logs"
            ;;
        MSG_SERVICE_STATUS)
            [ "$SYSTEM_LANG" = "zh" ] && echo "服务状态:" || echo "Service status:"
            ;;
        MSG_INSTALL_START)
            [ "$SYSTEM_LANG" = "zh" ] && echo "开始安装 $APP_NAME systemd服务..." || echo "Starting to install $APP_NAME systemd service..."
            ;;
        MSG_DETECTED_PATH)
            [ "$SYSTEM_LANG" = "zh" ] && echo "检测到程序路径: $APP_PATH" || echo "Detected program path: $APP_PATH"
            ;;
        MSG_INSTALL_COMPLETE)
            [ "$SYSTEM_LANG" = "zh" ] && echo "安装完成!" || echo "Installation completed!"
            ;;
        MSG_COMMON_COMMANDS)
            [ "$SYSTEM_LANG" = "zh" ] && echo "常用命令:" || echo "Common commands:"
            ;;
        MSG_CMD_STATUS)
            [ "$SYSTEM_LANG" = "zh" ] && echo "  查看状态: sudo systemctl status $APP_NAME" || echo "  Check status: sudo systemctl status $APP_NAME"
            ;;
        MSG_CMD_LOGS)
            [ "$SYSTEM_LANG" = "zh" ] && echo "  查看日志: sudo journalctl -u $APP_NAME -f" || echo "  View logs: sudo journalctl -u $APP_NAME -f"
            ;;
        MSG_CMD_RESTART)
            [ "$SYSTEM_LANG" = "zh" ] && echo "  重启服务: sudo systemctl restart $APP_NAME" || echo "  Restart service: sudo systemctl restart $APP_NAME"
            ;;
        MSG_CMD_STOP)
            [ "$SYSTEM_LANG" = "zh" ] && echo "  停止服务: sudo systemctl stop $APP_NAME" || echo "  Stop service: sudo systemctl stop $APP_NAME"
            ;;
        MSG_CMD_UNINSTALL)
            [ "$SYSTEM_LANG" = "zh" ] && echo "  卸载服务: sudo systemctl stop $APP_NAME && sudo systemctl disable $APP_NAME && sudo rm $SERVICE_FILE && sudo systemctl daemon-reload" || echo "  Uninstall service: sudo systemctl stop $APP_NAME && sudo systemctl disable $APP_NAME && sudo rm $SERVICE_FILE && sudo systemctl daemon-reload"
            ;;
    esac
}

# 颜色输出函数
print_info() {
    printf "\033[1;34m%s\033[0m %s\n" "$(get_msg MSG_INFO_PREFIX)" "$1"
}

print_error() {
    printf "\033[1;31m%s\033[0m %s\n" "$(get_msg MSG_ERROR_PREFIX)" "$1"
}

print_success() {
    printf "\033[1;32m%s\033[0m %s\n" "$(get_msg MSG_SUCCESS_PREFIX)" "$1"
}

# 检查是否以root权限运行
check_root() {
    if [ "$(id -u)" -eq 0 ]; then
        print_error "$(get_msg MSG_NO_ROOT)"
        exit 1
    fi
}

# 检测操作系统类型
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    elif [ -f /etc/redhat-release ]; then
        echo "rhel"
    elif [ -f /etc/debian_version ]; then
        echo "debian"
    else
        echo "unknown"
    fi
}

# 安装systemd
install_systemd() {
    OS=$(detect_os)
    print_info "$(get_msg MSG_INSTALLING_SYSTEMD)"

    case "$OS" in
        ubuntu|debian)
            sudo apt-get update
            sudo apt-get install -y systemd
            ;;
        centos|rhel|fedora)
            sudo yum install -y systemd
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm systemd
            ;;
        opensuse*)
            sudo zypper install -y systemd
            ;;
        alpine)
            print_error "$(get_msg MSG_UNSUPPORTED_OS)"
            print_info "Alpine Linux uses OpenRC, not systemd"
            exit 1
            ;;
        *)
            print_error "$(get_msg MSG_UNSUPPORTED_OS)"
            exit 1
            ;;
    esac

    if [ $? -eq 0 ]; then
        print_success "$(get_msg MSG_SYSTEMD_INSTALLED)"
    else
        print_error "$(get_msg MSG_SYSTEMD_INSTALL_FAILED)"
        exit 1
    fi
}

# 检查systemd是否可用
check_systemd() {
    if ! command -v systemctl >/dev/null 2>&1; then
        print_error "$(get_msg MSG_NO_SYSTEMD)"
        install_systemd

        # 再次检查是否安装成功
        if ! command -v systemctl >/dev/null 2>&1; then
            print_error "$(get_msg MSG_SYSTEMD_INSTALL_FAILED)"
            exit 1
        fi
    fi
}

# 检查可执行文件
check_executable() {
    if [ ! -f "$APP_PATH" ]; then
        print_error "$(get_msg MSG_NO_EXECUTABLE)"
        print_error "$(get_msg MSG_EXECUTABLE_PATH)"
        exit 1
    fi

    if [ ! -x "$APP_PATH" ]; then
        print_error "$(get_msg MSG_NOT_EXECUTABLE)"
        chmod +x "$APP_PATH"
    fi
}

# 创建systemd服务文件
create_service() {
    print_info "$(get_msg MSG_CREATE_SERVICE)"

    sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=Brook Tunnel Service
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=$APP_PATH
Restart=on-failure
RestartSec=5s
User=$USER
WorkingDirectory=$(dirname "$APP_PATH")
Environment="NOTIFY_SOCKET=%t/$APP_NAME/notify"
StandardOutput=journal
StandardError=journal

# 安全加固
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=$(dirname "$APP_PATH")

[Install]
WantedBy=multi-user.target
EOF

    print_success "$(get_msg MSG_SERVICE_CREATED)"
}

# 重载systemd
reload_systemd() {
    print_info "$(get_msg MSG_RELOAD_SYSTEMD)"
    sudo systemctl daemon-reload
    print_success "$(get_msg MSG_SYSTEMD_RELOADED)"
}

# 启用服务
enable_service() {
    print_info "$(get_msg MSG_ENABLE_SERVICE)"
    sudo systemctl enable "$APP_NAME"
    print_success "$(get_msg MSG_SERVICE_ENABLED)"
}

# 显示服务状态
show_status() {
    print_info "$(get_msg MSG_SERVICE_STATUS)"
    sudo systemctl status "$APP_NAME" --no-pager || true
}

# 主函数
main() {
    print_info "$(get_msg MSG_INSTALL_START)"
    print_info "$(get_msg MSG_DETECTED_PATH)"

    check_root
    check_systemd
    check_executable
    create_service
    reload_systemd
    enable_service
    show_status

    echo ""
    print_success "$(get_msg MSG_INSTALL_COMPLETE)"
    print_info "$(get_msg MSG_COMMON_COMMANDS)"
    echo "$(get_msg MSG_CMD_STATUS)"
    echo "$(get_msg MSG_CMD_LOGS)"
    echo "$(get_msg MSG_CMD_RESTART)"
    echo "$(get_msg MSG_CMD_STOP)"
    echo "$(get_msg MSG_CMD_UNINSTALL)"
}

main
