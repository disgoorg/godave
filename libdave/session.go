package libdave

// #include <stdlib.h>
// #include "lib/include/dave.h"
// extern void godaveGlobalFailureCallback(char* source, char* reason, void* userData);
// extern void godavePairwiseFingerprintCallback(uint8_t* fingerpint, size_t length, void* userData);
import "C"
import (
	"log/slog"
	"runtime"
	"runtime/cgo"
	"unsafe"
)

type sessionHandle = C.DAVESessionHandle

type Session struct {
	handle sessionHandle
}

//export godaveGlobalFailureCallback
func godaveGlobalFailureCallback(source *C.char, reason *C.char, userData unsafe.Pointer) {
	defaultLogger.Load().Error(
		C.GoString(reason),
		slog.String("source", C.GoString(source)),
		slog.String("authSessionID", C.GoString((*C.char)(userData))),
	)
}

//export godavePairwiseFingerprintCallback
func godavePairwiseFingerprintCallback(fingerprint *C.uint8_t, length C.size_t, userData unsafe.Pointer) {
	h := *(*cgo.Handle)(userData)
	retChan := h.Value().(chan []byte)

	// Copy the data over into Go land
	// No need to free the C array, as the library will do it for us
	view := unsafe.Slice((*byte)(fingerprint), length)
	slice := make([]byte, length)
	copy(slice, view)

	retChan <- slice
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
			C.DAVEMLSFailureCallback(unsafe.Pointer(C.godaveGlobalFailureCallback)),
			unsafe.Pointer(cAuthSessionID),
		),
	}

	runtime.SetFinalizer(session, func(s *Session) {
		C.daveSessionDestroy(s.handle)
	})

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

	return newByteSlice(authenticator, authenticatorLen)
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

	return newByteSlice(welcomeBytes, welcomeBytesLen)
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

	return newByteSlice(keyPackage, keyPackageLen)
}

func (s *Session) GetKeyRatchet(userID string) *KeyRatchet {
	cUserID := C.CString(userID)
	defer C.free(unsafe.Pointer(cUserID))

	return newKeyRatchet(C.daveSessionGetKeyRatchet(s.handle, cUserID))
}

func (s *Session) GetPairwiseFingerprint(version uint16, userID string) []byte {
	cUserID := C.CString(userID)
	defer C.free(unsafe.Pointer(cUserID))

	ch := make(chan []byte)
	handler := cgo.NewHandle(ch)
	defer handler.Delete()

	C.daveSessionGetPairwiseFingerprint(
		s.handle,
		C.uint16_t(version),
		cUserID,
		(C.DAVEPairwiseFingerprintCallback)(unsafe.Pointer(C.godavePairwiseFingerprintCallback)),
		unsafe.Pointer(&handler),
	)

	return <-ch
}
