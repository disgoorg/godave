package golibdave

import (
	"log/slog"

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
func NewSession(logger *slog.Logger, selfUserID godave.UserID, callbacks godave.Callbacks) godave.Session {
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
		decryptors:          make(map[godave.UserID]*libdave.Decryptor),
		preparedTransitions: make(map[uint16]uint16),
	}
}

type Session struct {
	selfUserID          godave.UserID
	channelID           godave.ChannelID
	logger              *slog.Logger
	callbacks           godave.Callbacks
	session             *libdave.Session
	encryptor           *libdave.Encryptor
	decryptors          map[godave.UserID]*libdave.Decryptor
	preparedTransitions map[uint16]uint16
}

func (s *Session) MaxSupportedProtocolVersion() int {
	return int(libdave.MaxSupportedProtocolVersion())
}

func (s *Session) SetChannelID(channelID godave.ChannelID) {
	s.channelID = channelID
}

func (s *Session) AssignSsrcToCodec(ssrc uint32, codec godave.Codec) {
	s.encryptor.AssignSsrcToCodec(ssrc, libdave.Codec(codec))
}

func (s *Session) Encrypt(ssrc uint32, frame []byte) ([]byte, error) {
	return s.encryptor.Encrypt(libdave.MediaTypeAudio, ssrc, frame)
}

func (s *Session) Decrypt(userID godave.UserID, frame []byte) ([]byte, error) {
	if decryptor, ok := s.decryptors[userID]; ok {
		return decryptor.Decrypt(libdave.MediaTypeAudio, frame)
	}

	// Assume passthrough
	return frame, nil
}

func (s *Session) AddUser(userID godave.UserID) {
	s.decryptors[userID] = libdave.NewDecryptor()
	s.setupKeyRatchetForUser(userID, s.session.GetProtocolVersion())
}

func (s *Session) RemoveUser(userID godave.UserID) {
	delete(s.decryptors, userID)
}

func (s *Session) OnSelectProtocolAck(protocolVersion uint16) {
	s.protocolInit(protocolVersion)
}

func (s *Session) OnDavePrepareTransition(transitionID uint16, protocolVersion uint16) {
	s.prepareTransition(transitionID, protocolVersion)

	if transitionID != initTransitionId {
		s.sendReadyForTransition(transitionID)
	}
}

func (s *Session) OnDaveExecuteTransition(transitionID uint16) {
	s.executeTransition(transitionID)
}

func (s *Session) OnDavePrepareEpoch(epoch int, protocolVersion uint16) {
	s.prepareEpoch(epoch, protocolVersion)

	if epoch == mlsNewGroupExpectedEpoch {
		s.sendMLSKeyPackage()
	}
}

func (s *Session) OnDaveMLSExternalSenderPackage(externalSenderPackage []byte) {
	s.session.SetExternalSender(externalSenderPackage)
}

func (s *Session) OnDaveMLSProposals(proposals []byte) {
	commitWelcome := s.session.ProcessProposals(proposals, s.recognizedUserIDs())

	if commitWelcome != nil {
		s.sendMLSCommitWelcome(commitWelcome)
	}
}

func (s *Session) OnDaveMLSPrepareCommitTransition(transitionID uint16, commitMessage []byte) {
	res := s.session.ProcessCommit(commitMessage)

	if res.IsIgnored() {
		return
	}

	if res.IsFailed() {
		s.sendInvalidCommitWelcome(transitionID)
		s.protocolInit(s.session.GetProtocolVersion())
		return
	}

	s.prepareTransition(transitionID, s.session.GetProtocolVersion())
	if transitionID != initTransitionId {
		s.sendReadyForTransition(transitionID)
	}
}

func (s *Session) OnDaveMLSWelcome(transitionID uint16, welcomeMessage []byte) {
	res := s.session.ProcessWelcome(welcomeMessage, s.recognizedUserIDs())

	if res == nil {
		s.sendInvalidCommitWelcome(transitionID)
		s.sendMLSKeyPackage()
		return
	}

	s.prepareTransition(transitionID, s.session.GetProtocolVersion())
	if transitionID != initTransitionId {
		s.sendReadyForTransition(transitionID)
	}
}

func (s *Session) recognizedUserIDs() []string {
	userIDs := make([]string, 0, len(s.decryptors)+1)

	userIDs = append(userIDs, string(s.selfUserID))

	for userID := range s.decryptors {
		userIDs = append(userIDs, string(userID))
	}

	return userIDs
}

func (s *Session) protocolInit(protocolVersion uint16) {
	if protocolVersion > disabledProtocolVersion {
		s.prepareEpoch(mlsNewGroupExpectedEpoch, protocolVersion)
		s.sendMLSKeyPackage()
	} else {
		s.prepareTransition(initTransitionId, protocolVersion)
		s.executeTransition(initTransitionId)
	}
}

func (s *Session) prepareEpoch(epoch int, protocolVersion uint16) {
	if epoch != mlsNewGroupExpectedEpoch {
		return
	}

	s.session.Init(protocolVersion, uint64(s.channelID), string(s.selfUserID))
}

func (s *Session) executeTransition(transitionID uint16) {
	protocolVersion, ok := s.preparedTransitions[transitionID]
	if !ok {
		return
	}

	delete(s.preparedTransitions, transitionID)

	if protocolVersion == disabledProtocolVersion {
		s.session.Reset()
	}

	s.setupKeyRatchetForUser(s.selfUserID, protocolVersion)
}

func (s *Session) prepareTransition(transitionID uint16, protocolVersion uint16) {
	for userID, _ := range s.decryptors {
		s.setupKeyRatchetForUser(userID, protocolVersion)
	}

	if transitionID == initTransitionId {
		s.setupKeyRatchetForUser(s.selfUserID, protocolVersion)
	} else {
		s.preparedTransitions[transitionID] = protocolVersion
	}
}

func (s *Session) setupKeyRatchetForUser(userID godave.UserID, protocolVersion uint16) {
	disabled := protocolVersion == disabledProtocolVersion

	if userID == s.selfUserID {
		s.encryptor.SetPassthroughMode(disabled)
		if !disabled {
			s.encryptor.SetKeyRatchet(s.session.GetKeyRatchet(string(userID)))
		}
		return
	}

	decryptor := s.decryptors[userID]
	decryptor.TransitionToPassthroughMode(disabled)
	if !disabled {
		decryptor.TransitionToKeyRatchet(s.session.GetKeyRatchet(string(userID)))
	}
}

func (s *Session) sendMLSKeyPackage() {
	if err := s.callbacks.SendMLSKeyPackage(s.session.GetMarshalledKeyPackage()); err != nil {
		s.logger.Error("failed to send MLS key package", slog.Any("err", err))
	}
}

func (s *Session) sendMLSCommitWelcome(message []byte) {
	if err := s.callbacks.SendMLSCommitWelcome(message); err != nil {
		s.logger.Error("failed to send MLS commit welcome", slog.Any("err", err))
	}
}

func (s *Session) sendReadyForTransition(transitionID uint16) {
	if err := s.callbacks.SendReadyForTransition(transitionID); err != nil {
		s.logger.Error("failed to send ready for transition", slog.Any("err", err))
	}
}

func (s *Session) sendInvalidCommitWelcome(transitionID uint16) {
	if err := s.callbacks.SendInvalidCommitWelcome(transitionID); err != nil {
		s.logger.Error("failed to send invalid commit welcome", slog.Any("err", err))
	}
}
