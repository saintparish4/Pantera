package repository

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		isLive bool
		prefix string
	}{
		{
			name:   "generate live key",
			isLive: true,
			prefix: PrefixLive,
		},
		{
			name:   "generate test key",
			isLive: false,
			prefix: PrefixTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateAPIKey(tt.isLive)
			if err != nil {
				t.Fatalf("failed to generate API key: %v", err)
			}

			// Check prefix
			if !strings.HasPrefix(key, tt.prefix) {
				t.Errorf("expected prefix %s, got %s", tt.prefix, key[:len(tt.prefix)])
			}

			// Check length
			expectedLength := len(tt.prefix) + 32 // prefix + 32 hex chars
			if len(key) != expectedLength {
				t.Errorf("expected length %d, got %d", expectedLength, len(key))
			}

			// Verify it's valid format
			if err := ValidateAPIKeyFormat(key); err != nil {
				t.Errorf("generated key has invalid format: %v", err)
			}
		})
	}
}

func TestGenerateAPIKey_Uniqueness(t *testing.T) {
	// Generate multiple keys and ensure they're unique
	keys := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		key, err := GenerateAPIKey(true)
		if err != nil {
			t.Fatalf("failed to generate API key: %v", err)
		}

		if keys[key] {
			t.Errorf("duplicate key generated: %s", key)
		}

		keys[key] = true
	}

	if len(keys) != iterations {
		t.Errorf("expected %d unique keys, got %d", iterations, len(keys))
	}
}

func TestHashAPIKey(t *testing.T) {
	key := "hm_live_abc123def456abc123def456abc12345"

	hash1 := HashAPIKey(key)
	hash2 := HashAPIKey(key)

	// Same key should produce same hash
	if hash1 != hash2 {
		t.Error("same key produced different hashes")
	}

	// Hash should be 64 hex characters (SHA-256)
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}

	// Different key should produce different hash
	differentKey := "hm_live_xyz789xyz789xyz789xyz789xyz78"
	hash3 := HashAPIKey(differentKey)

	if hash1 == hash3 {
		t.Error("different keys produced same hash")
	}
}

func TestGetKeyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "full key",
			key:      "hm_live_abc123def456abc123def456abc12345",
			expected: "hm_live_abc123def456",
		},
		{
			name:     "short key",
			key:      "short",
			expected: "short",
		},
		{
			name:     "exactly 20 chars",
			key:      "12345678901234567890",
			expected: "12345678901234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetKeyPrefix(tt.key)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateAPIKeyFormat(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid live key",
			key:     "hm_live_abc123def456abc123def456abc12345",
			wantErr: false,
		},
		{
			name:    "valid test key",
			key:     "hm_test_abc123def456abc123def456abc12345",
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			key:     "invalid_abc123def456abc123def456abc12345",
			wantErr: true,
		},
		{
			name:    "too short",
			key:     "hm_live_abc",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			key:     "hm_live_zzz123def456abc123def456abc12345",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKeyFormat(tt.key)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsLiveKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "live key",
			key:      "hm_live_abc123def456abc123def456abc12345",
			expected: true,
		},
		{
			name:     "test key",
			key:      "hm_test_abc123def456abc123def456abc12345",
			expected: false,
		},
		{
			name:     "invalid key",
			key:      "invalid_key",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLiveKey(tt.key)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "full key",
			key:      "hm_live_abc123def456abc123def456abc12345",
			expected: "hm_live_abc1************************",
		},
		{
			name:     "test key",
			key:      "hm_test_abc123def456abc123def456abc12345",
			expected: "hm_test_abc1************************",
		},
		{
			name:     "short key",
			key:      "short",
			expected: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskAPIKey(tt.key)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}

			// Verify masked portion
			if len(tt.key) >= 20 {
				if !strings.Contains(result, "*") {
					t.Error("expected masked key to contain asterisks")
				}

				if !strings.HasPrefix(result, tt.key[:12]) {
					t.Error("expected masked key to preserve first 12 characters")
				}
			}
		})
	}
}

func BenchmarkGenerateAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateAPIKey(true)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashAPIKey(b *testing.B) {
	key := "hm_live_abc123def456abc123def456abc12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashAPIKey(key)
	}
}
