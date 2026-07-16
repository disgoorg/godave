package libdave

// #include "dave.h"
import "C"
import "runtime"

type keyRatchetHandle = C.DAVEKeyRatchetHandle

type KeyRatchet struct {
	handle keyRatchetHandle
}

func newKeyRatchet(handle keyRatchetHandle) *KeyRatchet {
	if handle == nil {
		return nil
	}

	keyRatchet := &KeyRatchet{handle: handle}

	runtime.SetFinalizer(keyRatchet, func(k *KeyRatchet) {
		C.daveKeyRatchetDestroy(k.handle)
	})

	return keyRatchet
}
