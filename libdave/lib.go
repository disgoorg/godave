package libdave

// FIXME: Consider https://pkg.go.dev/cmd/cgo#hdr-Optimizing_calls_of_C_code

// #cgo linux,amd64 LDFLAGS: -L${SRCDIR}/vendor/lib/linux_x64 -ldave -Wl,-rpath,${SRCDIR}/vendor/lib/linux_x64
//
// #cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/vendor/lib/macos_x64 -ldave -Wl,-rpath,${SRCDIR}/vendor/lib/macos_x64
// #cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/vendor/lib/macos_arm64 -ldave -Wl,-rpath,${SRCDIR}/vendor/lib/macos_arm64
//
// #cgo windows,amd64 LDFLAGS: -L${SRCDIR}/vendor/lib/win_x64 -ldave
// #include "vendor/include/dave.h"
import "C"

// MaxSupportedProtocolVersion returns the maximum supported libdave protocol version.
func MaxSupportedProtocolVersion() uint16 {
	return uint16(C.daveMaxSupportedProtocolVersion())
}
