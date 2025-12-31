package godave

import (
	"log/slog"

	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/godave/libdave"
)

const (
	initTransitionId         = 0
	disabledProtocolVersion  = 0
	mlsNewGroupExpectedEpoch = 1
)

type Session struct {
	selfUserID          snowflake.ID
	channelID           snowflake.ID
	callbacks           voice.Callbacks
	logger              *slog.Logger
	mlsSession          *libdave.Session
	encryptor           *libdave.Encryptor
	decryptors          map[snowflake.ID]*libdave.Decryptor
	preparedTransitions map[uint16]uint16
}

func NewSession(selfUserID snowflake.ID, callbacks voice.Callbacks) *Session {
	m := Session{
		selfUserID: selfUserID,
		callbacks:  callbacks,
		// Context and authSessionID are only used with persistent key storage and can be ignored most of the time
		mlsSession:          libdave.NewSession("", ""),
		encryptor:           libdave.NewEncryptor(),
		decryptors:          make(map[snowflake.ID]*libdave.Decryptor),
		preparedTransitions: make(map[uint16]uint16),
	}

	// Start in Passthrough by default
	m.encryptor.SetPassthroughMode(true)

	return &m
}

func (m *Session) MaxSupportedProtocolVersion() int {
	return int(libdave.MaxSupportedProtocolVersion())
}

func (m *Session) SetChannelID(channelID snowflake.ID) {
	m.channelID = channelID
}

func (m *Session) AssignSsrcToCodec(codec int, ssrc uint32) {
	m.encryptor.AssignSsrcToCodec(ssrc, libdave.Codec(codec))
}

func (m *Session) EncryptOpus(ssrc uint32, frame []byte) ([]byte, error) {
	return m.encryptor.Encrypt(libdave.MediaTypeAudio, ssrc, frame)
}

func (m *Session) DecryptOpus(userID snowflake.ID, frame []byte) ([]byte, error) {
	if decryptor, ok := m.decryptors[userID]; ok {
		return decryptor.Decrypt(libdave.MediaTypeAudio, frame)
	}

	// Assume passthrough
	return frame, nil
}

func (m *Session) AddUser(userID snowflake.ID) {
	m.decryptors[userID] = libdave.NewDecryptor()
	m.setupKeyRatchetForUser(userID, m.mlsSession.GetProtocolVersion())
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
			globalLogger.Error("failed to send ready for transition", slog.Any("error", err))
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
	m.mlsSession.SetExternalSender(externalSenderPackage)
}

func (m *Session) OnDaveMLSProposals(proposals []byte) {
	commitWelcome := m.mlsSession.ProcessProposals(proposals, m.recognizedUserIDs())

	if commitWelcome != nil {
		m.sendMLSCommitWelcome(commitWelcome)
	}
}

func (m *Session) OnDaveMLSPrepareCommitTransition(transitionID uint16, commitMessage []byte) {
	res := m.mlsSession.ProcessCommit(commitMessage)

	if res.IsIgnored() {
		return
	}

	if res.IsFailed() {
		m.sendInvalidCommitWelcome(transitionID)
		m.protocolInit(m.mlsSession.GetProtocolVersion())
		return
	}

	m.prepareTransition(transitionID, m.mlsSession.GetProtocolVersion())
	if transitionID != initTransitionId {
		m.sendReadyForTransition(transitionID)
	}
}

func (m *Session) OnDaveMLSWelcome(transitionID uint16, welcomeMessage []byte) {
	res := m.mlsSession.ProcessWelcome(welcomeMessage, m.recognizedUserIDs())

	if res == nil {
		m.sendInvalidCommitWelcome(transitionID)
		m.sendMLSKeyPackage()
		return
	}

	m.prepareTransition(transitionID, m.mlsSession.GetProtocolVersion())
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

	m.mlsSession.Init(protocolVersion, uint64(m.channelID), m.selfUserID.String())
}

func (m *Session) executeTransition(transitionID uint16) {
	protocolVersion, ok := m.preparedTransitions[transitionID]
	if !ok {
		return
	}

	delete(m.preparedTransitions, transitionID)

	if protocolVersion == disabledProtocolVersion {
		m.mlsSession.Reset()
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
	//fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!%d == %d -> %t (userID: %s)\n", protocolVersion, disabledProtocolVersion, disabled, userID.String())

	if userID == m.selfUserID {
		m.encryptor.SetPassthroughMode(disabled)
		if !disabled {
			m.encryptor.SetKeyRatchet(m.mlsSession.GetKeyRatchet(userID.String()))
		}
		return
	}

	decryptor := m.decryptors[userID]
	decryptor.TransitionToPassthroughMode(disabled)
	if !disabled {
		decryptor.TransitionToKeyRatchet(m.mlsSession.GetKeyRatchet(userID.String()))
	}
}

func (m *Session) sendMLSKeyPackage() {
	err := m.callbacks.SendMLSKeyPackage(m.mlsSession.GetMarshalledKeyPackage())
	if err != nil {
		globalLogger.Error("failed to send MLS key package", slog.Any("error", err))
	}
}

func (m *Session) sendMLSCommitWelcome(message []byte) {
	err := m.callbacks.SendMLSCommitWelcome(message)
	if err != nil {
		globalLogger.Error("failed to send commit welcome", slog.Any("error", err))
	}
}

func (m *Session) sendReadyForTransition(transitionID uint16) {
	err := m.callbacks.SendReadyForTransition(transitionID)
	if err != nil {
		globalLogger.Error("failed to send ready for transition", slog.Any("error", err))
	}
}

func (m *Session) sendInvalidCommitWelcome(transitionID uint16) {
	err := m.callbacks.SendInvalidCommitWelcome(transitionID)
	if err != nil {
		globalLogger.Error("failed to send invalid commit welcome", slog.Any("error", err))
	}
}
