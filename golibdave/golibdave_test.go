package golibdave

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/disgoorg/godave"
)

type testCallbacks struct{}

func (testCallbacks) SendMLSKeyPackage([]byte) error        { return nil }
func (testCallbacks) SendMLSCommitWelcome([]byte) error     { return nil }
func (testCallbacks) SendReadyForTransition(uint16) error   { return nil }
func (testCallbacks) SendInvalidCommitWelcome(uint16) error { return nil }

func newTestSession() godave.Session {
	return NewSession(slog.New(slog.DiscardHandler), "1", testCallbacks{})
}

func TestReadyWhenProtocolDisabled(t *testing.T) {
	s := newTestSession()
	s.OnSelectProtocolAck(0)

	if !s.Ready() {
		t.Error("expected Ready to be true when the protocol is disabled")
	}

	frame := []byte{0xF8, 0xFF, 0xFE}
	out := make([]byte, s.MaxEncryptedFrameSize(len(frame)))

	n, err := s.Encrypt(1, frame, out)
	if err != nil || !bytes.Equal(out[:n], frame) {
		t.Errorf("expected passthrough encryption, got %x, %v", out[:n], err)
	}
}

func TestNotReadyBeforeEpoch(t *testing.T) {
	s := newTestSession()
	s.OnSelectProtocolAck(1)

	if s.Ready() {
		t.Error("expected Ready to be false before an epoch is established")
	}
}

func TestReadyAfterDowngradeExecutes(t *testing.T) {
	s := newTestSession()
	s.OnSelectProtocolAck(1)
	s.OnDavePrepareTransition(5, 0)

	if s.Ready() {
		t.Error("expected Ready to be false before the downgrade executes")
	}

	s.OnDaveExecuteTransition(5)

	if !s.Ready() {
		t.Error("expected Ready to be true once the downgrade executes")
	}
}
