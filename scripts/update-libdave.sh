#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

: "${LIBDAVE_REPO:?Error: LIBDAVE_REPO is not set}"
: "${LIBDAVE_RELEASE:?Error: LIBDAVE_RELEASE is not set}"

dest="$(pwd)/libdave/lib"

rm -rf "$dest"
mkdir -p "$dest"

rm -rf /tmp/libdave

gh release download --repo "$LIBDAVE_REPO" "$LIBDAVE_RELEASE" --dir /tmp/libdave

cd /tmp/libdave

zip_files=(libdave-*.zip)
if [ ${#zip_files[@]} -eq 0 ]; then
    echo "Error: No zip files found matching libdave-*.zip"
    exit 1
fi

for zip in "${zip_files[@]}"; do
    raw_platform=$(echo "$zip" | cut -d'-' -f2,3)

    case "$raw_platform" in
        "Linux-X64")   target_name="linux_x64" ;;
        "macOS-ARM64") target_name="macos_arm64" ;;
        "macOS-X64")   target_name="macos_x64" ;;
        "Windows-X64") target_name="win_x64" ;;
        *)             target_name="$raw_platform" ;;
    esac

    mkdir -p "$dest/build/$target_name"
    unzip -q -j "$zip" "lib/libdave.*" -d "$dest/build/$target_name"

    if [[ "$target_name" == "linux_x64" ]]; then
        unzip -q -j "$zip" "include/dave/dave.h" -d "$dest/include"
        unzip -q -j "$zip" "licenses/*" -d "$dest/licenses"
    fi
done

rm -f "$dest/release.txt"
echo "$LIBDAVE_RELEASE" >> "$dest/release.txt"


