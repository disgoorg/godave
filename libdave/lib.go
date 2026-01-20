package libdave

// FIXME: Consider https://pkg.go.dev/cmd/cgo#hdr-Optimizing_calls_of_C_code

// Windows unavailable for now, until we can get it to work
// #cgo windows,amd64 LDFLAGS: -L${SRCDIR}/lib/build/win_x64 -ldave

// #cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/build/linux_x64 -ldave -Wl,-rpath,${SRCDIR}/lib/build/linux_x64
//
// #cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/build/macos_x64 -ldave -Wl,-rpath,${SRCDIR}/lib/build/macos_x64
// #cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/build/macos_arm64 -ldave -Wl,-rpath,${SRCDIR}/lib/build/macos_arm64
// #include "lib/include/dave.h"
import "C"

// MaxSupportedProtocolVersion returns the maximum supported libdave protocol version.
func MaxSupportedProtocolVersion() uint16 {
	return uint16(C.daveMaxSupportedProtocolVersion())
}
