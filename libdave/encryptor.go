package libdave

// #include "dave.h"
import "C"
import (
	"runtime"
	"unsafe"
)

type encryptorResultCode int

const (
	encryptorResultCodeSuccess encryptorResultCode = iota
	encryptorResultCodeEncryptionFailure
)

func (r encryptorResultCode) ToError() error {
	switch r {
	case encryptorResultCodeEncryptionFailure:
		return ErrEncryptionFailure
	default:
		return nil
	}
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
	handle encryptionHandle
}

func NewEncryptor() *Encryptor {
	encryptor := &Encryptor{handle: C.daveEncryptorCreate()}

	runtime.SetFinalizer(encryptor, func(e *Encryptor) {
		C.daveEncryptorDestroy(e.handle)
	})

	return encryptor
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
	var res C.uint16_t
	res = C.daveEncryptorGetProtocolVersion(e.handle)
	return uint16(res)
}

func (e *Encryptor) GetMaxCiphertextByteSize(mediaType MediaType, frameSize int) int {
	var res C.size_t
	res = C.daveEncryptorGetMaxCiphertextByteSize(e.handle, C.DAVEMediaType(mediaType), C.size_t(frameSize))
	return int(res)
}

func (e *Encryptor) Encrypt(mediaType MediaType, ssrc uint32, frame []byte) ([]byte, error) {
	capacity := C.daveEncryptorGetMaxCiphertextByteSize(e.handle, C.DAVEMediaType(mediaType), C.size_t(len(frame)))
	outBuf := make([]byte, capacity)

	var bytesWritten C.size_t
	res := encryptorResultCode(C.daveEncryptorEncrypt(
		e.handle,
		C.DAVEMediaType(mediaType),
		C.uint32_t(ssrc),
		(*C.uint8_t)(unsafe.Pointer(&frame[0])),
		C.size_t(len(frame)),
		(*C.uint8_t)(unsafe.Pointer(&outBuf[0])),
		capacity,
		&bytesWritten,
	))

	if res != encryptorResultCodeSuccess {
		return nil, res.ToError()
	}

	return outBuf[:bytesWritten], nil
}

func (e *Encryptor) SetProtocolVersionChangedCallback() {
	panic("TODO")
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
