#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

LIBDAVE_REPO=https://github.com/discord/libdave
LIBDAVE_SHA=74979cb33febf4ddef0c2b66e57520b339550c17

# Determine OS
if [[ "$(uname -s)" == "darwin"* ]]; then
  PLATFORM=macos
else
  PLATFORM=linux
fi

# Dependencies
REQUIRED_CMDS=("git" "make" "cmake")
for cmd in "${REQUIRED_CMDS[@]}"; do
  if ! command -v "$cmd" &> /dev/null; then
    echo "Error: $cmd is not installed."
    if [ "$PLATFORM" == "macos" ]; then
      echo "Please run: brew install $cmd"
    else
      echo "Please install it using your package manager (apt, dnf, etc.)"
    fi
    exit 1
  fi
done


# Installation paths
LIB_DIR="$HOME/.local/lib"
INC_DIR="$HOME/.local/include"
PC_DIR="$LIB_DIR/pkgconfig"
PC_FILE="$PC_DIR/dave.pc"

# OS Specific extensions
if [ "$PLATFORM" == "macos" ]; then
    LIB_EXT="dylib"
    LIB_VAR="DYLD_LIBRARY_PATH"
else
    LIB_EXT="so"
    LIB_VAR="LD_LIBRARY_PATH"
fi

echo "-> Cloning repository"
WORK_DIR=$(mktemp -d)
cd "$WORK_DIR"

git clone "$LIBDAVE_REPO" libdave
cd libdave/cpp
git checkout "$LIBDAVE_SHA"

git submodule update --init --recursive
./vcpkg/bootstrap-vcpkg.sh -disableMetrics

echo "-> Building shared library for $PLATFORM"
make shared

echo "-> Installing to $LIB_DIR"
mkdir -p "$LIB_DIR" "$INC_DIR" "$PC_DIR"

cp includes/dave.h "$INC_DIR/"

# Handle potential naming variations in build output
if [ -f "build/libdave.$LIB_EXT" ]; then
    cp "build/libdave.$LIB_EXT" "$LIB_DIR/"
elif [ -f "build/libdave.so" ] && [ "$PLATFORM" == "macos" ]; then
    cp "build/libdave.so" "$LIB_DIR/libdave.dylib"
else
    cp build/libdave.* "$LIB_DIR/" 2>/dev/null || echo "Warning: Could not find build artifacts"
fi

echo "-> Generating pkg-config metadata"
cat <<EOF > "$PC_FILE"
prefix=$HOME/.local
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: dave
Description: Discord Audio & Video End-to-End Encryption (DAVE) Protocol
Version: $LIBDAVE_SHA
URL: $LIBDAVE_REPO
Libs: -L\${libdir} -ldave -Wl,-rpath,\${libdir}
Cflags: -I\${includedir}
EOF

echo "-> Cleaning up"
rm -rf "$WORK_DIR"

echo "--- Installation Complete ---"
echo "libdave revision installed: $LIBDAVE_SHA"
echo
echo "Please update your shell profile (.bashrc, .zshrc, etc) with the following lines:"
echo
echo "export PKG_CONFIG_PATH=\"$PC_DIR:\$PKG_CONFIG_PATH\""
echo "export $LIB_VAR=\"$LIB_DIR:\$$LIB_VAR\""
