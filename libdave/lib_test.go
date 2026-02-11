package libdave

import (
	"testing"
)

func TestMaxSupportedProtocolVersion(t *testing.T) {
	maxSupportedProtocolVersion := MaxSupportedProtocolVersion()

	if maxSupportedProtocolVersion != 1 {
		t.Errorf("expected 1, got %d", maxSupportedProtocolVersion)
	}
}
