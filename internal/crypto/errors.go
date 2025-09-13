package crypto

import "errors"

var (
	// ErrNoPrivateKey indicates no private key is available for signing
	ErrNoPrivateKey = errors.New("no private key available")

	// ErrInvalidKeySize indicates the key has an invalid size
	ErrInvalidKeySize = errors.New("invalid key size")

	// ErrInvalidSignature indicates the signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidPublicKey indicates the public key is invalid
	ErrInvalidPublicKey = errors.New("invalid public key")

	// ErrInvalidPrivateKey indicates the private key is invalid
	ErrInvalidPrivateKey = errors.New("invalid private key")

	// ErrSignatureTooShort indicates the signature is too short
	ErrSignatureTooShort = errors.New("signature too short")

	// ErrSignatureTooLong indicates the signature is too long
	ErrSignatureTooLong = errors.New("signature too long")

	// ErrThresholdNotMet indicates BLS threshold signature requirements not met
	ErrThresholdNotMet = errors.New("threshold not met")

	// ErrInvalidThreshold indicates the threshold value is invalid
	ErrInvalidThreshold = errors.New("invalid threshold")

	// ErrDuplicatePartialSignature indicates a duplicate partial signature
	ErrDuplicatePartialSignature = errors.New("duplicate partial signature")

	// ErrInvalidVRFProof indicates the VRF proof is invalid
	ErrInvalidVRFProof = errors.New("invalid VRF proof")

	// ErrRandomnessGenerationFailed indicates random number generation failed
	ErrRandomnessGenerationFailed = errors.New("randomness generation failed")
)