[![Go Reference](https://pkg.go.dev/badge/github.com/disgoorg/godave.svg)](https://pkg.go.dev/github.com/disgoorg/godave)
[![Go Report](https://goreportcard.com/badge/github.com/disgoorg/godave)](https://goreportcard.com/report/github.com/disgoorg/godave)
[![Go Version](https://img.shields.io/github/go-mod/go-version/disgoorg/godave)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![GoDave Version](https://img.shields.io/github/v/tag/disgoorg/godave?label=release)](https://github.com/disgoorg/godave/releases/latest)
[![DisGo Discord](https://discord.com/api/guilds/817327181659111454/widget.png)](https://discord.gg/TewhTfDpvW)

<img align="right" src="/.github/godave_gopher.png" width=192 alt="discord gopher">

# GoDave

GoDave is a library that provides Go bindings for [libdave](https://github.com/discord/libdave) and provides a generic DAVE interface allowing for
different implementations in the future.

## Summary

1. [Libdave Installation](#libdave-installation)
   1. [Windows Installation](#windows-instructions)
   2. [Installing manually](#manual-installation)
2. [Example Usage](#example-usage)
3. [License](#license)

## Libdave Installation

This library uses CGO and dynamic linking to use libdave.

We provide helpful scripts under [scripts/](https://github.com/disgoorg/godave/tree/master/scripts) to allow you to
download pre-built binaries of build them yourself, depending on your needs. Please audit them before executing!

> [!NOTE]
> Due to the nature of this project, it might be necessary to re-install libdave when updating to a new GoDave version.
> The version that require this may be indicated with a minor bump (for reference: `mayor.minor.patch`).
>
> You can see what version is required by checking [this file](https://github.com/disgoorg/godave/tree/master/libdave/release.txt)

### Linux/MacOS/WSL instructions

Open a terminal and execute the following commands:

```bash
# Set CC/CXX variables to change the compiler used (ie, for clang
#export CC=/usr/bin/clang CXX=/usr/bin/clang

./libdave_install.sh v1.1.0
```

### MUSL Linux

If you want to build a MUSL version of libdave, you can execute the following commands:

```bash
export VCPKG_FORCE_SYSTEM_BINARIES=1
export CC=/usr/bin/gcc CXX=/usr/bin/g++
export CXXFLAGS="-Wno-error=maybe-uninitialized"

# Install necessary packages
apk add build-base cmake ninja zip unzip curl git pkgconfig perl nams go

# FORCE_BUILD=1 as Discord do not provide pre-built binaries
FORCE_BUILD=1 ./libdave_install.sh v1.1.0
```

### Windows instructions

Open Powershell and execute the following commands:

```ps1
Set-ExecutionPolicy RemoteSigned –Scope Process
.\libdave_install.ps1 v1.1.0
```

## Example Usage

For an example of how to use GoDave, please see [here](https://github.com/disgoorg/disgo/tree/feature/dave/_examples/voice)

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE). See LICENSE for more information.
