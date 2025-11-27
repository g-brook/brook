#!/bin/bash

APP_NAME="brook-sev"
APP_PATH="$(pwd)/$APP_NAME"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"

echo "Creating systemd service for $APP_NAME..."
echo "Detected program path: $APP_PATH"

if [ ! -f "$APP_PATH" ]; then
    echo "Error: executable $APP_NAME not found in current directory."
    exit 1
fi

sudo bash -c "cat > $SERVICE_FILE" <<EOF
[Unit]
Description=Brook Tunnel Service
After=network.target

[Service]
ExecStart=$APP_PATH
Restart=always
User=$USER
WorkingDirectory=$(pwd)
Environment=NOTIFY_SOCKET=\$NOTIFY_SOCKET
Type=notify

[Install]
WantedBy=multi-user.target
EOF

echo "Reload systemd..."
sudo systemctl daemon-reload

echo "Enable service..."
sudo systemctl enable $APP_NAME

echo "Start service..."
sudo systemctl start $APP_NAME

echo "Installation completed!"
systemctl status $APP_NAME