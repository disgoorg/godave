package libdave

import (
	"runtime"
	"testing"
	"weak"
)

// The finalizer must not capture the decryptor it frees: a closure that does
// keeps the object reachable, so it is never collected and the C++ decryptor
// is never destroyed. With that bug none are collected; the test tolerates a
// few survivors because a platform may briefly keep the latest allocation
// reachable.
func TestDecryptorIsFreedOnceUnreachable(t *testing.T) {
	const n = 100
	refs := newUnreachableDecryptors(n)

	alive := n
	for i := 0; i < 10 && alive > n/2; i++ {
		runtime.GC()
		alive = 0
		for _, ref := range refs {
			if ref.Value() != nil {
				alive++
			}
		}
	}
	if alive > n/2 {
		t.Fatalf("%d of %d decryptors are still reachable after GC; their finalizers will never run", alive, n)
	}
}

// newUnreachableDecryptors creates decryptors and keeps only weak references
// to them. It is not inlined so that no strong reference survives in the
// caller's frame.
//
//go:noinline
func newUnreachableDecryptors(n int) []weak.Pointer[Decryptor] {
	refs := make([]weak.Pointer[Decryptor], n)
	for i := range refs {
		refs[i] = weak.Make(NewDecryptor())
	}
	return refs
}
