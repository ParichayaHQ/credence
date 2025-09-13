package statuslist

import (
	"testing"
)

func TestNewBitString(t *testing.T) {
	// Test valid size
	bs := NewBitString(100)
	if bs.Length() != 100 {
		t.Errorf("expected length 100, got %d", bs.Length())
	}
	
	// Test zero/negative size
	bs = NewBitString(0)
	if bs.Length() != 1 {
		t.Errorf("expected length 1 for zero size, got %d", bs.Length())
	}
	
	bs = NewBitString(-5)
	if bs.Length() != 1 {
		t.Errorf("expected length 1 for negative size, got %d", bs.Length())
	}
}

func TestBitString_SetGet(t *testing.T) {
	bs := NewBitString(64)
	
	// Test setting and getting bits
	testCases := []struct {
		index int
		value bool
	}{
		{0, true},
		{1, false},
		{7, true},
		{8, false},
		{15, true},
		{63, true},
	}
	
	for _, tc := range testCases {
		err := bs.Set(tc.index, tc.value)
		if err != nil {
			t.Fatalf("failed to set bit at index %d: %v", tc.index, err)
		}
		
		value, err := bs.Get(tc.index)
		if err != nil {
			t.Fatalf("failed to get bit at index %d: %v", tc.index, err)
		}
		
		if value != tc.value {
			t.Errorf("expected bit at index %d to be %v, got %v", tc.index, tc.value, value)
		}
	}
	
	// Test negative index
	err := bs.Set(-1, true)
	if err == nil {
		t.Error("expected error for negative index")
	}
	
	_, err = bs.Get(-1)
	if err == nil {
		t.Error("expected error for negative index")
	}
}

func TestBitString_AutoExpand(t *testing.T) {
	bs := NewBitString(8)
	
	// Set a bit beyond current length
	err := bs.Set(15, true)
	if err != nil {
		t.Fatalf("failed to set bit with auto-expand: %v", err)
	}
	
	if bs.Length() != 16 {
		t.Errorf("expected length 16 after auto-expand, got %d", bs.Length())
	}
	
	// Verify the bit was set
	value, err := bs.Get(15)
	if err != nil {
		t.Fatalf("failed to get expanded bit: %v", err)
	}
	
	if !value {
		t.Error("expected expanded bit to be true")
	}
}

func TestBitString_GetOutOfBounds(t *testing.T) {
	bs := NewBitString(8)
	
	// Getting out of bounds should return false, no error
	value, err := bs.Get(10)
	if err != nil {
		t.Fatalf("unexpected error for out of bounds get: %v", err)
	}
	
	if value {
		t.Error("expected out of bounds bit to be false")
	}
}

func TestBitString_SetRange(t *testing.T) {
	bs := NewBitString(16)
	
	values := []bool{true, false, true, true, false}
	err := bs.SetRange(5, values)
	if err != nil {
		t.Fatalf("failed to set range: %v", err)
	}
	
	// Verify the range was set correctly
	for i, expected := range values {
		value, err := bs.Get(5 + i)
		if err != nil {
			t.Fatalf("failed to get bit at index %d: %v", 5+i, err)
		}
		
		if value != expected {
			t.Errorf("expected bit at index %d to be %v, got %v", 5+i, expected, value)
		}
	}
}

func TestBitString_GetRange(t *testing.T) {
	bs := NewBitString(16)
	
	// Set some test values
	testValues := []bool{true, false, true, true, false}
	for i, value := range testValues {
		bs.Set(5+i, value)
	}
	
	// Get the range
	values, err := bs.GetRange(5, 5)
	if err != nil {
		t.Fatalf("failed to get range: %v", err)
	}
	
	if len(values) != 5 {
		t.Errorf("expected 5 values, got %d", len(values))
	}
	
	for i, expected := range testValues {
		if values[i] != expected {
			t.Errorf("expected value at index %d to be %v, got %v", i, expected, values[i])
		}
	}
}

func TestBitString_CountSetUnset(t *testing.T) {
	bs := NewBitString(10)
	
	// Set some bits
	bs.Set(0, true)
	bs.Set(2, true)
	bs.Set(4, true)
	bs.Set(9, true)
	
	setCount := bs.CountSet()
	if setCount != 4 {
		t.Errorf("expected 4 set bits, got %d", setCount)
	}
	
	unsetCount := bs.CountUnset()
	if unsetCount != 6 {
		t.Errorf("expected 6 unset bits, got %d", unsetCount)
	}
	
	if setCount+unsetCount != bs.Length() {
		t.Error("set + unset should equal total length")
	}
}

func TestBitString_FindFirstUnsetSet(t *testing.T) {
	bs := NewBitString(10)
	
	// Initially all bits are unset
	firstUnset := bs.FindFirstUnset()
	if firstUnset != 0 {
		t.Errorf("expected first unset at index 0, got %d", firstUnset)
	}
	
	firstSet := bs.FindFirstSet()
	if firstSet != -1 {
		t.Errorf("expected no set bits, got first set at %d", firstSet)
	}
	
	// Set first bit
	bs.Set(0, true)
	
	firstUnset = bs.FindFirstUnset()
	if firstUnset != 1 {
		t.Errorf("expected first unset at index 1, got %d", firstUnset)
	}
	
	firstSet = bs.FindFirstSet()
	if firstSet != 0 {
		t.Errorf("expected first set at index 0, got %d", firstSet)
	}
}

func TestBitString_Clear(t *testing.T) {
	bs := NewBitString(10)
	
	// Set all bits
	bs.SetAll()
	
	if bs.CountSet() != 10 {
		t.Error("expected all bits to be set")
	}
	
	// Clear all bits
	bs.Clear()
	
	if bs.CountSet() != 0 {
		t.Error("expected no bits to be set after clear")
	}
}

func TestBitString_SetAll(t *testing.T) {
	bs := NewBitString(17) // Non-byte-aligned size
	
	bs.SetAll()
	
	if bs.CountSet() != 17 {
		t.Errorf("expected 17 set bits, got %d", bs.CountSet())
	}
	
	// Check that extra bits in the last byte are not set
	if bs.Size() > 0 {
		lastByte := bs.bits[len(bs.bits)-1]
		expectedMask := byte((1 << (17 % 8)) - 1) // Should be 0x01 (only first bit set)
		if lastByte != expectedMask {
			t.Errorf("expected last byte to be %02x, got %02x", expectedMask, lastByte)
		}
	}
}

func TestBitString_ToFromCompressedBase64(t *testing.T) {
	bs := NewBitString(96) // Use a byte-aligned size to avoid padding issues
	
	// Set some test pattern
	for i := 0; i < 96; i += 3 {
		bs.Set(i, true)
	}
	
	// Convert to compressed base64
	encoded, err := bs.ToCompressedBase64(6)
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	
	if encoded == "" {
		t.Error("expected non-empty encoded string")
	}
	
	// Convert back
	decoded, err := FromCompressedBase64(encoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	
	// Compare the meaningful bits (up to original length)
	for i := 0; i < bs.Length(); i++ {
		originalBit, _ := bs.Get(i)
		decodedBit, _ := decoded.Get(i)
		if originalBit != decodedBit {
			t.Errorf("bit mismatch at index %d: original=%v, decoded=%v", i, originalBit, decodedBit)
		}
	}
}

func TestBitString_Clone(t *testing.T) {
	bs := NewBitString(32)
	
	// Set some bits
	bs.Set(5, true)
	bs.Set(15, true)
	bs.Set(25, true)
	
	// Clone
	clone := bs.Clone()
	
	// Verify clone is equal
	if !bs.Equals(clone) {
		t.Error("clone should equal original")
	}
	
	// Modify original
	bs.Set(10, true)
	
	// Verify clone is unchanged
	value, _ := clone.Get(10)
	if value {
		t.Error("clone should not be affected by changes to original")
	}
}

func TestBitString_Equals(t *testing.T) {
	bs1 := NewBitString(16)
	bs2 := NewBitString(16)
	
	// Initially should be equal
	if !bs1.Equals(bs2) {
		t.Error("empty bitstrings should be equal")
	}
	
	// Set same bits
	bs1.Set(5, true)
	bs2.Set(5, true)
	
	if !bs1.Equals(bs2) {
		t.Error("bitstrings with same bits should be equal")
	}
	
	// Set different bits
	bs1.Set(10, true)
	
	if bs1.Equals(bs2) {
		t.Error("bitstrings with different bits should not be equal")
	}
	
	// Test different lengths
	bs3 := NewBitString(20)
	if bs1.Equals(bs3) {
		t.Error("bitstrings with different lengths should not be equal")
	}
	
	// Test nil comparison
	if bs1.Equals(nil) {
		t.Error("bitstring should not equal nil")
	}
}

func TestFromCompressedBase64_Empty(t *testing.T) {
	bs, err := FromCompressedBase64("")
	if err != nil {
		t.Fatalf("failed to decode empty string: %v", err)
	}
	
	if bs.Length() != 0 {
		t.Errorf("expected empty bitstring to have length 0, got %d", bs.Length())
	}
}

func TestFromBytes(t *testing.T) {
	data := []byte{0xFF, 0x00, 0xAA}
	length := 24
	
	bs := FromBytes(data, length)
	
	if bs.Length() != length {
		t.Errorf("expected length %d, got %d", length, bs.Length())
	}
	
	// Check some expected bit patterns
	// First byte (0xFF) - all bits set
	for i := 0; i < 8; i++ {
		if value, _ := bs.Get(i); !value {
			t.Errorf("expected bit %d to be set", i)
		}
	}
	
	// Second byte (0x00) - no bits set
	for i := 8; i < 16; i++ {
		if value, _ := bs.Get(i); value {
			t.Errorf("expected bit %d to be unset", i)
		}
	}
	
	// Third byte (0xAA = 10101010) - alternating pattern
	expectedPattern := []bool{false, true, false, true, false, true, false, true}
	for i, expected := range expectedPattern {
		if value, _ := bs.Get(16 + i); value != expected {
			t.Errorf("expected bit %d to be %v, got %v", 16+i, expected, value)
		}
	}
}

func TestBitString_String(t *testing.T) {
	bs := NewBitString(16)
	
	// Set some bits in a recognizable pattern
	bs.Set(0, true)
	bs.Set(2, true)
	bs.Set(8, true)
	bs.Set(15, true)
	
	str := bs.String()
	
	// Should contain basic info
	if str == "" {
		t.Error("string representation should not be empty")
	}
	
	// Should contain length and size info
	// (We don't test exact format as it might change, just that it's not empty)
	t.Logf("BitString string representation: %s", str)
}

func BenchmarkBitString_Set(b *testing.B) {
	bs := NewBitString(1000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs.Set(i%1000000, true)
	}
}

func BenchmarkBitString_Get(b *testing.B) {
	bs := NewBitString(1000000)
	
	// Pre-populate with some data
	for i := 0; i < 1000000; i += 100 {
		bs.Set(i, true)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs.Get(i % 1000000)
	}
}

func BenchmarkBitString_ToCompressedBase64(b *testing.B) {
	bs := NewBitString(131072) // 128KB
	
	// Set every 10th bit to create realistic compression
	for i := 0; i < 131072; i += 10 {
		bs.Set(i, true)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bs.ToCompressedBase64(6)
	}
}

func BenchmarkBitString_FromCompressedBase64(b *testing.B) {
	bs := NewBitString(131072) // 128KB
	
	// Set every 10th bit to create realistic compression
	for i := 0; i < 131072; i += 10 {
		bs.Set(i, true)
	}
	
	encoded, _ := bs.ToCompressedBase64(6)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromCompressedBase64(encoded)
	}
}