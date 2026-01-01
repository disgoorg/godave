package libdave

// #include <stdlib.h>
// #include <stdint.h>
import "C"
import (
	"runtime"
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

// IMPORTANT: The cArray pointer passed here should not be used after this function to prevent a use-after-free
func newCBytesMemoryView(cArray *C.uint8_t, length C.size_t) []byte {
	// A bit of a hacky solution, but this allows tracking the underlying C allocated
	// memory with the Go slice, and cleaning it all up when it falls out of scope
	if length == 0 {
		return nil
	}

	slice := unsafe.Slice((*byte)(cArray), length)

	runtime.AddCleanup(&slice, func(cArray *C.uint8_t) {
		C.free(unsafe.Pointer(cArray))
	}, cArray)

	return slice
}

// IMPORTANT: The cArray pointer passed here should not be used after this function to prevent a use-after-free
func newCUint64MemoryView(cArray *C.uint64_t, length C.size_t) []uint64 {
	// A bit of a hacky solution, but this allows tracking the underlying C allocated
	// memory with the Go slice, and cleaning it all up when it falls out of scope
	if length == 0 {
		return nil
	}

	slice := unsafe.Slice((*uint64)(cArray), length)

	runtime.AddCleanup(&slice, func(cArray *C.uint64_t) {
		C.free(unsafe.Pointer(cArray))
	}, cArray)

	return slice
}
