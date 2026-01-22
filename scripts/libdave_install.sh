#!/bin/sh
set -u

# --- Configuration & Globals ---
REPO_OWNER="discord"
REPO_NAME="libdave"
VERSION="${1:-}"
SSL_FLAVOUR="${SSL_FLAVOUR:-boringssl}"
NON_INTERACTIVE=${NON_INTERACTIVE:-}
FORCE_BUILD=${FORCE_BUILD:-}

LIBDAVE_REPO="https://github.com/$REPO_OWNER/$REPO_NAME"
LIB_DIR="$HOME/.local/lib"
BIN_DIR="$HOME/.local/bin"
INC_DIR="$HOME/.local/include"
PC_DIR="$LIB_DIR/pkgconfig"
PC_FILE="$PC_DIR/dave.pc"

log() { echo "-> $*"; }
error_exit() { echo "Error: $1" >&2; exit 1; }

check_dependencies() {
    deps="$1"
    for cmd in $deps; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            error_exit "Missing dependency: $cmd."
        fi
    done
}

detect_environment() {
    PLATFORM=$(uname -s)
    ARCH=$(uname -m)

    case "${PLATFORM}" in
        Linux)
            OS="linux"; LIB_EXT="so"; LIB_VAR="LD_LIBRARY_PATH"; GITHUB_OS="Linux" ;;
        Darwin)
            OS="macos"; LIB_EXT="dylib"; LIB_VAR="DYLD_LIBRARY_PATH"; GITHUB_OS="macOS" ;;
        MSYS*|MINGW*|CYGWIN*)
            OS="win"; LIB_EXT="lib"; LIB_VAR="PATH"; GITHUB_OS="Windows" ;;
        *) error_exit "Unsupported OS: $PLATFORM" ;;
    esac

    case "${ARCH}" in
        x86_64) GITHUB_ARCH="X64" ;;
        arm64|aarch64) GITHUB_ARCH="ARM64" ;;
        *) GITHUB_ARCH="$ARCH" ;;
    esac

    # Check if stdin is a terminal for interactivity
    if [ ! -t 0 ]; then NON_INTERACTIVE=1; fi
}

try_download_prebuilt() {
    tag="$1"
    asset_pattern="libdave-$GITHUB_OS-$GITHUB_ARCH-$SSL_FLAVOUR.zip"

    check_dependencies "curl unzip"

    # Construct direct download URL (GitHub standard format)
    download_url="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$tag/$asset_pattern"

    log "Checking for prebuilt asset at: $download_url"

    tmp_zip="/tmp/libdave_prebuilt.zip"
    if curl -fsL "$download_url" -o "$tmp_zip"; then
        log "Found prebuilt asset. Extracting..."
        mkdir -p "$LIB_DIR" "$INC_DIR"

        unzip -j -o "$tmp_zip" "include/dave/dave.h" -d "$INC_DIR"
        if [ "$OS" = "win" ]; then
            unzip -j -o "$tmp_zip" "bin/libdave.dll" -d "$BIN_DIR"
            unzip -j -o "$tmp_zip" "lib/libdave.lib" -d "$LIB_DIR"
        else
            unzip -j -o "$tmp_zip" "lib/libdave.$LIB_EXT" -d "$LIB_DIR"
        fi
        rm -f "$tmp_zip"
        return 0
    else
        log "No prebuilt asset found. Falling back to build."
        return 1
    fi
}

manual_build() {
    checkout_ref="$1"
    log "Starting manual build process for ref: $checkout_ref"

    check_dependencies "git make cmake"

    WORK_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'libdave')

    git clone "$LIBDAVE_REPO" "$WORK_DIR/libdave"
    cd "$WORK_DIR/libdave/cpp" || exit 1

    git checkout "$checkout_ref"
    git submodule update --init --recursive

    if [ "$OS" = "win" ]; then
        ./vcpkg/bootstrap-vcpkg.bat -disableMetrics
    else
        ./vcpkg/bootstrap-vcpkg.sh -disableMetrics
    fi

    log "Compiling shared library..."
    make shared SSL="$SSL_FLAVOUR" BUILD_TYPE=Release

    log "Installing..."
    mkdir -p "$LIB_DIR" "$INC_DIR" "$PC_DIR"
    cp includes/dave/dave.h "$INC_DIR/"

    if [ "$OS" = "win" ]; then
        mkdir -p "$BIN_DIR"
        cp "build/Release/libdave.dll" "$BIN_DIR/"
        cp "build/Release/libdave.lib" "$LIB_DIR/"
    else
        cp "build/libdave.$LIB_EXT" "$LIB_DIR/"
    fi

    # Cleanup
    rm -rf "$WORK_DIR"
}

generate_pkg_config() {
    log "Generating pkg-config metadata..."
    mkdir -p "$PC_DIR"
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
}

update_shell_profile() {
    case "$SHELL" in
        *zsh*) profile="$HOME/.zshrc" ;;
        *) profile="$HOME/.bashrc" ;;
    esac

    pc_line="export PKG_CONFIG_PATH=\"\$HOME/.local/lib/pkgconfig:\$PKG_CONFIG_PATH\""
    lib_line="export $LIB_VAR=\"\$HOME/.local/lib:\$$LIB_VAR\""
    path_line="export \$PATH=\"\$HOME/.local/path:\$PATH\""

    # Use grep -F for fixed string matching
    needs_pc=0;
    needs_lib=0
    needs_path=0;
    if [ -f "$profile" ]; then
        grep -qF "$pc_line" "$profile" || needs_pc=1
        if [ "$OS" = "win" ]; then
            grep -qF "$path_line" "$profile" || needs_path=1
        else
            grep -qF "$lib_line" "$profile" || needs_lib=1
        fi
    else
        needs_pc=1
        if [ "$OS" = "win" ]; then
            needs_path=1
        else
            needs_lib=1
        fi
    fi

    if [ "$needs_pc" -eq 1 ] || [ "$needs_lib" -eq 1 ]; then
        printf "\n--- Action Required ---\n"
        printf "Add these to %s:\n" "$profile"
        [ "$needs_pc" -eq 1 ] && printf "    %s\n" "$pc_line"
        [ "$needs_lib" -eq 1 ] && printf "    %s\n" "$lib_line"
        [ "$needs_path" -eq 1 ] && printf "    %s\n" "$path_line"

        if [ -z "$NON_INTERACTIVE" ]; then
            printf "Apply changes automatically? (y/n): "
            read -r reply
            case "$reply" in
                [Yy]*)
                    {
                        [ "$needs_pc" -eq 1 ] && printf "%s\n" "$pc_line"
                        [ "$needs_lib" -eq 1 ] && printf "%s\n" "$lib_line"
                        [ "$needs_path" -eq 1 ] && printf "%s\n" "$path_line"
                    } >> "$PROFILE_FILE"
                    log "Profile updated."
                    ;;
            esac
        fi
    fi
}

main() {
    if [ -z "$VERSION" ]; then
        error_exit "Usage: $0 <version-tag> (e.g., v0.0.1 or GIT SHA)"
    fi

    detect_environment

    is_sha=$(echo "$VERSION" | grep -E "^[0-9a-fA-F]{7,40}$" >/dev/null 2>&1; echo $?)
    if [ "$is_sha" -eq 0 ]; then
        version="$VERSION"
    else
        # Normalize the version to add the '/cpp' extension if its missing
        stripped_version=$(echo "$VERSION" | sed 's/\/cpp$//')
        version="${stripped_version}/cpp"
    fi

    if [ "$is_sha" -eq 0 ] || [ -n "$FORCE_BUILD" ]; then
        manual_build "$version"
    else
        try_download_prebuilt "$version" || manual_build "$version"
    fi

    generate_pkg_config
    update_shell_profile

    log "Installation successful: libdave $VERSION ($ARCH)"
}

main "$@"
