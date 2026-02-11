package libdave

// #include "dave.h"
// extern void godaveGlobalLogCallback(DAVELoggingSeverity severity, char* file, int line, char* message);
import "C"
import (
	"context"
	"log/slog"
	"sync/atomic"
	"unsafe"
)

var (
	logLoggerLevel slog.LevelVar
	defaultLogger  atomic.Pointer[slog.Logger]
)

func init() {
	SetDefaultLogLoggerLevel(slog.LevelError)
	SetDefaultLogger(slog.New(newLogWrapper(slog.Default().Handler())).
		With(slog.String("name", "libdave")),
	)

	C.daveSetLogSinkCallback(C.DAVELogSinkCallback(unsafe.Pointer(C.godaveGlobalLogCallback)))
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

	defaultLogger.Load().Log(context.Background(), slogSeverity, C.GoString(message), slog.String("file", C.GoString(file)), slog.Int("line", int(line)))
}

// SetDefaultLogger sets the default logger used by libdave.
func SetDefaultLogger(logger *slog.Logger) {
	defaultLogger.Store(logger)
}

// SetDefaultLogLoggerLevel sets the log level for libdave logs.
// By default, the level is set to slog.LevelError.
// It returns the previous log level.
func SetDefaultLogLoggerLevel(level slog.Level) (oldLevel slog.Level) {
	oldLevel = logLoggerLevel.Level()
	logLoggerLevel.Set(level)
	return
}

var _ slog.Handler = (*logWrapper)(nil)

// newLogWrapper wraps the default slog.Handler and only enables logs at or above the given level.
func newLogWrapper(handler slog.Handler) *logWrapper {
	return &logWrapper{
		handler: handler,
	}
}

type logWrapper struct {
	handler slog.Handler
}

func (l *logWrapper) Enabled(_ context.Context, level slog.Level) bool {
	return level >= logLoggerLevel.Level()
}

func (l *logWrapper) Handle(ctx context.Context, record slog.Record) error {
	return l.handler.Handle(ctx, record)
}

func (l logWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logWrapper{
		handler: l.handler.WithAttrs(attrs),
	}
}

func (l logWrapper) WithGroup(name string) slog.Handler {
	return &logWrapper{
		handler: l.handler.WithGroup(name),
	}
}
