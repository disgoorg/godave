package godave

import (
	"log/slog"
)

var (
	_ SessionCreateFunc = NewNoopSession
	_ Session           = (*noopSession)(nil)
)

func NewNoopSession(logger *slog.Logger, _ UserID, _ Callbacks) Session {
	logger.Warn("Using noop dave session. Please migrate to an implementation of libdave or your audio connections will stop working on 01.03.2026")

	return &noopSession{}
}

type noopSession struct{}

func (n *noopSession) MaxSupportedProtocolVersion() int {
	return 0
}
func (n *noopSession) MaxEncryptedFrameSize(frameSize int) int {
	return frameSize
}
func (n *noopSession) Encrypt(_ uint32, frame []byte, encryptedFrame []byte) (int, error) {
	return copy(encryptedFrame, frame), nil
}

func (n *noopSession) MaxDecryptedFrameSize(_ UserID, frameSize int) int {
	return frameSize
}
func (n *noopSession) Decrypt(_ UserID, frame []byte, decryptedFrame []byte) (int, error) {
	return copy(decryptedFrame, frame), nil
}
func (n *noopSession) SetChannelID(_ ChannelID)                            {}
func (n *noopSession) AssignSsrcToCodec(_ uint32, _ Codec)                 {}
func (n *noopSession) AddUser(_ UserID)                                    {}
func (n *noopSession) RemoveUser(_ UserID)                                 {}
func (n *noopSession) OnSelectProtocolAck(_ uint16)                        {}
func (n *noopSession) OnDavePrepareTransition(_ uint16, _ uint16)          {}
func (n *noopSession) OnDaveExecuteTransition(_ uint16)                    {}
func (n *noopSession) OnDavePrepareEpoch(_ int, _ uint16)                  {}
func (n *noopSession) OnDaveMLSExternalSenderPackage(_ []byte)             {}
func (n *noopSession) OnDaveMLSProposals(_ []byte)                         {}
func (n *noopSession) OnDaveMLSPrepareCommitTransition(_ uint16, _ []byte) {}
func (n *noopSession) OnDaveMLSWelcome(_ uint16, _ []byte)                 {}
