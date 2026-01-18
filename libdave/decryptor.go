package libdave

// #include "lib/include/dave.h"
import "C"
import (
	"runtime"
	"unsafe"
)

type decryptorResultCode int

const (
	decryptorResultCodeSuccess decryptorResultCode = iota
	decryptorResultCodeDecryptionFailure
	decryptorResultCodeMissingKeyRatchet
	decryptorResultCodeInvalidNonce
	decryptorResultCodeMissingCryptor
)

func (r decryptorResultCode) ToError() error {
	switch r {
	case decryptorResultCodeSuccess:
		return nil
	case decryptorResultCodeMissingKeyRatchet:
		return ErrMissingKeyRatchet
	case decryptorResultCodeInvalidNonce:
		return ErrInvalidNonce
	case decryptorResultCodeMissingCryptor:
		return ErrMissingCryptor
	default:
		return ErrGenericDecryptionFailure
	}
}

type DecryptorStats struct {
	PassthroughCount         uint64
	DecryptSuccessCount      uint64
	DecryptFailureCount      uint64
	DecryptDuration          uint64
	DecryptAttempts          uint64
	DecryptMissingKeyCount   uint64
	DecryptInvalidNonceCount uint64
}

type decryptorHandle = C.DAVEDecryptorHandle

type Decryptor struct {
	handle decryptorHandle
}

func NewDecryptor() *Decryptor {
	decryptor := &Decryptor{
		handle: C.daveDecryptorCreate(),
	}

	runtime.AddCleanup(decryptor, func(handle decryptorHandle) {
		C.daveDecryptorDestroy(handle)
	}, decryptor.handle)

	return decryptor
}

func (d *Decryptor) TransitionToKeyRatchet(keyRatchet *KeyRatchet) {
	C.daveDecryptorTransitionToKeyRatchet(d.handle, keyRatchet.handle)
}

func (d *Decryptor) TransitionToPassthroughMode(passthroughMode bool) {
	C.daveDecryptorTransitionToPassthroughMode(d.handle, C.bool(passthroughMode))
}

func (d *Decryptor) GetMaxPlaintextByteSize(mediaType MediaType, encryptedFrameSize int) int {
	return int(C.daveDecryptorGetMaxPlaintextByteSize(d.handle, C.DAVEMediaType(mediaType), C.size_t(encryptedFrameSize)))
}

func (d *Decryptor) Decrypt(mediaType MediaType, frame []byte, decryptedFrame []byte) (int, error) {
	var bytesWritten C.size_t
	res := decryptorResultCode(C.daveDecryptorDecrypt(
		d.handle,
		C.DAVEMediaType(mediaType),
		(*C.uint8_t)(unsafe.Pointer(&frame[0])),
		C.size_t(len(frame)),
		(*C.uint8_t)(unsafe.Pointer(&decryptedFrame[0])),
		C.size_t(cap(decryptedFrame)),
		&bytesWritten,
	))

	return int(bytesWritten), res.ToError()
}

func (d *Decryptor) GetStats(mediaType MediaType) *DecryptorStats {
	var cStats C.DAVEDecryptorStats
	C.daveDecryptorGetStats(d.handle, C.DAVEMediaType(mediaType), &cStats)

	return &DecryptorStats{
		PassthroughCount:         uint64(cStats.passthroughCount),
		DecryptSuccessCount:      uint64(cStats.decryptSuccessCount),
		DecryptFailureCount:      uint64(cStats.decryptFailureCount),
		DecryptDuration:          uint64(cStats.decryptDuration),
		DecryptAttempts:          uint64(cStats.decryptAttempts),
		DecryptMissingKeyCount:   uint64(cStats.decryptMissingKeyCount),
		DecryptInvalidNonceCount: uint64(cStats.decryptInvalidNonceCount),
	}
}
