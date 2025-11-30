#!/usr/bin/env bash
#
# Copyright ©  sixh sixh@apache.org
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

# build.sh - Compatible with Linux and macOS (including bash 3.2)

set -e

# =========================
# User input
# =========================
printf "Enter application name (default: brook-cli): "
read -r APP_NAME
APP_NAME=${APP_NAME:-brook-cli}

printf "Enter version number (default: 1.0.0): "
read -r VERSION
VERSION=${VERSION:-1.0.0}

echo "Select target OS (multiple choices allowed, e.g. 1,2,5):"
echo "1) Linux x86_64"
echo "2) Linux ARM64"
echo "3) macOS ARM64 (Apple M)"
echo "4) macOS Intel"
echo "5) Windows x86_64"
echo "6) Docker ARM64"
echo "7) Docker AMD64"
printf "Choose [1-7, comma separated]: "
read -r OS_CHOICES

printf "Copy resource directories? (y/n, default y): "
read -r COPY_RES
COPY_RES=${COPY_RES:-y}

rm -rf ./dist

# =========================
# Build target mapping (bash 3.2 compatible)
# =========================
# macOS default bash (3.2) does not support associative arrays.
# Use a case statement to map choices to build params.

resolve_target() {
    choice="$1"
    case "$choice" in
        1)
            BUILD_OS="linux"; BUILD_ARCH="amd64"; FILE_DESC="Linux-x86_64(amd64)"; DOCKER_BUILD=""; PLATFORM="" ;;
        2)
            BUILD_OS="linux"; BUILD_ARCH="arm64"; FILE_DESC="Linux-arm64"; DOCKER_BUILD=""; PLATFORM="" ;;
        3)
            BUILD_OS="darwin"; BUILD_ARCH="arm64"; FILE_DESC="macOS-ARM64(Apple-M)"; DOCKER_BUILD=""; PLATFORM="" ;;
        4)
            BUILD_OS="darwin"; BUILD_ARCH="amd64"; FILE_DESC="macOS-Intel"; DOCKER_BUILD=""; PLATFORM="" ;;
        5)
            BUILD_OS="windows"; BUILD_ARCH="amd64"; FILE_DESC="Windows-x86_64"; DOCKER_BUILD=""; PLATFORM="" ;;
        6)
            BUILD_OS="linux"; BUILD_ARCH="arm64"; FILE_DESC="Docker-ARM64"; DOCKER_BUILD="true"; PLATFORM="linux/arm64" ;;
        7)
            BUILD_OS="linux"; BUILD_ARCH="amd64"; FILE_DESC="Docker-AMD64"; DOCKER_BUILD="true"; PLATFORM="linux/amd64" ;;
        *)
            echo "⚠️  Invalid choice: $choice"; return 1 ;;
    esac
    return 0
}

# =========================
# Helper: Build for one target
# =========================
build_target() {
    i="$1"
    # resolve params for this choice
    if ! resolve_target "$i"; then
        echo "Skipping choice $i"
        return
    fi

    echo ""
    echo "=============================="
    echo " Building for $FILE_DESC ($BUILD_OS/$BUILD_ARCH)"
    echo "=============================="

    OUTPUT_DIR="dist/${APP_NAME}_${BUILD_OS}_${BUILD_ARCH}"
    TAR_NAME="dist/${APP_NAME}_${FILE_DESC}.tar.gz"

    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR/logs"
    mkdir -p "$OUTPUT_DIR/fdb"

    cp client.json "$OUTPUT_DIR"

    # Build Go binary
    echo "→ Building Go executable..."
    OUTPUT_FILE="$OUTPUT_DIR/$APP_NAME"
    if [ "$BUILD_OS" = "windows" ]; then
        OUTPUT_FILE="$OUTPUT_FILE.exe"
    fi

    BUILD_ARGS=""
    if [ "$BUILD_OS" = "windows" ]; then
        BUILD_ARGS='-ldflags=-H=windowsgui'
        GOOS=$BUILD_OS GOARCH=$BUILD_ARCH go build  $BUILD_ARGS -o "$OUTPUT_FILE" ./main.go
    else
        GOOS=$BUILD_OS GOARCH=$BUILD_ARCH go build -ldflags="-s -w" -o "$OUTPUT_FILE" ./main.go
    fi

    # Copy resources
    if [ "$COPY_RES" = "y" ] || [ "$COPY_RES" = "Y" ]; then
        for r in config static; do
            if [ -d "$r" ]; then
                cp -r "$r" "$OUTPUT_DIR/"
                echo "Copied: $r"
            fi
        done
    fi

    # Package
    if [ "$DOCKER_BUILD" = "true" ]; then
        echo "→ Building Docker image..."
        docker buildx build --build-arg APP_PATH="$OUTPUT_DIR" --platform "$PLATFORM" -t "$APP_NAME:$VERSION-$BUILD_ARCH" -f Dockerfile .
    else
        # Compatible way to delete files
        if command -v find >/dev/null 2>&1; then
            find "$OUTPUT_DIR" -name ".DS_Store" -type f -delete 2>/dev/null || true
            find "$OUTPUT_DIR" -name "._*" -type f -delete 2>/dev/null || true
        fi
        tar -czf "$TAR_NAME" -C "$OUTPUT_DIR" .
        echo "→ Packaged: $TAR_NAME"
    fi
}

# =========================
# Build multiple targets
# =========================
# Split comma-separated choices (bash 3.2 compatible)
OLD_IFS="$IFS"
IFS=','
set -- $OS_CHOICES
IFS="$OLD_IFS"

for choice in "$@"; do
    # Trim whitespace
    choice=$(echo "$choice" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
    build_target "$choice"
done

echo ""
echo "✅ All builds completed successfully!"