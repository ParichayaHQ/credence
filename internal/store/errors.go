package store

import (
	"errors"
	"fmt"
)

// Common storage errors
var (
	ErrNotFound     = errors.New("not found")
	ErrExists       = errors.New("already exists")
	ErrTooLarge     = errors.New("data too large")
	ErrInvalidCID   = errors.New("invalid CID")
	ErrInvalidEpoch = errors.New("invalid epoch")
	ErrClosed       = errors.New("store is closed")
	ErrReadOnly     = errors.New("store is read-only")
)

// StoreError wraps errors with context
type StoreError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
	
	// Context
	CID   string
	DID   string
	Epoch string
	Key   string
}

func (e *StoreError) Error() string {
	if e.CID != "" {
		return fmt.Sprintf("store %s: %v (cid: %s)", e.Op, e.Err, e.CID)
	}
	if e.DID != "" {
		return fmt.Sprintf("store %s: %v (did: %s)", e.Op, e.Err, e.DID)
	}
	if e.Epoch != "" {
		return fmt.Sprintf("store %s: %v (epoch: %s)", e.Op, e.Err, e.Epoch)
	}
	if e.Key != "" {
		return fmt.Sprintf("store %s: %v (key: %s)", e.Op, e.Err, e.Key)
	}
	return fmt.Sprintf("store %s: %v", e.Op, e.Err)
}

func (e *StoreError) Unwrap() error {
	return e.Err
}

// Convenience constructors for common error patterns

func ErrNotFoundCID(cid string) error {
	return &StoreError{
		Op:  "get",
		Err: ErrNotFound,
		CID: cid,
	}
}

func ErrNotFoundDID(did string) error {
	return &StoreError{
		Op:  "query",
		Err: ErrNotFound,
		DID: did,
	}
}

func ErrNotFoundEpoch(epoch string) error {
	return &StoreError{
		Op:    "get",
		Err:   ErrNotFound,
		Epoch: epoch,
	}
}

func ErrInvalidConfig(msg string) error {
	return &StoreError{
		Op:  "config",
		Err: errors.New(msg),
	}
}

func ErrDatabase(op string, err error) error {
	return &StoreError{
		Op:  op,
		Err: err,
	}
}

func ErrDatabaseKey(op, key string, err error) error {
	return &StoreError{
		Op:  op,
		Err: err,
		Key: key,
	}
}

// IsNotFound checks if error is a "not found" error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return errors.Is(storeErr.Err, ErrNotFound)
	}
	
	return errors.Is(err, ErrNotFound)
}

// IsExists checks if error is an "already exists" error
func IsExists(err error) bool {
	if err == nil {
		return false
	}
	
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return errors.Is(storeErr.Err, ErrExists)
	}
	
	return errors.Is(err, ErrExists)
}

// IsTooLarge checks if error is a "too large" error
func IsTooLarge(err error) bool {
	if err == nil {
		return false
	}
	
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return errors.Is(storeErr.Err, ErrTooLarge)
	}
	
	return errors.Is(err, ErrTooLarge)
}