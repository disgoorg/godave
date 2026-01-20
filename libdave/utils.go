package libdave

// #include <stdlib.h>
// #include <stdint.h>
import "C"
import (
	"unsafe"
)

func stringSliceToC(strings []string) (**C.char, func()) {
	cArray := make([]*C.char, len(strings))
	for i, s := range strings {
		cArray[i] = C.CString(s)
	}

	freeFunc := func() {
		for _, ptr := range cArray {
			C.free(unsafe.Pointer(ptr))
		}
	}

	return &cArray[0], freeFunc
}

// IMPORTANT: This function will free the underlying C memory, so cArray becomes unsafe to use
// after this function call
func newByteSlice(cArray *C.uint8_t, length C.size_t) []byte {
	view := unsafe.Slice((*byte)(cArray), length)

	slice := make([]byte, length)
	copy(slice, view)

	C.free(unsafe.Pointer(cArray))

	return slice
}

// IMPORTANT: This function will free the underlying C memory, so cArray becomes unsafe to use
// after this function call
func newUint64Slice(cArray *C.uint64_t, length C.size_t) []uint64 {
	view := unsafe.Slice((*uint64)(cArray), length)

	slice := make([]uint64, length)
	copy(slice, view)

	C.free(unsafe.Pointer(cArray))

	return slice
}
