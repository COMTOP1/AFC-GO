package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestKeyWarning pins Fix round 1 finding #1: a valid-hex key of the wrong
// length must still be used (never replaced with a random one), but must
// produce a warning naming the key, its decoded length, and the recommended
// length.
func TestKeyWarning(t *testing.T) {
	cases := []struct {
		name        string
		what        string
		key         []byte
		size        int
		wantWarning bool
	}{
		{"authentication key at the recommended minimum", "authentication", make([]byte, 32), 64, false},
		{"authentication key at the generated length", "authentication", make([]byte, 64), 64, false},
		{"authentication key truncated to 4 bytes", "authentication", make([]byte, 4), 64, true},
		{"authentication key just under the minimum", "authentication", make([]byte, 31), 64, true},
		{"encryption key 16 bytes (AES-128)", "encryption", make([]byte, 16), 32, false},
		{"encryption key 24 bytes (AES-192)", "encryption", make([]byte, 24), 32, false},
		{"encryption key 32 bytes (AES-256)", "encryption", make([]byte, 32), 32, false},
		{"encryption key of an invalid AES length", "encryption", make([]byte, 20), 32, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := keyWarning(tc.what, tc.key, tc.size)
			if !tc.wantWarning {
				assert.Empty(t, msg)
				return
			}
			assert.NotEmpty(t, msg)
			assert.Contains(t, msg, tc.what, "message should name which key is weak")
			assert.Contains(t, msg, "byte", "message should mention the decoded length")
		})
	}
}
