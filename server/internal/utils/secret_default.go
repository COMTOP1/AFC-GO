//go:build !goexperiment.runtimesecret

package utils

// WithSecret runs f directly. Built without GOEXPERIMENT=runtimesecret,
// so no memory erasure is performed; see secret_runtimesecret.go.
func WithSecret(f func()) {
	f()
}
