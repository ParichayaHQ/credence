package cid

import "errors"

var (
	// ErrEmptyData indicates attempting to generate CID from empty data
	ErrEmptyData = errors.New("cannot generate CID from empty data")

	// ErrInvalidCID indicates the CID format is invalid
	ErrInvalidCID = errors.New("invalid CID format")

	// ErrCIDNotFound indicates content not found for the given CID
	ErrCIDNotFound = errors.New("content not found for CID")

	// ErrContentTooLarge indicates the content exceeds size limits
	ErrContentTooLarge = errors.New("content too large")

	// ErrInvalidMultihash indicates the multihash is invalid
	ErrInvalidMultihash = errors.New("invalid multihash")

	// ErrUnsupportedHashFunction indicates unsupported hash function
	ErrUnsupportedHashFunction = errors.New("unsupported hash function")

	// ErrStorageFull indicates storage capacity is full
	ErrStorageFull = errors.New("storage full")

	// ErrCorruptedContent indicates stored content is corrupted
	ErrCorruptedContent = errors.New("corrupted content")

	// ErrProviderNotFound indicates no providers found for CID
	ErrProviderNotFound = errors.New("no providers found for CID")

	// ErrNetworkTimeout indicates network operation timed out
	ErrNetworkTimeout = errors.New("network timeout")
)