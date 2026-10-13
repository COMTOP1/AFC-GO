// Package mtlstest makes throwaway CAs and certificates for mTLS tests.
package mtlstest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CA is a self-signed certificate authority.
type CA struct {
	Cert *x509.Certificate
	// PEM is the CA certificate, PEM-encoded.
	PEM []byte
	key *ecdsa.PrivateKey
}

// NewCA makes a CA valid for the next hour.
func NewCA(t *testing.T, name string) CA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return CA{Cert: cert, PEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), key: key}
}

// Issue returns a PEM bundle of a certificate for name followed by its key.
func (ca CA) Issue(t *testing.T, name string, usage x509.ExtKeyUsage) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: name},
		DNSNames:     []string{name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{usage},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.Cert, &key.PublicKey, ca.key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	return append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})...)
}

// WriteFile writes data to name in dir and returns its path.
func WriteFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

// Client returns an HTTP client that trusts ca for serverName and, when
// clientPEM is non-nil, always presents it as its client certificate.
func Client(t *testing.T, ca CA, serverName string, clientPEM []byte) *http.Client {
	t.Helper()
	roots := x509.NewCertPool()
	roots.AddCert(ca.Cert)
	conf := &tls.Config{RootCAs: roots, ServerName: serverName, MinVersion: tls.VersionTLS12}
	if clientPEM != nil {
		cert, err := tls.X509KeyPair(clientPEM, clientPEM)
		require.NoError(t, err)
		// Always send it, even when the server's CA list doesn't match, so the
		// server is what rejects a certificate from the wrong CA.
		conf.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &cert, nil
		}
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: conf}}
}

// Get requests url and returns the status code.
func Get(t *testing.T, c *http.Client, url string) (int, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, http.NoBody)
	require.NoError(t, err)
	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	return res.StatusCode, nil
}
