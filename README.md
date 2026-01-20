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

This library uses CGO and dynamic linking to use libdave. We automatically pull the latest libdave version and link
against the [shared libraries published by Discord](https://github.com/discord/libdave/releases).

As such, if you are using an operating system and architecture combination which does not have any pre-built libraries,
you will have to do a bit more tinkering (see bellow)

If you are using an operating system which is covered under the releases provided by Discord (and you are not on Windows)
then there is nothing else for you to  do, you can use GoDave directly!

### Windows instructions

If you are using Windows, the binaries provided by Discord are actually a static library, so you will have to download
and extract the `libdave.lib` and place it next to your final executable.

### Manual Installation
If you know CGO, you can manually build and set the correct compiler flags for it to be resolved correctly. For those
who don't know CGO, don't worry, we have got you covered!

To build libdave, we recommend using our [libdave_build.sh](https://github.com/disgoorg/godave/tree/master/scripts/libdave_build.sh).
After auditing its contents, you can download it and execute it like this:

```bash
bash libdave_build.sh <version>
```

After it is done building, add the following to your `main.go` file:

```go
// #cgo pkg-config: dave
import "C"
```

With this, you should now be able to use libdave.

> [!NOTE]
> Due to the nature of this project, it might be necessary to re-install libdave when updating to a new GoDave version.
>
> You can see what version is required by checking [this file](https://github.com/disgoorg/godave/tree/master/libdave/lib/release.txt)

## Example Usage

For an example of how to use GoDave, please see [here](https://github.com/disgoorg/disgo/tree/feature/dave/_examples/voice)

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE). See LICENSE for more information.
