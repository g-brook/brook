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


#!/bin/bash
# build.sh
# Go project build and packaging script
# Generates a versioned tar.gz package
#!/bin/bash
# interactive_build.sh
# Interactive Go build and tar.gz packaging

set -e

# =========================
# Interactive user input
# =========================
read -p "Enter application name (default: brook-sev): " APP_NAME
APP_NAME=${APP_NAME:-brook-sev}

read -p "Enter version number (default: 1.0.0): " VERSION
VERSION=${VERSION:-1.0.0}

echo "Select target OS:"
echo "1) Linux x86_64"
echo "2) macOS ARM64 (Apple M)"
echo "3) macOS Intel"
echo "4) Windows x86_64"
echo "5) docker ARM64"
echo "6) docker AMD64"
read -p "Choose [1-6]: " OS_CHOICE

case $OS_CHOICE in
    1) BUILD_OS=linux;  BUILD_ARCH=amd64 ;;
    2) BUILD_OS=darwin; BUILD_ARCH=arm64 ;;
    3) BUILD_OS=darwin; BUILD_ARCH=amd64 ;;
    4) BUILD_OS=windows; BUILD_ARCH=amd64 ;;
    5) BUILD_OS=linux; BUILD_ARCH=arm64; BUILD_DOCKER=true ;PLATFORMS=linux/arm64 ;;
    6) BUILD_OS=linux; BUILD_ARCH=amd64; BUILD_DOCKER=true ;PLATFORMS=linux/amd64 ;;
    *) echo "Invalid choice, default to Linux x86_64"; BUILD_OS=linux; BUILD_ARCH=amd64 ;;
esac

read -p "Copy resource directories? (y/n, default y): " COPY_RES
COPY_RES=${COPY_RES:-y}

read -p "Copy database file? (y/n, default y): " COPY_DB
COPY_DB=${COPY_DB:-y}

# =========================
# Setup directories
# =========================
OUTPUT_DIR="dist/${APP_NAME}"
TAR_NAME="${APP_NAME}_${VERSION}_${BUILD_OS}_${BUILD_ARCH}.tar.gz"

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"/logs
mkdir -p "$OUTPUT_DIR"/fdb

# Copy mandatory files
cp server.json "$OUTPUT_DIR"

# Optional database copy
if [[ "$COPY_DB" == "y" || "$COPY_DB" == "Y" ]]; then
    if [ -f db-emp.db ]; then
        cp db-emp.db "$OUTPUT_DIR/db.db"
        echo "Database copied and renamed."
    else
        echo "Warning: db-emp.db not found, skipping."
    fi
fi

# =========================
# Build Go executable
# =========================
echo "Building Go executable for $BUILD_OS/$BUILD_ARCH..."
OUTPUT_FILE="$OUTPUT_DIR/$APP_NAME"
if [ "$BUILD_OS" == "windows" ]; then
    OUTPUT_FILE="$OUTPUT_FILE.exe"
fi

GOOS=$BUILD_OS GOARCH=$BUILD_ARCH go build -o "$OUTPUT_FILE" ./main.go
echo "Build completed: $OUTPUT_FILE"

# =========================
# Copy resource files
# =========================
if [[ "$COPY_RES" == "y" || "$COPY_RES" == "Y" ]]; then
    RESOURCES=("config" "static")
    for r in "${RESOURCES[@]}"; do
        if [ -d "$r" ]; then
            cp -r "$r" "$OUTPUT_DIR/"
            echo "Copied resource directory: $r"
        fi
    done
fi

# =========================
# Create tar.gz package
# =========================
if [ "$BUILD_DOCKER" = true  ]; then
  docker buildx build --platform "$PLATFORMS" -t "$APP_NAME":"$VERSION" -f Dockerfile .
fi
else
  tar -czf "$TAR_NAME" -C "$OUTPUT_DIR" .
fi

echo "Packaging completed: $TAR_NAME"