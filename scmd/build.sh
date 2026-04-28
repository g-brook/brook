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

#!/bin/sh
# build.sh

set -e

# =========================
# User input
# =========================
printf "Enter application name (default: brook-sev): "
read APP_NAME
APP_NAME=${APP_NAME:-brook-sev}

printf "Enter version number (default: 1.0.0): "
read VERSION
VERSION=${VERSION:-1.0.0}

DOCKER_IMAGE=""

echo "Select target OS (multiple choices allowed, e.g. 1,2,5):"
echo "1) Linux x86_64"
echo "2) Linux ARM64"
echo "3) macOS ARM64 (Apple M)"
echo "4) macOS Intel"
echo "5) Windows x86_64"
echo "6) Windows ARM64"
echo "7) Docker ARM64"
echo "8) Docker AMD64"
echo "9) Docker (ARM64&AMD64)"
printf "Choose [1-9, comma separated]: "
read OS_CHOICES

printf "Copy resource directories? (y/n, default y): "
read COPY_RES
COPY_RES=${COPY_RES:-y}

printf "Copy database file? (y/n, default y): "
read COPY_DB
COPY_DB=${COPY_DB:-y}

# =========================
# Build target mapping (POSIX compatible)
# =========================
# Use case statement to map choices to build params.

resolve_target() {
    choice="$1"
    case "$choice" in
        1)
            BUILD_OS="linux"; BUILD_ARCH="amd64"; FILE_DESC="Linux-x86_64(amd64)"; DOCKER_BUILD=""; PLATFORM=""; MULTI_ARCH="" ;;
        2)
            BUILD_OS="linux"; BUILD_ARCH="arm64"; FILE_DESC="Linux-arm64"; DOCKER_BUILD=""; PLATFORM=""; MULTI_ARCH="" ;;
        3)
            BUILD_OS="darwin"; BUILD_ARCH="arm64"; FILE_DESC="macOS-ARM64(Apple-M)"; DOCKER_BUILD=""; PLATFORM=""; MULTI_ARCH="" ;;
        4)
            BUILD_OS="darwin"; BUILD_ARCH="amd64"; FILE_DESC="macOS-Intel"; DOCKER_BUILD=""; PLATFORM=""; MULTI_ARCH="" ;;
        5)
            BUILD_OS="windows"; BUILD_ARCH="amd64"; FILE_DESC="Windows-x86_64"; DOCKER_BUILD=""; PLATFORM=""; MULTI_ARCH="" ;;
        6)
            BUILD_OS="windows"; BUILD_ARCH="arm64"; FILE_DESC="Windows-ARM64"; DOCKER_BUILD=""; PLATFORM=""; MULTI_ARCH="" ;;
        7)
            BUILD_OS="linux"; BUILD_ARCH="arm64"; FILE_DESC="Docker-ARM64"; DOCKER_BUILD="true"; PLATFORM="linux/arm64"; MULTI_ARCH="" ;;
        8)
            BUILD_OS="linux"; BUILD_ARCH="amd64"; FILE_DESC="Docker-AMD64"; DOCKER_BUILD="true"; PLATFORM="linux/amd64"; MULTI_ARCH="" ;;
        9)
            BUILD_OS="linux"; BUILD_ARCH="multi"; FILE_DESC="Docker-AMD64&ARM64"; DOCKER_BUILD="true"; PLATFORM="linux/amd64,linux/arm64"; MULTI_ARCH="true" ;;
        *)
            echo "⚠️  Invalid choice: $choice"; return 1 ;;
    esac
    return 0
}

# =========================
# Helper: Build for one target
# =========================
prepare_output_dir() {
    build_os="$1"
    build_arch="$2"

    OUTPUT_DIR="dist/${APP_NAME}_${build_os}_${build_arch}"
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR/logs"
    mkdir -p "$OUTPUT_DIR/fdb"

    cp server.json "$OUTPUT_DIR"

    if [ "$COPY_DB" = "y" ] || [ "$COPY_DB" = "Y" ]; then
        if [ -f db-emp.db ]; then
            cp db-emp.db "$OUTPUT_DIR/db.db"
            echo "Database copied."
        else
            echo "Warning: db-emp.db not found."
        fi
    fi

    echo "→ Building Go executable..."
    OUTPUT_FILE="$OUTPUT_DIR/$APP_NAME"
    [ "$build_os" = "windows" ] && OUTPUT_FILE="$OUTPUT_FILE.exe"

    if [ "$build_os" = "windows" ]; then
      cp run.bat "$OUTPUT_DIR"
      GOOS=$build_os GOARCH=$build_arch go build -o "$OUTPUT_FILE" ./main.go
    else
      GOOS=$build_os GOARCH=$build_arch go build -ldflags="-s -w" -o "$OUTPUT_FILE" ./main.go
    fi

    if [ "$COPY_RES" = "y" ] || [ "$COPY_RES" = "Y" ]; then
        for r in config static; do
            if [ -d "$r" ]; then
                cp -r "$r" "$OUTPUT_DIR/" && echo "Copied: $r"
            fi
        done
    fi
}

build_docker_image() {
    platform="$1"
    tag="$2"
    ensure_multiarch_builder
    docker buildx build --builder "$BUILDER_NAME" --build-arg APP_NAME="$APP_NAME" --platform "$platform" -t "$tag" -f Dockerfile . --load
}

ensure_multiarch_builder() {
    BUILDER_NAME="${BUILDER_NAME:-brook-multiarch}"
    if docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
        return 0
    fi
    docker buildx create --name "$BUILDER_NAME" --driver docker-container --use >/dev/null
    docker buildx inspect --bootstrap "$BUILDER_NAME" >/dev/null
}

get_multiarch_image_name() {
    if [ -n "$DOCKER_IMAGE" ]; then
        echo "$DOCKER_IMAGE" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
        return 0
    fi
    default_image="sixh/$APP_NAME"
    printf "Enter docker image name for multi-arch push (e.g. sixh/brook-sev; default: %s): " "$default_image" >&2
    read DOCKER_IMAGE
    DOCKER_IMAGE=${DOCKER_IMAGE:-$default_image}
    DOCKER_IMAGE=$(echo "$DOCKER_IMAGE" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    echo "$DOCKER_IMAGE"
}

build_multiarch_local() {
    image="$1"
    version="$2"
    tag_amd64="${image}:${version}-amd64"
    tag_arm64="${image}:${version}-arm64"

    echo "→ Building Docker image (linux/amd64)..."
    build_docker_image "linux/amd64" "$tag_amd64"

    echo "→ Building Docker image (linux/arm64)..."
    build_docker_image "linux/arm64" "$tag_arm64"

    echo ""
    echo "Local images built:"
    echo "  - $tag_amd64"
    echo "  - $tag_arm64"
    echo ""
    echo "Manual push (after docker login, and make sure the repository exists):"
    echo "  docker push $tag_amd64"
    echo "  docker push $tag_arm64"
    echo "  docker buildx imagetools create -t ${image}:${version} $tag_amd64 $tag_arm64"
    echo ""
}

build_target() {
    i="$1"
    # resolve params for this choice
    resolve_target "$i" || { echo "Skipping choice $i"; return; }

    echo ""
    echo "=============================="
    echo " Building for $FILE_DESC ($BUILD_OS/$BUILD_ARCH)"
    echo "=============================="

    TAR_NAME="dist/${APP_NAME}_${FILE_DESC}.tar.gz"

    if [ "$MULTI_ARCH" = "true" ]; then
        prepare_output_dir "linux" "amd64"
        prepare_output_dir "linux" "arm64"
    else
        prepare_output_dir "$BUILD_OS" "$BUILD_ARCH"
    fi

    # Package
    if [ "$DOCKER_BUILD" = "true" ]; then
        echo "→ Building Docker image..."
        if [ "$MULTI_ARCH" = "true" ]; then
            get_multiarch_image_name >/dev/null
            build_multiarch_local "$DOCKER_IMAGE" "$VERSION"
        else
            image="${DOCKER_IMAGE:-$APP_NAME}"
            build_docker_image "$PLATFORM" "$image:$VERSION-$BUILD_ARCH"
        fi
    else
        OUTPUT_DIR="dist/${APP_NAME}_${BUILD_OS}_${BUILD_ARCH}"
        find "$OUTPUT_DIR" -name ".DS_Store" -delete
        find "$OUTPUT_DIR" -name "._*" -delete
        tar -czf  "$TAR_NAME" -C "$OUTPUT_DIR" .
        echo "→ Packaged: $TAR_NAME"
    fi
}

# =========================
# Build multiple targets
# =========================
OLD_IFS="$IFS"
IFS=','
set -- $OS_CHOICES
IFS="$OLD_IFS"

for choice in "$@"; do
    choice=$(echo "$choice" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    build_target "$choice"
done

echo ""
echo "✅ All builds completed successfully!"
