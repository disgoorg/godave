#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Configuration
VERSION="v1.1.0"
SSL_FLAVOUR="boringssl"
REPO="discord/libdave"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

LIB_DIR="$HOME/.local/lib"
INC_DIR="$HOME/.local/include"
PC_DIR="$LIB_DIR/pkgconfig"
PC_FILE="$PC_DIR/dave.pc"

# Determine platform and architecture
PLATFORM=$(uname -s)
ARCH=$(uname -m)

# Map OS and ARCH to Discord's release structure
case "${PLATFORM}" in
    Darwin)
      OS_KEY="macOS"
      LIB_EXT="dylib"
      LIB_VAR="DYLD_LIBRARY_PATH"
      ;;
    Linux)
      OS_KEY="Linux"
      LIB_EXT="so"
      LIB_VAR="LD_LIBRARY_PATH"
      ;;
    "MINGW"*|"MSYS"*|"CYGWIN"*)
      OS_KEY="Windows"
      LIB_EXT="lib"
      LIB_VAR="PATH"
      ;;
    *) echo "Unsupported OS"; exit 1 ;;
esac

case "${ARCH}" in
    x86_64|amd64) ARCH_KEY="X64" ;;
    arm64|aarch64) ARCH_KEY="ARM64" ;;
    *) echo "Unsupported Arch"; exit 1 ;;
esac

# Dependencies
REQUIRED_CMDS=("curl" "unzip" "pkg-config")
for cmd in "${REQUIRED_CMDS[@]}"; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "Error: $cmd is not installed."
        exit 1
    fi
done

# Find a matching release
DOWNLOAD_URL=$(curl -s "$API_URL" | \
    grep "browser_download_url" | \
    grep -i "$SSL_FLAVOUR" | \
    grep -i "$OS_KEY" | \
    grep -i "$ARCH_KEY" | \
    head -n 1 | \
    cut -d '"' -f 4)

if [[ -z "$DOWNLOAD_URL" ]]; then
    echo "Error: Could not find a matching release asset for $OS_KEY-$ARCH_KEY on GitHub."
    exit 1
fi

FILE_NAME=$(basename "$DOWNLOAD_URL")

# Download and install
echo "-> Downloading $FILE_NAME for $PLATFORM ($ARCH)"
WORK_DIR=$(mktemp -d)
cd "$WORK_DIR"
curl -L -o "$FILE_NAME" "$DOWNLOAD_URL"

echo "-> Extracting files"
unzip "$FILE_NAME" -d libdave

echo "-> Installing to $LIB_DIR"
mkdir -p "$LIB_DIR" "$INC_DIR" "$PC_DIR"

# Copy headers and libraries from the extracted folder
cp libdave/include/dave/dave.h "$INC_DIR/"
cp "libdave/lib/libdave.$LIB_EXT" "$LIB_DIR/"

echo "-> Generating pkg-config metadata"
cat <<EOF > "$PC_FILE"
prefix=$HOME/.local
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: dave
Description: Discord Audio & Video End-to-End Encryption (DAVE) Protocol
Version: $VERSION
URL: https://github.com/$REPO
Libs: -L\${libdir} -ldave -Wl,-rpath,\${libdir}
Cflags: -I\${includedir}
EOF

echo "-> Cleaning up"
rm -rf "$WORK_DIR"

echo "--- Installation Complete ---"
echo "libdave version installed: $VERSION ($ARCH)"
echo
echo "Please update your shell profile with the following lines (if not done so already):"
echo
echo "export PKG_CONFIG_PATH=\"$PC_DIR:\$PKG_CONFIG_PATH\""
echo "export $LIB_VAR=\"$LIB_DIR:\$$LIB_VAR\""
