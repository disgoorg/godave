package libdave

import "errors"

var (
	ErrEncryptionFailure = errors.New("failed to encrypt frame")
	ErrDecryptionFailure = errors.New("failed to decrypt frame")
	ErrMissingKeyRatchet = errors.New("missing key ratchet")
	ErrInvalidNonce      = errors.New("invalid nonce")
	ErrMissingCryptor    = errors.New("missing cryptor")
)
