package libdave

import "errors"

var (
	ErrGenericEncryptionFailure = errors.New("failed to encrypt frame")
	ErrGenericDecryptionFailure = errors.New("failed to decrypt frame")
	ErrMissingKeyRatchet        = errors.New("missing key ratchet")
	ErrInvalidNonce             = errors.New("invalid nonce")
	ErrMissingCryptor           = errors.New("missing cryptor")
	ErrTooManyAttempts          = errors.New("too many attempts to encrypt the frame failed")
)
