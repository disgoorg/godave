package libdave

import (
	"runtime"
	"testing"
	"time"
)

// The finalizer must not capture the decryptor it frees: a closure that does
// keeps the object reachable, so the finalizer never runs and the C++
// decryptor is never destroyed.
func TestDecryptorIsDestroyedOnceUnreachable(t *testing.T) {
	const n = 100
	before := destroyedDecryptors.Load()
	dropDecryptors(n)

	destroyed := int64(0)
	for i := 0; i < 20 && destroyed < n/2; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
		destroyed = destroyedDecryptors.Load() - before
	}
	if destroyed < n/2 {
		t.Fatalf("%d of %d dropped decryptors were destroyed; their finalizers do not run", destroyed, n)
	}
}

// dropDecryptors creates decryptors and keeps no reference to them. It is not
// inlined so that none survives in the caller's frame.
//
//go:noinline
func dropDecryptors(n int) {
	for range n {
		NewDecryptor()
	}
}
