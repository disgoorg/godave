package golibdave

import (
	"log/slog"

	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/godave"

	"github.com/disgoorg/godave/libdave"
)

const (
	initTransitionId         = 0
	disabledProtocolVersion  = 0
	mlsNewGroupExpectedEpoch = 1
)

var (
	_ godave.SessionCreateFunc = NewSession
	_ godave.Session           = (*Session)(nil)
)

// NewSession returns a new DAVE session using libdave.
func NewSession(logger *slog.Logger, selfUserID snowflake.ID, callbacks godave.Callbacks) godave.Session {
	encryptor := libdave.NewEncryptor()
	// Start in Passthrough by default
	encryptor.SetPassthroughMode(true)

	return &Session{
		selfUserID: selfUserID,
		callbacks:  callbacks,
		logger:     logger,
		// Context and authSessionID are only used with persistent key storage and can be ignored most of the time
		session:             libdave.NewSession("", ""),
		encryptor:           encryptor,
		decryptors:          make(map[snowflake.ID]*libdave.Decryptor),
		preparedTransitions: make(map[uint16]uint16),
	}
}

type Session struct {
	selfUserID          snowflake.ID
	channelID           snowflake.ID
	logger              *slog.Logger
	callbacks           godave.Callbacks
	session             *libdave.Session
	encryptor           *libdave.Encryptor
	decryptors          map[snowflake.ID]*libdave.Decryptor
	preparedTransitions map[uint16]uint16
}

func (m *Session) MaxSupportedProtocolVersion() int {
	return int(libdave.MaxSupportedProtocolVersion())
}

func (m *Session) SetChannelID(channelID snowflake.ID) {
	m.channelID = channelID
}

func (m *Session) AssignSsrcToCodec(ssrc uint32, codec godave.Codec) {
	m.encryptor.AssignSsrcToCodec(ssrc, libdave.Codec(codec))
}

func (m *Session) Encrypt(ssrc uint32, frame []byte) ([]byte, error) {
	return m.encryptor.Encrypt(libdave.MediaTypeAudio, ssrc, frame)
}

func (m *Session) Decrypt(userID snowflake.ID, frame []byte) ([]byte, error) {
	if decryptor, ok := m.decryptors[userID]; ok {
		return decryptor.Decrypt(libdave.MediaTypeAudio, frame)
	}

	// Assume passthrough
	return frame, nil
}

func (m *Session) AddUser(userID snowflake.ID) {
	m.decryptors[userID] = libdave.NewDecryptor()
	m.setupKeyRatchetForUser(userID, m.session.GetProtocolVersion())
}

func (m *Session) RemoveUser(userID snowflake.ID) {
	delete(m.decryptors, userID)
}

func (m *Session) OnSelectProtocolAck(protocolVersion uint16) {
	m.protocolInit(protocolVersion)
}

func (m *Session) OnDavePrepareTransition(transitionID uint16, protocolVersion uint16) {
	if _, ok := m.preparedTransitions[transitionID]; ok {

	}
	m.prepareTransition(transitionID, protocolVersion)

	if transitionID != initTransitionId {
		err := m.callbacks.SendReadyForTransition(transitionID)
		if err != nil {
			m.logger.Error("failed to send ready for transition", slog.Any("error", err))
		}
	}
}

func (m *Session) OnDaveExecuteTransition(transitionID uint16) {
	m.executeTransition(transitionID)
}

func (m *Session) OnDavePrepareEpoch(epoch int, protocolVersion uint16) {
	m.prepareEpoch(epoch, protocolVersion)

	if epoch == mlsNewGroupExpectedEpoch {
		m.sendMLSKeyPackage()
	}
}

func (m *Session) OnDaveMLSExternalSenderPackage(externalSenderPackage []byte) {
	m.session.SetExternalSender(externalSenderPackage)
}

func (m *Session) OnDaveMLSProposals(proposals []byte) {
	commitWelcome := m.session.ProcessProposals(proposals, m.recognizedUserIDs())

	if commitWelcome != nil {
		m.sendMLSCommitWelcome(commitWelcome)
	}
}

func (m *Session) OnDaveMLSPrepareCommitTransition(transitionID uint16, commitMessage []byte) {
	res := m.session.ProcessCommit(commitMessage)

	if res.IsIgnored() {
		return
	}

	if res.IsFailed() {
		m.sendInvalidCommitWelcome(transitionID)
		m.protocolInit(m.session.GetProtocolVersion())
		return
	}

	m.prepareTransition(transitionID, m.session.GetProtocolVersion())
	if transitionID != initTransitionId {
		m.sendReadyForTransition(transitionID)
	}
}

func (m *Session) OnDaveMLSWelcome(transitionID uint16, welcomeMessage []byte) {
	res := m.session.ProcessWelcome(welcomeMessage, m.recognizedUserIDs())

	if res == nil {
		m.sendInvalidCommitWelcome(transitionID)
		m.sendMLSKeyPackage()
		return
	}

	m.prepareTransition(transitionID, m.session.GetProtocolVersion())
	if transitionID != initTransitionId {
		m.sendReadyForTransition(transitionID)
	}
}

func (m *Session) recognizedUserIDs() []string {
	userIDs := make([]string, 0, len(m.decryptors)+1)

	userIDs = append(userIDs, m.selfUserID.String())

	for userID, _ := range m.decryptors {
		userIDs = append(userIDs, userID.String())
	}

	return userIDs
}

func (m *Session) protocolInit(protocolVersion uint16) {
	if protocolVersion > disabledProtocolVersion {
		m.prepareEpoch(mlsNewGroupExpectedEpoch, protocolVersion)
		m.sendMLSKeyPackage()
	} else {
		m.prepareTransition(initTransitionId, protocolVersion)
		m.executeTransition(initTransitionId)
	}
}

func (m *Session) prepareEpoch(epoch int, protocolVersion uint16) {
	if epoch != mlsNewGroupExpectedEpoch {
		return
	}

	m.session.Init(protocolVersion, uint64(m.channelID), m.selfUserID.String())
}

func (m *Session) executeTransition(transitionID uint16) {
	protocolVersion, ok := m.preparedTransitions[transitionID]
	if !ok {
		return
	}

	delete(m.preparedTransitions, transitionID)

	if protocolVersion == disabledProtocolVersion {
		m.session.Reset()
	}

	m.setupKeyRatchetForUser(m.selfUserID, protocolVersion)
}

func (m *Session) prepareTransition(transitionID uint16, protocolVersion uint16) {
	for userID, _ := range m.decryptors {
		m.setupKeyRatchetForUser(userID, protocolVersion)
	}

	if transitionID == initTransitionId {
		m.setupKeyRatchetForUser(m.selfUserID, protocolVersion)
	} else {
		m.preparedTransitions[transitionID] = protocolVersion
	}
}

func (m *Session) setupKeyRatchetForUser(userID snowflake.ID, protocolVersion uint16) {
	disabled := protocolVersion == disabledProtocolVersion

	if userID == m.selfUserID {
		m.encryptor.SetPassthroughMode(disabled)
		if !disabled {
			m.encryptor.SetKeyRatchet(m.session.GetKeyRatchet(userID.String()))
		}
		return
	}

	decryptor := m.decryptors[userID]
	decryptor.TransitionToPassthroughMode(disabled)
	if !disabled {
		decryptor.TransitionToKeyRatchet(m.session.GetKeyRatchet(userID.String()))
	}
}

func (m *Session) sendMLSKeyPackage() {
	err := m.callbacks.SendMLSKeyPackage(m.session.GetMarshalledKeyPackage())
	if err != nil {
		m.logger.Error("failed to send MLS key package", slog.Any("error", err))
	}
}

func (m *Session) sendMLSCommitWelcome(message []byte) {
	err := m.callbacks.SendMLSCommitWelcome(message)
	if err != nil {
		m.logger.Error("failed to send commit welcome", slog.Any("error", err))
	}
}

func (m *Session) sendReadyForTransition(transitionID uint16) {
	err := m.callbacks.SendReadyForTransition(transitionID)
	if err != nil {
		m.logger.Error("failed to send ready for transition", slog.Any("error", err))
	}
}

func (m *Session) sendInvalidCommitWelcome(transitionID uint16) {
	err := m.callbacks.SendInvalidCommitWelcome(transitionID)
	if err != nil {
		m.logger.Error("failed to send invalid commit welcome", slog.Any("error", err))
	}
}
