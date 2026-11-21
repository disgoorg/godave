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

// A frame from a user without a decryptor is passed through unchanged. The
// copy into the output buffer must not run backwards, or the caller gets
// silence and the input frame is overwritten.
func TestDecryptPassesThroughFramesFromUnknownUsers(t *testing.T) {
	s := newTestSession()

	frame := []byte{0xF8, 0xFF, 0xFE}
	out := make([]byte, s.MaxDecryptedFrameSize("2", len(frame)))

	n, err := s.Decrypt("2", frame, out)
	if err != nil || !bytes.Equal(out[:n], []byte{0xF8, 0xFF, 0xFE}) {
		t.Errorf("expected the frame passed through, got %x, %v", out[:n], err)
	}
	if !bytes.Equal(frame, []byte{0xF8, 0xFF, 0xFE}) {
		t.Errorf("input frame was overwritten: %x", frame)
	}
}

// Callers size the output by capacity, as the decryptor path does: disgo grows
// its buffer to MaxDecryptedFrameSize with slices.Grow, which leaves the length
// where it was. A passthrough frame longer than that length must still come
// back whole.
func TestDecryptPassesThroughIntoCapacity(t *testing.T) {
	s := newTestSession()

	frame := bytes.Repeat([]byte{0xAB}, 900)
	out := make([]byte, 512, s.MaxDecryptedFrameSize("2", len(frame)))

	n, err := s.Decrypt("2", frame, out)
	if err != nil || n != len(frame) || !bytes.Equal(out[:n], frame) {
		t.Errorf("expected the %d-byte frame passed through, got %d bytes, %v", len(frame), n, err)
	}
}
