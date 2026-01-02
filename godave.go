package godave

import (
	"log/slog"
)

// SessionCreateFunc is an agnostic function type for creating DAVE sessions.
type SessionCreateFunc func(logger *slog.Logger, userID UserID, callbacks Callbacks) Session

// Callbacks represents the callbacks used by a DAVE session to send messages back to the voice gateway.
type Callbacks interface {
	// SendMLSKeyPackage sends a MLS Key Package to the voice gateway.
	SendMLSKeyPackage(mlsKeyPackage []byte) error
	// SendMLSCommitWelcome sends a MLS Commit Welcome to the voice gateway.
	SendMLSCommitWelcome(mlsCommitWelcome []byte) error
	// SendReadyForTransition notifies the voice gateway that the client is ready for the transition.
	SendReadyForTransition(transitionID uint16) error
	// SendInvalidCommitWelcome notifies the voice gateway that the commit welcome is invalid.
	SendInvalidCommitWelcome(transitionID uint16) error
}

type UserID string

type ChannelID uint64

// Codec represents an audio codec used in the DAVE protocol.
type Codec int

const (
	// CodecOpus represents the OPUS audio codec.
	CodecOpus Codec = 1
)

// Session is an interface representing a DAVE session.
// Implementations of this interface should handle encryption, decryption, and DAVE protocol events.
type Session interface {
	// MaxSupportedProtocolVersion returns the maximum supported DAVE version for this session.
	MaxSupportedProtocolVersion() int

	// SetChannelID sets the channel ID for this session.
	SetChannelID(channelID ChannelID)

	// AssignSsrcToCodec maps a given SSRC to a specific Codec.
	AssignSsrcToCodec(ssrc uint32, codec Codec)

	// MaxEncryptedFrameSize returns the maximum size of an encrypted OPUS frame given the frame size.
	MaxEncryptedFrameSize(frameSize int) int

	// Encrypt encrypts an OPUS frame.
	Encrypt(ssrc uint32, frame []byte, encryptedFrame []byte) (int, error)

	// MaxDecryptedFrameSize returns the maximum size of a decrypted OPUS frame given the frame size.
	MaxDecryptedFrameSize(userID UserID, frameSize int) int

	// Decrypt decrypts an OPUS frame.
	Decrypt(userID UserID, frame []byte, decryptedFrame []byte) (int, error)

	// AddUser adds a user to the MLS group.
	AddUser(userID UserID)

	// RemoveUser removes a use from the MLS group.
	RemoveUser(userID UserID)

	// OnSelectProtocolAck is to be called when SELECT_PROTOCOL_ACK (4) is received.
	OnSelectProtocolAck(protocolVersion uint16)

	// OnDavePrepareTransition is to be called when DAVE_PROTOCOL_PREPARE_TRANSITION (21) is received.
	OnDavePrepareTransition(transitionID uint16, protocolVersion uint16)

	// OnDaveExecuteTransition is to be called when DAVE_PROTOCOL_EXECUTE_TRANSITION (22) is received.
	OnDaveExecuteTransition(protocolVersion uint16)

	// OnDavePrepareEpoch is to be called when DAVE_PROTOCOL_PREPARE_EPOCH (24) is received.
	OnDavePrepareEpoch(epoch int, protocolVersion uint16)

	// OnDaveMLSExternalSenderPackage is to be called when DAVE_MLS_EXTERNAL_SENDER_PACKAGE (25) is received.
	OnDaveMLSExternalSenderPackage(externalSenderPackage []byte)

	// OnDaveMLSProposals is to be called when DAVE_MLS_PROPOSALS (27) is received.
	OnDaveMLSProposals(proposals []byte)

	// OnDaveMLSPrepareCommitTransition is to be called when DAVE_MLS_ANNOUNCE_COMMIT_TRANSITION (29) is received.
	OnDaveMLSPrepareCommitTransition(transitionID uint16, commitMessage []byte)

	// OnDaveMLSWelcome is to be called when DAVE_MLS_WELCOME (30) is received.
	OnDaveMLSWelcome(transitionID uint16, welcomeMessage []byte)
}
