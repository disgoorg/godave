package libdave

// #include "dave.h"
// extern void godaveProtocolVersionChangedCallback(void* userData);
import "C"
import (
	"log/slog"
	"runtime"
	"runtime/cgo"
	"unsafe"
	"weak"
)

type encryptorResultCode int

const (
	encryptorResultCodeSuccess encryptorResultCode = iota
	encryptorResultCodeEncryptionFailure
	encryptorResultCodeMissingKeyRatchet
	encryptionResultCodeMissingCryptor
	encryptionResultCodeTooManyAttempts
)

func (r encryptorResultCode) ToError() error {
	switch r {
	case encryptorResultCodeSuccess:
		return nil
	case encryptorResultCodeMissingKeyRatchet:
		return ErrMissingKeyRatchet
	case encryptionResultCodeMissingCryptor:
		return ErrMissingCryptor
	case encryptionResultCodeTooManyAttempts:
		return ErrTooManyAttempts
	default:
		return ErrGenericEncryptionFailure
	}
}

//export godaveProtocolVersionChangedCallback
func godaveProtocolVersionChangedCallback(userData unsafe.Pointer) {
	h := *(*cgo.Handle)(userData)
	encryptor := h.Value().(weak.Pointer[Encryptor]).Value()

	if encryptor == nil {
		return
	}

	defaultLogger.Load().Debug("protocol version changed", slog.Int("newVersion", int(encryptor.GetProtocolVersion())))
}

type EncryptorStats struct {
	PassthroughCount       uint64
	EncryptSuccessCount    uint64
	EncryptFailureCount    uint64
	EncryptDuration        uint64
	EncryptAttempts        uint64
	EncryptMaxAttempts     uint64
	EncryptMissingKeyCount uint64
}

type encryptionHandle = C.DAVEEncryptorHandle

type Encryptor struct {
	handle    encryptionHandle
	pinner    runtime.Pinner
	cgoHandle cgo.Handle
}

func NewEncryptor() *Encryptor {
	encryptor := &Encryptor{
		handle: C.daveEncryptorCreate(),
	}

	// A weak pointer is necessary here to avoid circular refs
	encryptor.cgoHandle = cgo.NewHandle(weak.Make(encryptor))

	C.daveEncryptorSetProtocolVersionChangedCallback(
		encryptor.handle,
		C.DAVEEncryptorProtocolVersionChangedCallback(C.godaveProtocolVersionChangedCallback),
		unsafe.Pointer(&encryptor.cgoHandle),
	)

	runtime.SetFinalizer(encryptor, func(e *Encryptor) {
		C.daveEncryptorDestroy(e.handle)
		e.cgoHandle.Delete()
	})

	return encryptor
}

func (e *Encryptor) HasKeyRatchet() bool {
	return bool(C.daveEncryptorHasKeyRatchet(e.handle))
}

func (e *Encryptor) IsPassthroughMode() bool {
	return bool(C.daveEncryptorIsPassthroughMode(e.handle))
}

func (e *Encryptor) SetKeyRatchet(keyRatchet *KeyRatchet) {
	C.daveEncryptorSetKeyRatchet(e.handle, keyRatchet.handle)
}

func (e *Encryptor) SetPassthroughMode(passthroughMode bool) {
	C.daveEncryptorSetPassthroughMode(e.handle, C.bool(passthroughMode))
}

func (e *Encryptor) AssignSsrcToCodec(ssrc uint32, codec Codec) {
	C.daveEncryptorAssignSsrcToCodec(e.handle, C.uint32_t(ssrc), C.DAVECodec(codec))
}

func (e *Encryptor) GetProtocolVersion() uint16 {
	return uint16(C.daveEncryptorGetProtocolVersion(e.handle))
}

func (e *Encryptor) GetMaxCiphertextByteSize(mediaType MediaType, frameSize int) int {
	return int(C.daveEncryptorGetMaxCiphertextByteSize(e.handle, C.DAVEMediaType(mediaType), C.size_t(frameSize)))
}

func (e *Encryptor) Encrypt(mediaType MediaType, ssrc uint32, frame []byte, encryptedFrame []byte) (int, error) {
	var bytesWritten C.size_t
	res := encryptorResultCode(C.daveEncryptorEncrypt(
		e.handle,
		C.DAVEMediaType(mediaType),
		C.uint32_t(ssrc),
		(*C.uint8_t)(unsafe.Pointer(&frame[0])),
		C.size_t(len(frame)),
		(*C.uint8_t)(unsafe.Pointer(&encryptedFrame[0])),
		C.size_t(cap(encryptedFrame)),
		&bytesWritten,
	))

	return int(bytesWritten), res.ToError()
}

func (e *Encryptor) GetStats(mediaType MediaType) *EncryptorStats {
	var cStats C.DAVEEncryptorStats
	C.daveEncryptorGetStats(e.handle, C.DAVEMediaType(mediaType), &cStats)

	return &EncryptorStats{
		PassthroughCount:       uint64(cStats.passthroughCount),
		EncryptSuccessCount:    uint64(cStats.encryptSuccessCount),
		EncryptFailureCount:    uint64(cStats.encryptFailureCount),
		EncryptDuration:        uint64(cStats.encryptDuration),
		EncryptAttempts:        uint64(cStats.encryptAttempts),
		EncryptMaxAttempts:     uint64(cStats.encryptMaxAttempts),
		EncryptMissingKeyCount: uint64(cStats.encryptMissingKeyCount),
	}
}
