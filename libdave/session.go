package libdave

// #include <stdlib.h>
// #include "dave.h"
// extern void libdaveGlobalFailureCallback(char* source, char* reason);
import "C"
import (
	"log/slog"
	"runtime"
	"unsafe"
)

type sessionHandle = C.DAVESessionHandle

type Session struct {
	handle sessionHandle
}

//export libdaveGlobalFailureCallback
func libdaveGlobalFailureCallback(source *C.char, reason *C.char) {
	defaultLogger.Load().Error(C.GoString(reason), slog.String("source", C.GoString(source)))
}

func NewSession(context string, authSessionID string) *Session {
	cContext := C.CString(context)
	defer C.free(unsafe.Pointer(cContext))

	cAuthSessionID := C.CString(authSessionID)
	defer C.free(unsafe.Pointer(cAuthSessionID))

	session := &Session{
		handle: C.daveSessionCreate(
			unsafe.Pointer(cContext),
			cAuthSessionID,
			C.DAVEMLSFailureCallback(unsafe.Pointer(C.libdaveGlobalFailureCallback)),
		),
	}

	runtime.AddCleanup(session, func(handle sessionHandle) {
		C.daveSessionDestroy(handle)
	}, session.handle)

	return session
}

func (s *Session) Init(version uint16, channelID uint64, selfUserID string) {
	cSelfUserID := C.CString(selfUserID)
	defer C.free(unsafe.Pointer(cSelfUserID))

	C.daveSessionInit(s.handle, C.uint16_t(version), C.uint64_t(channelID), cSelfUserID)
}

func (s *Session) Reset() {
	C.daveSessionReset(s.handle)
}

func (s *Session) SetProtocolVersion(version uint16) {
	C.daveSessionSetProtocolVersion(s.handle, C.uint16_t(version))
}

func (s *Session) GetProtocolVersion() uint16 {
	return uint16(C.daveSessionGetProtocolVersion(s.handle))
}

func (s *Session) GetLastEpochAuthenticator() []byte {
	var (
		authenticator    *C.uint8_t
		authenticatorLen C.size_t
	)
	C.daveSessionGetLastEpochAuthenticator(s.handle, &authenticator, &authenticatorLen)

	return newCBytesMemoryView(authenticator, authenticatorLen)
}

func (s *Session) SetExternalSender(externalSender []byte) {
	C.daveSessionSetExternalSender(s.handle, (*C.uint8_t)(unsafe.Pointer(&externalSender[0])), C.size_t(len(externalSender)))
}

func (s *Session) ProcessProposals(proposals []byte, recognizedUserIDs []string) []byte {
	cRecognizedUserIDs, free := stringSliceToC(recognizedUserIDs)
	defer free()

	var (
		welcomeBytes    *C.uint8_t
		welcomeBytesLen C.size_t
	)
	C.daveSessionProcessProposals(
		s.handle,
		(*C.uint8_t)(unsafe.Pointer(&proposals[0])),
		C.size_t(len(proposals)),
		cRecognizedUserIDs,
		C.size_t(len(recognizedUserIDs)),
		&welcomeBytes,
		&welcomeBytesLen,
	)

	return newCBytesMemoryView(welcomeBytes, welcomeBytesLen)
}

func (s *Session) ProcessCommit(commit []byte) *CommitResult {
	return newCommitResult(C.daveSessionProcessCommit(s.handle, (*C.uint8_t)(unsafe.Pointer(&commit[0])), C.size_t(len(commit))))
}

func (s *Session) ProcessWelcome(welcome []byte, recognizedUserIDs []string) *WelcomeResult {
	cRecognizedUserIDs, free := stringSliceToC(recognizedUserIDs)
	defer free()

	return newWelcomeResult(C.daveSessionProcessWelcome(
		s.handle,
		(*C.uint8_t)(unsafe.Pointer(&welcome[0])),
		C.size_t(len(welcome)),
		cRecognizedUserIDs,
		C.size_t(len(recognizedUserIDs)),
	))
}

func (s *Session) GetMarshalledKeyPackage() []byte {
	var (
		keyPackage    *C.uint8_t
		keyPackageLen C.size_t
	)
	C.daveSessionGetMarshalledKeyPackage(s.handle, &keyPackage, &keyPackageLen)

	return newCBytesMemoryView(keyPackage, keyPackageLen)
}

func (s *Session) GetKeyRatchet(userID string) *KeyRatchet {
	cUserID := C.CString(userID)
	defer C.free(unsafe.Pointer(cUserID))

	return newKeyRatchet(C.daveSessionGetKeyRatchet(s.handle, cUserID))
}

// FIXME: Implement using trampoline when https://github.com/discord/libdave/issues/10 is implemented
//        An alternative is to use a global cgo.Handle, but it will prevent concurrent calls to GetPairwiseFingerprint
// func (session *Session) GetPairwiseFingerprint(version uint16, userID string) []byte {
// 	cUserID := C.CString(userID)
// 	defer C.free(unsafe.Pointer(cUserID))
//
// 	ch := make(chan []byte)
// 	callback := func(fingerprint *C.uint8_t, length C.size_t) {
// 		ch <- newCBytesMemoryView(fingerprint, length)
// 	}
//
// 	fHandle := cgo.NewHandle(callback)
// 	defer fHandle.Delete()
// 	C.daveSessionGetPairwiseFingerprint(session.handle, C.uint16_t(version), cUserID, (C.DAVEPairwiseFingerprintCallback)(unsafe.Pointer(fHandle)))
//
// 	return <-ch
// }
