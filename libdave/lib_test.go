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

func TestRecognizedUsersCanBeAddedAndRemoved(t *testing.T) {
	session := NewSession("", "")
	session.AddRecognizedUser("1")
	session.AddRecognizedUser("2")
	session.RemoveRecognizedUser("1")
	session.RemoveRecognizedUser("unknown")
}
