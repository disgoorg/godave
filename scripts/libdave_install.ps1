$ErrorActionPreference = "Stop"

# Check Dependencies
$requiredCmds = @("git", "cmake")
foreach ($cmd in $requiredCmds) {
    if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
        Write-Error "Error: $cmd is not installed. Please install it and try again."
        exit 1
    }
}

$LIBDAVE_REPO = "https://github.com/discord/libdave"
$LIBDAVE_SHA = "74979cb33febf4ddef0c2b66e57520b339550c17"

$INSTALL_ROOT = Join-Path $HOME ".local"
$LIB_DIR = Join-Path $INSTALL_ROOT "lib"
$INC_DIR = Join-Path $INSTALL_ROOT "include"
$PC_DIR = Join-Path $LIB_DIR "pkgconfig"
$PC_FILE = Join-Path $PC_DIR "dave.pc"

$TEMP_DIR = Join-Path $env:TEMP "libdave_build"
if (Test-Path $TEMP_DIR) { Remove-Item -Recurse -Force $TEMP_DIR }
New-Item -ItemType Directory -Path $TEMP_DIR | Out-Null

Write-Host "-> Cloning repository"
Set-Location $TEMP_DIR
git clone $LIBDAVE_REPO libdave
Set-Location libdave/cpp
git checkout $LIBDAVE_SHA

git submodule update --init --recursive
.\vcpkg\bootstrap-vcpkg.bat -disableMetrics

Write-Host "-> Building shared library"
cmake --build build --target libdave

Write-Host "-> Installing files"
if (-not (Test-Path $LIB_DIR)) { New-Item -ItemType Directory -Path $LIB_DIR }
if (-not (Test-Path $INC_DIR)) { New-Item -ItemType Directory -Path $INC_DIR }
if (-not (Test-Path $PC_DIR)) { New-Item -ItemType Directory -Path $PC_DIR }

Copy-Item "build\Release\dave.dll" -Destination $LIB_DIR
Copy-Item "build\Release\dave.lib" -Destination $LIB_DIR
Copy-Item "includes\dave.h" -Destination $INC_DIR

Write-Host "-> Generating pkg-config metadata"
$PC_CONTENT = @"
prefix=$($INSTALL_ROOT.Replace('\', '/'))
exec_prefix=\${prefix}
libdir=\${prefix}/lib
includedir=\${prefix}/include

Name: dave
Description: Discord Audio & Video End-to-End Encryption (DAVE) Protocol
Version: $LIBDAVE_SHA
URL: $LIBDAVE_REPO
Libs: -L\${libdir} -ldave
Cflags: -I\${includedir}
"@
$PC_CONTENT | Out-File -Encoding ascii $PC_FILE

Write-Host "-> Cleaning up"
Set-Location $HOME
Remove-Item -Recurse -Force $TEMP_DIR

Write-Host "--- Installation Complete ---"
Write-Host "Add $LIB_DIR to your PATH environment variable."
Write-Host "Set PKG_CONFIG_PATH to $PC_DIR"