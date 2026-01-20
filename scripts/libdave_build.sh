#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo "Please specify as an argument the version to download"
    exit 1
fi

# Configuration
LIBDAVE_REPO="https://github.com/discord/libdave"
VERSION="${VERSION%/cpp}/cpp"
SSL_FLAVOUR="${SSL_FLAVOUR:-boringssl}"
NON_INTERACTIVE=${NON_INTERACTIVE:-}

LIB_DIR="$HOME/.local/lib"
INC_DIR="$HOME/.local/include"
PC_DIR="$LIB_DIR/pkgconfig"
PC_FILE="$PC_DIR/dave.pc"

# Set NON_INTERACTIVE if in a non-interactive shell
case $- in
    *i*) ;;
      *) NON_INTERACTIVE=1;;
esac

# Determine platform and architecture
PLATFORM=$(uname -s)
ARCH=$(uname -m)

case "${PLATFORM}" in
    Darwin)
      LIB_EXT="dylib"
      LIB_VAR="DYLD_LIBRARY_PATH"
      OS="macos"
      ;;
    Linux)
      LIB_EXT="so"
      LIB_VAR="LD_LIBRARY_PATH"
      OS="linux"
      ;;
    "MINGW"*|"MSYS"*|"CYGWIN"*)
      LIB_EXT="lib"
      LIB_VAR="PATH"
      OS="windows"
      ;;
    *) echo "Unsupported OS"; exit 1 ;;
esac

# Dependencies
REQUIRED_CMDS=("git" "make" "cmake" "curl" "zip")
for cmd in "${REQUIRED_CMDS[@]}"; do
  if ! command -v "$cmd" &> /dev/null; then
    echo "Error: $cmd is not installed."
    if [ "$OS" == "macos" ]; then
      echo "Please run: brew install $cmd"
    else
      echo "Please install it using your package manager (apt, dnf, etc.)"
    fi
    exit 1
  fi
done

echo "-> Cloning repository"
WORK_DIR=$(mktemp -d)
cd "$WORK_DIR"

git clone "$LIBDAVE_REPO" libdave
cd libdave/cpp
git checkout "$VERSION"

git submodule update --init --recursive
./vcpkg/bootstrap-vcpkg.sh -disableMetrics

echo "-> Building shared library for $PLATFORM (SSL: $SSL_FLAVOUR)"
make shared SSL="$SSL_FLAVOUR" BUILD_TYPE=Release

echo "-> Installing to $LIB_DIR"
mkdir -p "$LIB_DIR" "$INC_DIR" "$PC_DIR"

cp includes/dave/dave.h "$INC_DIR/"

if [ "$OS" == "windows" ]; then
  # We need to copy the DLL and LIB files
  cp "build/Release/libdave.$LIB_EXT" "$LIB_DIR/"
  cp "build/Release/libdave.dll" "$LIB_DIR/"
else
  cp "build/libdave.$LIB_EXT" "$LIB_DIR/"
fi

echo "-> Generating pkg-config metadata"
cat <<EOF > "$PC_FILE"
prefix=$HOME/.local
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: dave
Description: Discord Audio & Video End-to-End Encryption (DAVE) Protocol
Version: $VERSION
URL: $LIBDAVE_REPO
Libs: -L\${libdir} -ldave
Cflags: -I\${includedir}
EOF

echo "-> Cleaning up"
rm -rf "$WORK_DIR"

echo "--- Installation Complete ---"
echo "libdave version installed: $VERSION ($ARCH)"

# Identify the shell profile (defaults to .bashrc, or .zshrc if on macOS/zsh)
PROFILE_FILE="$HOME/.bashrc"
[[ "$SHELL" == *"zsh"* ]] && PROFILE_FILE="$HOME/.zshrc"

PC_LINE="export PKG_CONFIG_PATH=\"\$HOME/.local/lib/pkgconfig:\$PKG_CONFIG_PATH\""
LIB_LINE="export $LIB_VAR=\"\$HOME/.local/lib:\$$LIB_VAR\""

if [ -f "$PROFILE_FILE" ]; then
  NEEDS_PC=$(grep -qF -- "$PC_LINE" "$PROFILE_FILE"; echo $?)
  NEEDS_LIB=$(grep -qF -- "$LIB_LINE" "$PROFILE_FILE"; echo $?)
fi

# Check if lines already exist
if [ "$NEEDS_PC" -eq 1 ] || [ "$NEEDS_LIB" -eq 1 ]; then
  echo
  echo "The following lines are missing from your $PROFILE_FILE:"
  [[ "$NEEDS_PC" -eq 1 ]] && echo "    $PC_LINE"
  [[ "$NEEDS_LIB" -eq 1 ]] && echo "    $LIB_LINE"

  if [ -z "$NON_INTERACTIVE" ] ; then
    read -p "Would you like to add them now? (y/n) " -r
  else
    REPLY="y"
  fi

  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    {
        [[ "$NEEDS_PC" -eq 1 ]] && echo "$PC_LINE"
        [[ "$NEEDS_LIB" -eq 1 ]] && echo "$LIB_LINE"
    } >> "$PROFILE_FILE"

    echo "Profile updated! Please run 'source $PROFILE_FILE' or restart your terminal for the changes to apply"
  else
    echo "Skipped. Please add the lines manually to use libdave"
  fi
fi
