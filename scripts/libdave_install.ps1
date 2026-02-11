# libdave-install.ps1
# Usage: .\libdave-install.ps1 -Version "v0.0.1"

[CmdletBinding(PositionalBinding=$false)]
param (
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Version,
    [switch]$ForceBuild,
    [string]$SslFlavour = "boringssl"
)

$ErrorActionPreference = "Stop"

# --- Configuration ---
$RepoOwner = "discord"
$RepoName = "libdave"
$LibDaveRepo = "https://github.com/$RepoOwner/$RepoName"

$InstallBase = Join-Path $env:LOCALAPPDATA "libdave"
$BinDir = Join-Path $InstallBase "bin"
$LibDir = Join-Path $InstallBase "lib"
$IncDir = Join-Path $InstallBase "include"
$PcDir = Join-Path $env:LOCALAPPDATA "pkgconfig"
$PcFile = Join-Path $PcDir "dave.pc"

function Log-Info ([string]$Msg) { Write-Host "-> $Msg" -ForegroundColor Cyan }

function Check-Dependencies {
    $deps = @("git", "make", "cmake")
    foreach ($cmd in $deps) {
        if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
            Write-Error "Missing dependency: $cmd. Please install it via winget or choco."
        }
    }
}

function Get-Environment {
    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64"   { "X64" }
        "ARM64"   { "ARM64" }
        Default   { $_ }
    }
    return @{ Arch = $arch }
}

function Install-Prebuilt {
    param($Tag, $Env)
    $AssetPattern = "libdave-Windows-$($Env.Arch)-$SslFlavour.zip"
    $DownloadUrl = "$LibDaveRepo/releases/download/$Tag/$AssetPattern"
    $TempZip = Join-Path $env:TEMP "libdave_prebuilt.zip"

    Log-Info "Checking for prebuilt asset at: $DownloadUrl"

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
    } catch {
        Log-Info "No prebuilt asset found. Falling back to build."
        return $false
    }

    Log-Info "Found prebuilt asset. Extracting..."

    if (-not (Test-Path $InstallBase)) { New-Item -ItemType Directory -Path $InstallBase }

    Expand-Archive -Path $TempZip -DestinationPath "$env:TEMP\libdave_stage" -Force

    # Copy specific files to the install directories
    Remove-Item $InstallBase -Recurse
    New-Item -ItemType Directory -Path $BinDir, $LibDir, $IncDir -Force | Out-Null
    Copy-Item "$env:TEMP\libdave_stage\include\dave\dave.h" -Destination $IncDir -Recurse
    Copy-Item "$env:TEMP\libdave_stage\bin\libdave.dll" -Destination $BinDir
    Copy-Item "$env:TEMP\libdave_stage\lib\libdave.lib" -Destination $LibDir

    Remove-Item $TempZip -Force
    return $true
}

function Build-Manual {
    param($Ref)
    Log-Info "Starting manual build process for ref: $Ref ($SslFlavour)"
    Check-Dependencies

    $WorkDir = Join-Path $env:TEMP "libdave_build_$(New-Guid)"
    New-Item -ItemType Directory -Path $WorkDir | Out-Null

    git clone $LibDaveRepo $WorkDir
    $CurrentDir = Get-Location
    Set-Location (Join-Path $WorkDir "cpp")

    git checkout $Ref
    git submodule update --init --recursive

    Log-Info "Bootstrapping vcpkg..."
    .\vcpkg\bootstrap-vcpkg.bat -disableMetrics

    Log-Info "Compiling shared library..."
    make shared "SSL=$SslFlavour" BUILD_TYPE=Release

    Log-Info "Installing..."

    Remove-Item $InstallBase -Recurse
    New-Item -ItemType Directory -Path $BinDir, $LibDir, $IncDir -Force | Out-Null
    Copy-Item "includes\dave\dave.h" -Destination $IncDir
    Copy-Item "build\Release\libdave.dll" -Destination $BinDir
    Copy-Item "build\Release\libdave.lib" -Destination $LibDir

    Set-Location $CurrentDir
    Remove-Item $WorkDir -Recurse -Force
}

function Generate-PkgConfig {
    Log-Info "Generating pkg-config metadata..."

    if (-not (Test-Path $PcDir)) { New-Item -ItemType Directory -Path $PcDir -Force | Out-Null }

    # We use forward slashes for the .pc file as many pkg-config tools
    # on Windows (like those in MSYS2/Cygwin) prefer them.
    $Prefix = $InstallBase.Replace('\', '/')

    # For some reason, pkgconfiglite doesnt't like variables, so always expand $Prefix for now until a fix is found
    $PcContent = @"
prefix=$Prefix
exec_prefix=$Prefix/bin
libdir=$Prefix/lib
includedir=$Prefix/include

Name: dave
Description: Discord Audio & Video End-to-End Encryption (DAVE) Protocol
Version: $Version
URL: $LibDaveRepo
Libs: -L`${libdir} -ldave
Cflags: -I`${includedir}
"@

    Out-File -FilePath $PcFile -InputObject $PcContent -Encoding UTF8
    Log-Info "Created $PcFile"
}

function Update-EnvironmentVariables {
    Log-Info "Updating User PATH..."
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($CurrentPath -notlike "*$BinDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$BinDir;$CurrentPath", "User")
    }

    Log-Info "Updating PKG_CONFIG_PATH..."
    $CurrentPkgPath = [Environment]::GetEnvironmentVariable("PKG_CONFIG_PATH", "User")
    if ($CurrentPkgPath -notlike "*$PcDir*") {
        $NewPkgPath = if ([string]::IsNullOrEmpty($CurrentPkgPath)) { $PcDir } else { "$PcDir;$CurrentPkgPath" }
        [Environment]::SetEnvironmentVariable("PKG_CONFIG_PATH", $NewPkgPath, "User")
    }
}


# --- Main Logic ---
$CurrentDir = Get-Location
try {
    $EnvInfo = Get-Environment
    $IsSha = $Version -match "^[0-9a-fA-F]{7,40}$"
    $BuildRef = if ($IsSha) { $Version } else { "$($Version.Replace('/cpp',''))/cpp" }

    if ($IsSha -or $ForceBuild) {
        Build-Manual -Ref $BuildRef
    } else {
        $Success = Install-Prebuilt -Tag $BuildRef -Env $EnvInfo
        if (-not $Success) { Build-Manual -Ref $BuildRef }
    }

    Generate-PkgConfig
    Update-EnvironmentVariables
    Log-Info "Installation successful: libdave $Version ($($EnvInfo.Arch))"
} finally {
    Set-Location $CurrentDir
}
