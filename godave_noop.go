package godave

import (
	"log/slog"

	"github.com/disgoorg/snowflake/v2"
)

var (
	_ SessionCreate = NewNoopSession
	_ Session       = (*noopSession)(nil)
)

func NewNoopSession(logger *slog.Logger, _ snowflake.ID, _ Callbacks) Session {
	logger.Warn("Using noop dave session. Please migrate to an implementation of libdave or your audio connections will stop working on 01.03.2026")

	return &noopSession{}
}

type noopSession struct{}

func (n *noopSession) MaxSupportedProtocolVersion() int {
	return 0
}
func (n *noopSession) Encrypt(_ uint32, frame []byte) ([]byte, error) {
	return frame, nil
}
func (n *noopSession) Decrypt(_ snowflake.ID, frame []byte) ([]byte, error) {
	return frame, nil
}
func (n *noopSession) SetChannelID(_ snowflake.ID)                         {}
func (n *noopSession) AssignSsrcToCodec(_ uint32, _ Codec)                 {}
func (n *noopSession) AddUser(_ snowflake.ID)                              {}
func (n *noopSession) RemoveUser(_ snowflake.ID)                           {}
func (n *noopSession) OnSelectProtocolAck(_ uint16)                        {}
func (n *noopSession) OnDavePrepareTransition(_ uint16, _ uint16)          {}
func (n *noopSession) OnDaveExecuteTransition(_ uint16)                    {}
func (n *noopSession) OnDavePrepareEpoch(_ int, _ uint16)                  {}
func (n *noopSession) OnDaveMLSExternalSenderPackage(_ []byte)             {}
func (n *noopSession) OnDaveMLSProposals(_ []byte)                         {}
func (n *noopSession) OnDaveMLSPrepareCommitTransition(_ uint16, _ []byte) {}
func (n *noopSession) OnDaveMLSWelcome(_ uint16, _ []byte)                 {}
