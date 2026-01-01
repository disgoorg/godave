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

	session := &Session{handle: C.daveSessionCreate(unsafe.Pointer(cContext), cAuthSessionID, C.DAVEMLSFailureCallback(unsafe.Pointer(C.libdaveGlobalFailureCallback)))}

	runtime.SetFinalizer(session, func(e *Session) {
		C.daveSessionDestroy(e.handle)
	})

	return session
}

func (session *Session) Init(version uint16, channelID uint64, selfUserID string) {
	cSelfUserID := C.CString(selfUserID)
	defer C.free(unsafe.Pointer(cSelfUserID))

	C.daveSessionInit(session.handle, C.uint16_t(version), C.uint64_t(channelID), cSelfUserID)
}

func (session *Session) Reset() {
	C.daveSessionReset(session.handle)
}

func (session *Session) SetProtocolVersion(version uint16) {
	C.daveSessionSetProtocolVersion(session.handle, C.uint16_t(version))
}

func (session *Session) GetProtocolVersion() uint16 {
	return uint16(C.daveSessionGetProtocolVersion(session.handle))
}

func (session *Session) GetLastEpochAuthenticator() []byte {
	var authenticator *C.uint8_t
	var authenticatorLen C.size_t

	C.daveSessionGetLastEpochAuthenticator(session.handle, &authenticator, &authenticatorLen)

	return newCBytesMemoryView(authenticator, authenticatorLen)
}

func (session *Session) SetExternalSender(externalSender []byte) {
	C.daveSessionSetExternalSender(session.handle, (*C.uint8_t)(unsafe.Pointer(&externalSender[0])), C.size_t(len(externalSender)))
}

func (session *Session) ProcessProposals(proposals []byte, recognizedUserIDs []string) []byte {
	cRecognizedUserIDs, free := stringSliceToC(recognizedUserIDs)
	defer free()

	var welcomeBytes *C.uint8_t
	var welcomeBytesLen C.size_t

	C.daveSessionProcessProposals(
		session.handle,
		(*C.uint8_t)(unsafe.Pointer(&proposals[0])),
		C.size_t(len(proposals)),
		cRecognizedUserIDs,
		C.size_t(len(recognizedUserIDs)),
		&welcomeBytes,
		&welcomeBytesLen,
	)

	return newCBytesMemoryView(welcomeBytes, welcomeBytesLen)
}

func (session *Session) ProcessCommit(commit []byte) *CommitResult {
	var res C.DAVECommitResultHandle

	res = C.daveSessionProcessCommit(session.handle, (*C.uint8_t)(unsafe.Pointer(&commit[0])), C.size_t(len(commit)))

	return newCommitResult(res)
}

func (session *Session) ProcessWelcome(welcome []byte, recognizedUserIDs []string) *WelcomeResult {
	cRecognizedUserIDs, free := stringSliceToC(recognizedUserIDs)
	defer free()

	var handle C.DAVEWelcomeResultHandle
	handle = C.daveSessionProcessWelcome(
		session.handle,
		(*C.uint8_t)(unsafe.Pointer(&welcome[0])),
		C.size_t(len(welcome)),
		cRecognizedUserIDs,
		C.size_t(len(recognizedUserIDs)),
	)

	return newWelcomeResult(handle)
}

func (session *Session) GetMarshalledKeyPackage() []byte {
	var keyPackage *C.uint8_t
	var keyPackageLen C.size_t

	C.daveSessionGetMarshalledKeyPackage(session.handle, &keyPackage, &keyPackageLen)

	return newCBytesMemoryView(keyPackage, keyPackageLen)
}

func (session *Session) GetKeyRatchet(userID string) *KeyRatchet {
	cUserID := C.CString(userID)
	defer C.free(unsafe.Pointer(cUserID))

	return newKeyRatchet(C.daveSessionGetKeyRatchet(session.handle, cUserID))
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
