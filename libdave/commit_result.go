package libdave

// #include "lib/include/dave.h"
import "C"
import "runtime"

type commitResultHandle = C.DAVECommitResultHandle

type CommitResult struct {
	handle commitResultHandle
}

func newCommitResult(handle commitResultHandle) *CommitResult {
	commitResult := &CommitResult{
		handle: handle,
	}

	runtime.SetFinalizer(commitResult, func(c *CommitResult) {
		C.daveCommitResultDestroy(c.handle)
	})

	return commitResult
}

func (r *CommitResult) IsFailed() bool {
	return bool(C.daveCommitResultIsFailed(r.handle))
}

func (r *CommitResult) IsIgnored() bool {
	return bool(C.daveCommitResultIsIgnored(r.handle))
}

func (r *CommitResult) GetRosterMemberIDs() []uint64 {
	var (
		rosterIDs       *C.uint64_t
		rosterIDsLength C.size_t
	)
	C.daveCommitResultGetRosterMemberIds(r.handle, &rosterIDs, &rosterIDsLength)

	return newUint64Slice(rosterIDs, rosterIDsLength)
}

func (r *CommitResult) GetRosterMemberSignature(rosterID uint64) []byte {
	var (
		rosterMemberSignature       *C.uint8_t
		rosterMemberSignatureLength C.size_t
	)
	C.daveCommitResultGetRosterMemberSignature(r.handle, C.uint64_t(rosterID), &rosterMemberSignature, &rosterMemberSignatureLength)

	return newByteSlice(rosterMemberSignature, rosterMemberSignatureLength)
}
