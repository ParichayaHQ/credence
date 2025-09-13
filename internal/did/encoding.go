package did

import (
	"errors"
	"math/big"
)

// Base58 alphabet used by Bitcoin and similar systems
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	base58Map [128]int8
	bigBase   = big.NewInt(58)
	bigZero   = big.NewInt(0)
)

func init() {
	// Initialize the base58 character map
	for i := range base58Map {
		base58Map[i] = -1
	}
	for i, c := range base58Alphabet {
		base58Map[c] = int8(i)
	}
}

// base58Encode encodes bytes to base58 string
func base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	
	// Count leading zeros
	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}
	
	// Convert to big integer
	num := new(big.Int).SetBytes(input)
	
	// Build result string
	var result []byte
	for num.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, bigBase, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}
	
	// Add leading zeros as '1' characters
	for i := 0; i < zeros; i++ {
		result = append(result, '1')
	}
	
	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	
	return string(result)
}

// base58Decode decodes base58 string to bytes
func base58Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}
	
	// Count leading '1' characters
	zeros := 0
	for zeros < len(input) && input[zeros] == '1' {
		zeros++
	}
	
	// Convert string to big integer
	num := big.NewInt(0)
	multi := big.NewInt(1)
	
	for i := len(input) - 1; i >= zeros; i-- {
		if int(input[i]) >= len(base58Map) || base58Map[input[i]] < 0 {
			return nil, errors.New("invalid base58 character")
		}
		
		temp := big.NewInt(int64(base58Map[input[i]]))
		temp.Mul(temp, multi)
		num.Add(num, temp)
		multi.Mul(multi, bigBase)
	}
	
	// Convert to bytes
	result := num.Bytes()
	
	// Add leading zeros
	if zeros > 0 {
		leadingZeros := make([]byte, zeros)
		result = append(leadingZeros, result...)
	}
	
	return result, nil
}

// IsValidBase58 checks if a string is valid base58
func IsValidBase58(s string) bool {
	for _, c := range s {
		if int(c) >= len(base58Map) || base58Map[c] < 0 {
			return false
		}
	}
	return true
}

// multibaseEncode encodes bytes with multibase prefix
func multibaseEncode(encoding byte, data []byte) string {
	switch encoding {
	case 'z': // base58btc
		return "z" + base58Encode(data)
	default:
		return base58Encode(data) // fallback
	}
}

// multibaseDecode decodes multibase string
func multibaseDecode(encoded string) ([]byte, error) {
	if len(encoded) == 0 {
		return nil, errors.New("empty multibase string")
	}
	
	prefix := encoded[0]
	data := encoded[1:]
	
	switch prefix {
	case 'z': // base58btc
		return base58Decode(data)
	default:
		return base58Decode(encoded) // fallback for legacy support
	}
}