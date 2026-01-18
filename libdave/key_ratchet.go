package libdave

// #include "lib/include/dave.h"
import "C"
import "runtime"

type keyRatchetHandle = C.DAVEKeyRatchetHandle

type KeyRatchet struct {
	handle keyRatchetHandle
}

func newKeyRatchet(handle keyRatchetHandle) *KeyRatchet {
	keyRatchet := &KeyRatchet{handle: handle}

	runtime.AddCleanup(keyRatchet, func(handle keyRatchetHandle) {
		C.daveKeyRatchetDestroy(handle)
	}, keyRatchet.handle)

	return keyRatchet
}
