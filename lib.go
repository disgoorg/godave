package godave

import (
	"log/slog"

	"github.com/disgoorg/godave/libdave"
)

var globalLogger = slog.Default().With("name", "godave")

func init() {
	libdave.SetLogger(globalLogger)
}
