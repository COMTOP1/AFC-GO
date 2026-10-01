//go:build goexperiment.runtimesecret

package utils

import "runtime/secret"

// WithSecret runs f, erasing any registers, stack space, and heap
// allocations it used once it returns. It is used to bound how long
// plaintext passwords and derived key material remain in memory
// during hashing.
//
// Only active when built with GOEXPERIMENT=runtimesecret; see
// secret_default.go for the fallback used otherwise.
func WithSecret(f func()) {
	secret.Do(f)
}
