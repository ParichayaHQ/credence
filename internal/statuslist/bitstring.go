package statuslist

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

// NewBitString creates a new BitString with the specified initial size
func NewBitString(size int) *BitString {
	if size <= 0 {
		size = 1
	}
	
	// Calculate number of bytes needed (round up to nearest byte)
	numBytes := (size + 7) / 8
	
	return &BitString{
		bits:   make([]byte, numBytes),
		length: size,
	}
}

// FromCompressedBase64 creates a BitString from compressed base64 encoded data
func FromCompressedBase64(encoded string) (*BitString, error) {
	if encoded == "" {
		return &BitString{bits: []byte{}, length: 0}, nil
	}
	
	// Decode base64
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorEncodingError, "failed to decode base64", err.Error())
	}
	
	// Decompress with gzip
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorCompressionError, "failed to create gzip reader", err.Error())
	}
	defer reader.Close()
	
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, NewStatusListErrorWithDetails(ErrorCompressionError, "failed to decompress data", err.Error())
	}
	
	// Create BitString from decompressed data
	// Note: We use the byte length * 8 as the bit length since we don't store
	// the original bit length in the compressed format. This means some trailing
	// bits might be considered part of the bitstring even if they weren't in the original.
	return &BitString{
		bits:   decompressed,
		length: len(decompressed) * 8,
	}, nil
}

// Set sets the bit at the specified index to the given value
func (bs *BitString) Set(index int, value bool) error {
	if index < 0 {
		return NewStatusListError(ErrorInvalidIndex, "index cannot be negative")
	}
	
	// Expand if necessary
	if index >= bs.length {
		if err := bs.Expand(index + 1); err != nil {
			return err
		}
	}
	
	byteIndex := index / 8
	bitIndex := uint(index % 8)
	
	if value {
		// Set bit to 1
		bs.bits[byteIndex] |= 1 << bitIndex
	} else {
		// Set bit to 0
		bs.bits[byteIndex] &^= 1 << bitIndex
	}
	
	return nil
}

// Get returns the value of the bit at the specified index
func (bs *BitString) Get(index int) (bool, error) {
	if index < 0 {
		return false, NewStatusListError(ErrorInvalidIndex, "index cannot be negative")
	}
	
	if index >= bs.length {
		// Out of bounds reads return false (unset)
		return false, nil
	}
	
	byteIndex := index / 8
	bitIndex := uint(index % 8)
	
	return (bs.bits[byteIndex] & (1 << bitIndex)) != 0, nil
}

// Length returns the current length of the BitString
func (bs *BitString) Length() int {
	return bs.length
}

// Size returns the number of bytes used by the BitString
func (bs *BitString) Size() int {
	return len(bs.bits)
}

// Expand expands the BitString to accommodate at least the specified length
func (bs *BitString) Expand(newLength int) error {
	if newLength <= bs.length {
		return nil // Already large enough
	}
	
	// Calculate new byte array size
	newNumBytes := (newLength + 7) / 8
	
	if newNumBytes > len(bs.bits) {
		// Create new larger byte array
		newBits := make([]byte, newNumBytes)
		copy(newBits, bs.bits)
		bs.bits = newBits
	}
	
	bs.length = newLength
	return nil
}

// ToCompressedBase64 encodes the BitString as compressed base64
func (bs *BitString) ToCompressedBase64(compressionLevel int) (string, error) {
	if len(bs.bits) == 0 || bs.length == 0 {
		return "", nil
	}
	
	// Compress with gzip
	var compressed bytes.Buffer
	writer, err := gzip.NewWriterLevel(&compressed, compressionLevel)
	if err != nil {
		return "", NewStatusListErrorWithDetails(ErrorCompressionError, "failed to create gzip writer", err.Error())
	}
	
	if _, err := writer.Write(bs.bits); err != nil {
		writer.Close()
		return "", NewStatusListErrorWithDetails(ErrorCompressionError, "failed to compress data", err.Error())
	}
	
	if err := writer.Close(); err != nil {
		return "", NewStatusListErrorWithDetails(ErrorCompressionError, "failed to close gzip writer", err.Error())
	}
	
	// Encode to base64
	return base64.StdEncoding.EncodeToString(compressed.Bytes()), nil
}

// Clone creates a deep copy of the BitString
func (bs *BitString) Clone() *BitString {
	newBits := make([]byte, len(bs.bits))
	copy(newBits, bs.bits)
	
	return &BitString{
		bits:   newBits,
		length: bs.length,
	}
}

// SetRange sets multiple consecutive bits starting at the given index
func (bs *BitString) SetRange(startIndex int, values []bool) error {
	for i, value := range values {
		if err := bs.Set(startIndex+i, value); err != nil {
			return err
		}
	}
	return nil
}

// GetRange gets multiple consecutive bits starting at the given index
func (bs *BitString) GetRange(startIndex, count int) ([]bool, error) {
	if startIndex < 0 {
		return nil, NewStatusListError(ErrorInvalidIndex, "start index cannot be negative")
	}
	
	if count <= 0 {
		return []bool{}, nil
	}
	
	values := make([]bool, count)
	for i := 0; i < count; i++ {
		value, err := bs.Get(startIndex + i)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	
	return values, nil
}

// CountSet returns the number of bits set to true
func (bs *BitString) CountSet() int {
	count := 0
	for i := 0; i < bs.length; i++ {
		if value, _ := bs.Get(i); value {
			count++
		}
	}
	return count
}

// CountUnset returns the number of bits set to false
func (bs *BitString) CountUnset() int {
	return bs.length - bs.CountSet()
}

// FindFirstUnset finds the index of the first unset bit (false value)
func (bs *BitString) FindFirstUnset() int {
	for i := 0; i < bs.length; i++ {
		if value, _ := bs.Get(i); !value {
			return i
		}
	}
	return -1 // All bits are set
}

// FindFirstSet finds the index of the first set bit (true value)
func (bs *BitString) FindFirstSet() int {
	for i := 0; i < bs.length; i++ {
		if value, _ := bs.Get(i); value {
			return i
		}
	}
	return -1 // No bits are set
}

// Clear resets all bits to false
func (bs *BitString) Clear() {
	for i := range bs.bits {
		bs.bits[i] = 0
	}
}

// ClearRange resets a range of bits to false
func (bs *BitString) ClearRange(startIndex, count int) error {
	for i := 0; i < count; i++ {
		if err := bs.Set(startIndex+i, false); err != nil {
			return err
		}
	}
	return nil
}

// SetAll sets all bits to true
func (bs *BitString) SetAll() {
	// Set all complete bytes to 0xFF
	for i := 0; i < len(bs.bits); i++ {
		bs.bits[i] = 0xFF
	}
	
	// Handle the last partial byte if necessary
	remainingBits := bs.length % 8
	if remainingBits > 0 {
		lastByteIndex := len(bs.bits) - 1
		mask := byte((1 << remainingBits) - 1)
		bs.bits[lastByteIndex] = mask
	}
}

// String returns a string representation of the BitString for debugging
func (bs *BitString) String() string {
	result := fmt.Sprintf("BitString{length: %d, size: %d bytes, bits: ", bs.length, len(bs.bits))
	
	maxDisplay := 64 // Limit display for readability
	displayLength := bs.length
	if displayLength > maxDisplay {
		displayLength = maxDisplay
	}
	
	for i := 0; i < displayLength; i++ {
		if value, _ := bs.Get(i); value {
			result += "1"
		} else {
			result += "0"
		}
		
		if i > 0 && (i+1)%8 == 0 {
			result += " "
		}
	}
	
	if bs.length > maxDisplay {
		result += "..."
	}
	
	result += "}"
	return result
}

// Bytes returns a copy of the underlying byte array
func (bs *BitString) Bytes() []byte {
	result := make([]byte, len(bs.bits))
	copy(result, bs.bits)
	return result
}

// FromBytes creates a BitString from a byte array
func FromBytes(data []byte, length int) *BitString {
	bits := make([]byte, len(data))
	copy(bits, data)
	
	if length <= 0 {
		length = len(data) * 8
	}
	
	return &BitString{
		bits:   bits,
		length: length,
	}
}

// Equals compares two BitStrings for equality
func (bs *BitString) Equals(other *BitString) bool {
	if other == nil {
		return false
	}
	
	if bs.length != other.length {
		return false
	}
	
	// Compare only the used portion of the byte arrays
	usedBytes := (bs.length + 7) / 8
	
	for i := 0; i < usedBytes; i++ {
		if i < len(bs.bits) && i < len(other.bits) {
			// For the last byte, only compare the used bits
			if i == usedBytes-1 {
				remainingBits := bs.length % 8
				if remainingBits > 0 {
					mask := byte((1 << remainingBits) - 1)
					if (bs.bits[i] & mask) != (other.bits[i] & mask) {
						return false
					}
				} else {
					if bs.bits[i] != other.bits[i] {
						return false
					}
				}
			} else {
				if bs.bits[i] != other.bits[i] {
					return false
				}
			}
		} else {
			// One has more bytes than the other
			return false
		}
	}
	
	return true
}