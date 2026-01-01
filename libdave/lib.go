package libdave

import "C"
import (
	"context"
	"log/slog"
	"unsafe"
)

//// #cgo linux,386  LDFLAGS: -L${SRCDIR}/build/linux -ldave_x86
// #cgo linux,amd64  LDFLAGS: -L${SRCDIR}/build/linux -ldave_x86-64
// #cgo linux,arm64  LDFLAGS: -L${SRCDIR}/build/linux -ldave_aarch64
//
// #cgo darwin,amd64  LDFLAGS: -L${SRCDIR}/build/darwin -ldave_x86-64
// #cgo darwin,arm64  LDFLAGS: -L${SRCDIR}/build/darwin -ldave_aarch64e
//
//// #cgo windows,386 LDFLAGS: -L${SRCDIR}/build/windows -ldave_x86
//// #cgo windows,amd64 LDFLAGS: -L${SRCDIR}/build/windows -ldave_x86-64
//// #cgo windows,arm64 LDFLAGS: -L${SRCDIR}/build/windows -ldave_aarch64
// #include "dave.h"
// extern void godaveGlobalLogCallback(DAVELoggingSeverity severity, char* file, int line, char* message);
import "C"

var globalLogger = slog.Default().With("name", "libdave")

func MaxSupportedProtocolVersion() uint16 {
	return uint16(C.daveMaxSupportedProtocolVersion())
}

func SetLogger(logger *slog.Logger) {
	globalLogger = logger
}

//export godaveGlobalLogCallback
func godaveGlobalLogCallback(severity C.DAVELoggingSeverity, file *C.char, line C.int, message *C.char) {
	var slogSeverity slog.Level
	switch severity {
	case C.DAVE_LOGGING_SEVERITY_VERBOSE:
		slogSeverity = slog.LevelDebug
	case C.DAVE_LOGGING_SEVERITY_INFO:
		slogSeverity = slog.LevelInfo
	case C.DAVE_LOGGING_SEVERITY_WARNING:
		slogSeverity = slog.LevelWarn
	case C.DAVE_LOGGING_SEVERITY_ERROR:
		slogSeverity = slog.LevelError
	case C.DAVE_LOGGING_SEVERITY_NONE:
		return
	}

	globalLogger.Log(context.Background(), slogSeverity, C.GoString(message), slog.String("file", C.GoString(file)), slog.Int("line", int(line)))
}

func init() {
	C.daveSetLogSinkCallback(C.DAVELogSinkCallback(unsafe.Pointer(C.godaveGlobalLogCallback)))
}
