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
    1. [Installation Script (Recommended)](#installation-script-recommended)
    2. [Manual Build](#manual-build)
2. [Example Usage](#example-usage)
3. [License](#license)

## Libdave Installation

This library uses CGO and dynamic linking to use libdave. As such, it needs to be installed in the system beforehand
to build this library.

> [!NOTE]
> Due to the nature of this project, it might be necessary to re-install libdave when updating to a new GoDave version.
>
> If you have compilation errors, please ensure that you have installed a compatible libdave version.
> 
> Versions requiring this will be denoted with a bump in the major version (for reference: major.minor.patch).

### Installation Script (Recommended)

We provide helpful [libdave_install.sh](https://github.com/disgoorg/godave/tree/master/scripts/libdave_install.sh) script
to simplify installing a compatible libdave version. After auditing the contents of the script, run it and follow any
instructions it might output.

A simple one-liner would be:

```bash
curl https://raw.githubusercontent.com/disgoorg/godave/refs/heads/master/scripts/libdave_install.sh | bash
```

Once you have successfully installed the shared libdave library, you can continue with the installation of GoDave.

### Manual Build

For a manual build, please clone https://github.com/discord/libdave and use revision
`d6874165b9a7c8d2cc59712c7aceaa8dffb189b4`.

> [!NOTE]
> We provide no guarantees for this version of GoDave to run for other revisions other than that the one mentioned above.
> 
> As the library evolves and new versions of libdave are released, the above revision will be updated to match the
> GoDave version

Once checked out, please follow the
[build instructions](https://github.com/discord/libdave/tree/d6874165b9a7c8d2cc59712c7aceaa8dffb189b4/cpp#building) and
setup the appropriate `pkg-config` file and configuration to allow for discovery at compilation time.

## Example Usage

For an example of how to use GoDave, please see [here](https://github.com/disgoorg/disgo/tree/feature/dave/_examples/voice)

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE). See LICENSE for more information.
