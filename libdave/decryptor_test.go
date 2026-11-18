package libdave

import (
	"runtime"
	"testing"
	"weak"
)

// The finalizer must not capture the decryptor it frees: a closure that does
// keeps the object reachable, so it is never collected and the C++ decryptor
// is never destroyed.
func TestDecryptorIsFreedOnceUnreachable(t *testing.T) {
	ref := weak.Make(NewDecryptor())
	for i := 0; i < 5 && ref.Value() != nil; i++ {
		runtime.GC()
	}
	if ref.Value() != nil {
		t.Fatal("decryptor is still reachable after GC; its finalizer will never run")
	}
}
