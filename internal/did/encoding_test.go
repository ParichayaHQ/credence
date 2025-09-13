package did

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestBase58EncodeDecodeEmpty(t *testing.T) {
	// Test empty input
	encoded := base58Encode(nil)
	if encoded != "" {
		t.Errorf("expected empty string for nil input, got %s", encoded)
	}

	encoded = base58Encode([]byte{})
	if encoded != "" {
		t.Errorf("expected empty string for empty input, got %s", encoded)
	}

	decoded, err := base58Decode("")
	if err != nil {
		t.Errorf("unexpected error decoding empty string: %v", err)
	}
	if decoded != nil {
		t.Errorf("expected nil for empty string, got %v", decoded)
	}
}

func TestBase58EncodeDecodeZeros(t *testing.T) {
	// Test leading zeros
	input := []byte{0, 0, 0, 1, 2, 3}
	encoded := base58Encode(input)
	decoded, err := base58Decode(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if !bytes.Equal(input, decoded) {
		t.Errorf("round trip failed: input %v, decoded %v", input, decoded)
	}

	// Should have leading '1' characters for leading zeros
	if len(encoded) < 3 || encoded[:3] != "111" {
		t.Errorf("expected leading 111 for 3 zero bytes, got %s", encoded)
	}
}

func TestBase58EncodeDecodeRandom(t *testing.T) {
	// Test with random data
	for i := 0; i < 100; i++ {
		size := i%50 + 1 // 1 to 50 bytes
		input := make([]byte, size)
		if _, err := rand.Read(input); err != nil {
			t.Fatalf("failed to generate random data: %v", err)
		}

		encoded := base58Encode(input)
		decoded, err := base58Decode(encoded)
		if err != nil {
			t.Fatalf("decode error for input %v: %v", input, err)
		}

		if !bytes.Equal(input, decoded) {
			t.Errorf("round trip failed for input %v: encoded %s, decoded %v", input, encoded, decoded)
		}
	}
}

func TestBase58DecodeInvalidCharacters(t *testing.T) {
	invalidInputs := []string{
		"0",     // '0' is not in base58 alphabet
		"O",     // 'O' is not in base58 alphabet  
		"I",     // 'I' is not in base58 alphabet
		"l",     // 'l' is not in base58 alphabet
		"hello0world",
		"test+data",
		"test/data",
	}

	for _, input := range invalidInputs {
		_, err := base58Decode(input)
		if err == nil {
			t.Errorf("expected error for invalid input %s", input)
		}
	}
}

func TestIsValidBase58(t *testing.T) {
	validTests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"1", true},
		{"123456789", true},
		{"ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz", true},
		{"z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK", true},
		{"0", false},  // '0' not in alphabet
		{"O", false},  // 'O' not in alphabet
		{"I", false},  // 'I' not in alphabet
		{"l", false},  // 'l' not in alphabet
		{"hello0", false},
		{"test+", false},
		{"test/", false},
	}

	for _, tt := range validTests {
		result := IsValidBase58(tt.input)
		if result != tt.expected {
			t.Errorf("IsValidBase58(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestMultibaseEncode(t *testing.T) {
	input := []byte{1, 2, 3, 4, 5}

	// Test base58btc encoding (z prefix)
	encoded := multibaseEncode('z', input)
	expected := "z" + base58Encode(input)
	if encoded != expected {
		t.Errorf("expected %s, got %s", expected, encoded)
	}

	// Test fallback for unknown encoding
	encoded = multibaseEncode('x', input)
	expected = base58Encode(input) // fallback
	if encoded != expected {
		t.Errorf("expected fallback %s, got %s", expected, encoded)
	}
}

func TestMultibaseDecode(t *testing.T) {
	input := []byte{1, 2, 3, 4, 5}
	base58Encoded := base58Encode(input)

	tests := []struct {
		name     string
		encoded  string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "base58btc with z prefix",
			encoded:  "z" + base58Encoded,
			expected: input,
			wantErr:  false,
		},
		{
			name:     "fallback without prefix",
			encoded:  base58Encoded,
			expected: input,
			wantErr:  false,
		},
		{
			name:     "empty string",
			encoded:  "",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := multibaseDecode(tt.encoded)
			
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(decoded, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, decoded)
			}
		})
	}
}

func TestBase58KnownVectors(t *testing.T) {
	// Test with known vectors to ensure compatibility
	vectors := []struct {
		decoded []byte
		encoded string
	}{
		{[]byte{0}, "1"},
		{[]byte{0, 0}, "11"},
		{[]byte{0, 0, 0}, "111"},
		{[]byte{1}, "2"},
		{[]byte{255}, "5Q"},
		{[]byte{0, 255}, "15Q"},
		{[]byte{255, 255}, "LUv"},
		{[]byte("Hello World!"), "2NEpo7TZRRrLZSi2U"},
		{[]byte("The quick brown fox jumps over the lazy dog."), "USm3fpXnKG5EUBx2ndxBDMPVciP5hGey2Jh4NDv6gmeo1LkMeiKrLJUUBk6Z"},
	}

	for i, v := range vectors {
		encoded := base58Encode(v.decoded)
		if encoded != v.encoded {
			t.Errorf("vector %d encode: expected %s, got %s", i, v.encoded, encoded)
		}

		decoded, err := base58Decode(v.encoded)
		if err != nil {
			t.Errorf("vector %d decode error: %v", i, err)
		}

		if !bytes.Equal(decoded, v.decoded) {
			t.Errorf("vector %d decode: expected %v, got %v", i, v.decoded, decoded)
		}
	}
}

func BenchmarkBase58Encode(b *testing.B) {
	data := make([]byte, 32) // Typical key size
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base58Encode(data)
	}
}

func BenchmarkBase58Decode(b *testing.B) {
	data := make([]byte, 32)
	rand.Read(data)
	encoded := base58Encode(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base58Decode(encoded)
	}
}