#!/bin/bash
set -euo pipefail
IFS=$'\n\t'


VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo "Please specify as an argument the version to download"
    exit 1
fi

VERSION="${VERSION%/cpp}/cpp"
LIBDAVE_REPO="discord/libdave"
SSL_FLAVOUR="boringssl"

dest="$(pwd)/libdave/vendor"

rm -rf "$dest"
mkdir -p "$dest"

rm -rf /tmp/libdave

gh release download --repo "$LIBDAVE_REPO" "$VERSION" --dir /tmp/libdave

cd /tmp/libdave

zip_files=(libdave-*-"$SSL_FLAVOUR".zip)
if [ ${#zip_files[@]} -eq 0 ]; then
    echo "Error: No zip files found matching libdave-*-$SSL_FLAVOUR.zip"
    exit 1
fi

for zip in "${zip_files[@]}"; do
  echo "-> Extracting $zip"

  raw_platform=$(echo "$zip" | cut -d'-' -f2,3)

  os_part="${raw_platform%-*}"
  arch_part="${raw_platform##*-}"
  case "$os_part" in
    "Linux")   os_slug="linux" ;;
    "macOS")   os_slug="macos" ;;
    "Windows") os_slug="win"   ;;
    *)         os_slug=$(echo "$os_part" | tr '[:upper:]' '[:lower:]') ;;
  esac
  case "$arch_part" in
    "X64")   arch_slug="x64" ;;
    "ARM64") arch_slug="arm64" ;;
    *)       arch_slug=$(echo "$arch_part" | tr '[:upper:]' '[:lower:]') ;;
  esac
  target_name="${os_slug}_${arch_slug}"

  mkdir -p "$dest/lib/$target_name"
  unzip -j "$zip" "lib/libdave.*" -d "$dest/lib/$target_name"

  if [[ "$os_slug" == "win" ]]; then
    echo "--> Extracting bin"
    mkdir -p "$dest/bin/$target_name"
    unzip -j "$zip" "bin/libdave.dll" -d "$dest/bin/$target_name"
  fi

  if [[ "$target_name" == "linux_x64" ]]; then
    echo "--> Extracting header and licenses"
    # All the folders will have the same, so we might as well only extract it from one
    unzip -j "$zip" "include/dave/dave.h" -d "$dest/include"
    unzip -j "$zip" "licenses/*" -d "$dest/licenses"
  fi
done

rm -f "$dest/release.txt"
echo "$VERSION" >> "$dest/release.txt"
