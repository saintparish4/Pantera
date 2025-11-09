package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// KeyLength is the total length of the API key (prefix + random part)
	KeyLength = 32

	// PrefixLive is the prefix for live API keys
	PrefixLive = "hm_live_"

	// PrefixTest is the prefix for test API keys
	PrefixTest = "hm_test_"
)

// GenerateAPIKey generates a new API key with the specified prefix
func GenerateAPIKey(isLive bool) (string, error) {
	prefix := PrefixTest
	if isLive {
		prefix = PrefixLive
	}

	// Generate random bytes for the key
	randomBytes := make([]byte, 16) // 16 bytes = 32 hex chars
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Convert to hex string
	randomPart := hex.EncodeToString(randomBytes)

	// Combine prefix and random part
	key := prefix + randomPart

	return key, nil
}

// HashAPIKey creates a SHA-256 hash of an API key
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// GetKeyPrefix extracts the display prefix from an API key
// Returns the first 20 characters for display purposes
func GetKeyPrefix(key string) string {
	if len(key) < 20 {
		return key
	}
	return key[:20]
}

// ValidateAPIKeyFormat checks if an API key has the correct format
func ValidateAPIKeyFormat(key string) error {
	// Check length
	if len(key) != len(PrefixLive)+32 && len(key) != len(PrefixTest)+32 {
		return fmt.Errorf("invalid key length")
	}

	// Check prefix
	if !strings.HasPrefix(key, PrefixLive) && !strings.HasPrefix(key, PrefixTest) {
		return fmt.Errorf("invalid key prefix")
	}

	// Check that the rest is valid hex
	prefix := ""
	if strings.HasPrefix(key, PrefixLive) {
		prefix = PrefixLive
	} else {
		prefix = PrefixTest
	}

	hexPart := key[len(prefix):]
	if _, err := hex.DecodeString(hexPart); err != nil {
		return fmt.Errorf("invalid key format: not valid hex")
	}

	return nil
}

// IsLiveKey checks if an API key is a live key (vs test key)
func IsLiveKey(key string) bool {
	return strings.HasPrefix(key, PrefixLive)
}

// MaskAPIKey returns a masked version of the API key for display
// Example: hm_live_abc123... -> hm_live_abc1****************
func MaskAPIKey(key string) string {
	if len(key) < 20 {
		return key
	}

	prefix := key[:12] // Show first 12 chars (prefix + few random chars)
	masked := strings.Repeat("*", len(key)-12)

	return prefix + masked
}
