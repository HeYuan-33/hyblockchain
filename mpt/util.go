package mpt

import (
	"crypto/sha256"
	"encoding/hex"
)

// StringToNibbles converts a string into a slice of nibbles (4-bit units).
func StringToNibbles(str string) []uint8 {
	bytes := []byte(str)
	nibbles := make([]uint8, 0, len(bytes)*2)
	for _, b := range bytes {
		nibbles = append(nibbles, b>>4)
		nibbles = append(nibbles, b&0x0F)
	}
	return nibbles
}

// NibblesToString converts a slice of nibbles back to its original string.
// This assumes that the input nibbles are even-length.
func NibblesToString(nibbles []uint8) string {
	if len(nibbles)%2 != 0 {
		// ignore the last half-byte if odd length
		nibbles = nibbles[:len(nibbles)-1]
	}
	bytes := make([]byte, 0, len(nibbles)/2)
	for i := 0; i < len(nibbles); i += 2 {
		b := (nibbles[i] << 4) | nibbles[i+1]
		bytes = append(bytes, b)
	}
	return string(bytes)
}

// HashString returns the SHA-256 hash of a given string as a hex-encoded string.
func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}
