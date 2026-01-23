package libdave

// FIXME: Consider https://pkg.go.dev/cmd/cgo#hdr-Optimizing_calls_of_C_code

// #cgo pkg-config: dave
// #include "dave.h"
import "C"

// MaxSupportedProtocolVersion returns the maximum supported libdave protocol version.
func MaxSupportedProtocolVersion() uint16 {
	return uint16(C.daveMaxSupportedProtocolVersion())
}
