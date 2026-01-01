package godave

import (
	"log/slog"

	"github.com/disgoorg/snowflake/v2"
)

func CreateNoopSession(userID snowflake.ID, callbacks Callbacks) Session {
	slog.Warn("Using passthrough dave session. Please migrate to an implementation of libdave or your audio connections will stop working on 01.03.2026")

	return &noopSession{}
}

type noopSession struct{}

func (n *noopSession) MaxSupportedProtocolVersion() int {
	return 0
}
func (n *noopSession) Encrypt(ssrc uint32, frame []byte) ([]byte, error) {
	return frame, nil
}
func (n *noopSession) Decrypt(userID snowflake.ID, frame []byte) ([]byte, error) {
	return frame, nil
}
func (n *noopSession) SetChannelID(channelID snowflake.ID)                                 {}
func (n *noopSession) AssignSsrcToCodec(ssrc uint32, codec Codec)                          {}
func (n *noopSession) AddUser(userID snowflake.ID)                                         {}
func (n *noopSession) RemoveUser(userID snowflake.ID)                                      {}
func (n *noopSession) OnSelectProtocolAck(protocolVersion uint16)                          {}
func (n *noopSession) OnDavePrepareTransition(transitionID uint16, protocolVersion uint16) {}
func (n *noopSession) OnDaveExecuteTransition(protocolVersion uint16)                      {}
func (n *noopSession) OnDavePrepareEpoch(epoch int, protocolVersion uint16)                {}
func (n *noopSession) OnDaveMLSExternalSenderPackage(externalSenderPackage []byte)         {}
func (n *noopSession) OnDaveMLSProposals(proposals []byte)                                 {}
func (n *noopSession) OnDaveMLSPrepareCommitTransition(transitionID uint16, commit []byte) {}
func (n *noopSession) OnDaveMLSWelcome(transitionID uint16, welcomeMessage []byte)         {}
