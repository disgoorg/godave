[![Go Reference](https://pkg.go.dev/badge/github.com/disgoorg/godave.svg)](https://pkg.go.dev/github.com/disgoorg/godave)
[![Go Report](https://goreportcard.com/badge/github.com/disgoorg/godave)](https://goreportcard.com/report/github.com/disgoorg/godave)
[![Go Version](https://img.shields.io/github/go-mod/go-version/disgoorg/godave)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![GoDave Version](https://img.shields.io/github/v/tag/disgoorg/godave?label=release)](https://github.com/disgoorg/godave/releases/latest)
[![DisGo Discord](https://discord.com/api/guilds/817327181659111454/widget.png)](https://discord.gg/TewhTfDpvW)

<img align="right" src="/.github/godave_gopher.png" width=192 alt="discord gopher">

# GoDave

GoDave is a library that provides Go bindings for [libdave](https://github.com/discord/libdave) and provides a generic DAVE interface allowing for different implementations in the future.

## Summary
1. [Libdave Installation](#libdave-installation)
2. [Installation Script (Recommended)](#installation-script-recommended)
3. [Manual Build](#manual-build)
4. [Example Usage](#example-usage)
4. [License](#license)]()

## Libdave Installation

This library uses CGO and dynamic linking to use libdave. As such, it needs to be installed in the system beforehand
to build this library.

> [!NOTE]
> Due to the nature of this project, it might be necessary to re-install libdave when updating to a new GoDave version.
> 
> Versions requiring this will be denoted with a bump in the major version (for reference: major.minor.patch).

### Installation Script (Recommended)

We provide helpful scripts in [scripts/](https://github.com/disgoorg/godave/tree/master/scripts) to simplify installing
a compatible libdave version. Grab whichever one is applicable to your OS (`.sh` for Linux and MacOS; `ps1` for
Windows PowerShell) and (after auditing its contents) run it and follow any instructions it might output.

Once that step is complete, you can continue with the installation of GoDave.

### Manual Build

For a manual build, please clone https://github.com/discord/libdave and use revision
`74979cb33febf4ddef0c2b66e57520b339550c17`.

> [!NOTE]
> We provide no guarantees for this version of GoDave to run for other revisions other than that the one mentioned above.
> 
> As the library evolves and new versions of libdave are released, the above revision will be updated to match the
> GoDave version

Once checked out, please follow the
[build instructions](https://github.com/discord/libdave/tree/74979cb33febf4ddef0c2b66e57520b339550c17/cpp#building) and
setup the appropriate `pkg-config` file and configuration to allow for discovery at compilation time.

## Example Usage

For an example of how to use GoDave, please refer to the [here](https://github.com/disgoorg/disgo/tree/feature/dave/_examples/voice)

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE). See LICENSE for more information.
