package libdave

// #include "dave.h"
import "C"
import "runtime"

type welcomeResultHandle = C.DAVEWelcomeResultHandle

type WelcomeResult struct {
	handle welcomeResultHandle
}

func newWelcomeResult(handle welcomeResultHandle) *WelcomeResult {
	if handle == nil {
		return nil
	}

	welcomeResult := &WelcomeResult{handle: handle}

	runtime.SetFinalizer(welcomeResult, func(welcomeResult *WelcomeResult) {
		C.daveWelcomeResultDestroy(welcomeResult.handle)
	})

	return welcomeResult
}

func (w *WelcomeResult) GetRosterMemberIDs() []uint64 {
	var rosterIDs *C.uint64_t
	var rosterIDsLength C.size_t

	C.daveWelcomeResultGetRosterMemberIds(w.handle, &rosterIDs, &rosterIDsLength)

	return newCUint64MemoryView(rosterIDs, rosterIDsLength)
}

func (w *WelcomeResult) GetRosterMemberSignature(rosterID uint64) []byte {
	var rosterMemberSignature *C.uint8_t
	var rosterMemberSignatureLength C.size_t

	C.daveWelcomeResultGetRosterMemberSignature(w.handle, C.uint64_t(rosterID), &rosterMemberSignature, &rosterMemberSignatureLength)

	return newCBytesMemoryView(rosterMemberSignature, rosterMemberSignatureLength)
}
