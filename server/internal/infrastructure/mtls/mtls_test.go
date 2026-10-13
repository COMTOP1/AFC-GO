package mtls_test

import (
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mtls"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mtls/mtlstest"
)

const serverName = "afc-go-prod.nomad.internal"

// startServer serves an echo app over mTLS and returns its URL, the Server
// and the path of the server certificate file.
func startServer(t *testing.T, ca mtlstest.CA) (string, *mtls.Server, string) {
	t.Helper()
	dir := t.TempDir()
	certFile := mtlstest.WriteFile(t, dir, "server.pem", ca.Issue(t, serverName, x509.ExtKeyUsageServerAuth))
	s, err := mtls.New(mtls.Config{
		CertFile:       certFile,
		ClientCAFile:   mtlstest.WriteFile(t, dir, "ca.pem", ca.PEM),
		AllowedClients: []string{"nginx.internal"},
	})
	require.NoError(t, err)

	e := echo.New()
	e.Pre(s.Middleware("/api/health"))
	e.GET("/*", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	srv := httptest.NewUnstartedServer(e)
	srv.TLS = s.TLSConfig()
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv.URL, s, certFile
}

func TestMiddleware(t *testing.T) {
	ca := mtlstest.NewCA(t, "internal")
	url, _, _ := startServer(t, ca)

	tests := []struct {
		name   string
		client []byte
		path   string
		want   int
	}{
		{"allowed client", ca.Issue(t, "nginx.internal", x509.ExtKeyUsageClientAuth), "/", http.StatusOK},
		{"no client certificate", nil, "/", http.StatusForbidden},
		{"client not in allow list", ca.Issue(t, "other.nomad.internal", x509.ExtKeyUsageClientAuth), "/", http.StatusForbidden},
		{"health check without certificate", nil, "/api/health", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := mtlstest.Get(t, mtlstest.Client(t, ca, serverName, tt.client), url+tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, status)
		})
	}
}

func TestClientFromOtherCAFailsHandshake(t *testing.T) {
	ca := mtlstest.NewCA(t, "internal")
	url, _, _ := startServer(t, ca)

	other := mtlstest.NewCA(t, "someone else")
	clientPEM := other.Issue(t, "nginx.internal", x509.ExtKeyUsageClientAuth)
	_, err := mtlstest.Get(t, mtlstest.Client(t, ca, serverName, clientPEM), url+"/")
	require.Error(t, err)
}

func TestReloadPicksUpNewCertificate(t *testing.T) {
	oldCA := mtlstest.NewCA(t, "old")
	url, s, certFile := startServer(t, oldCA)

	newCA := mtlstest.NewCA(t, "new")
	require.NoError(t, os.WriteFile(certFile, newCA.Issue(t, serverName, x509.ExtKeyUsageServerAuth), 0o600))
	require.NoError(t, s.Reload())

	// The client CA file was not changed, so the client still uses the old CA.
	clientPEM := oldCA.Issue(t, "nginx.internal", x509.ExtKeyUsageClientAuth)
	_, err := mtlstest.Get(t, mtlstest.Client(t, oldCA, serverName, clientPEM), url+"/")
	require.Error(t, err, "server should no longer present the old certificate")

	status, err := mtlstest.Get(t, mtlstest.Client(t, newCA, serverName, clientPEM), url+"/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
}

func TestReloadKeepsOldCertificateOnError(t *testing.T) {
	ca := mtlstest.NewCA(t, "internal")
	url, s, certFile := startServer(t, ca)

	require.NoError(t, os.WriteFile(certFile, []byte("not a certificate"), 0o600))
	require.Error(t, s.Reload())

	clientPEM := ca.Issue(t, "nginx.internal", x509.ExtKeyUsageClientAuth)
	status, err := mtlstest.Get(t, mtlstest.Client(t, ca, serverName, clientPEM), url+"/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
}

func TestNewValidatesConfig(t *testing.T) {
	_, err := mtls.New(mtls.Config{CertFile: "x", AllowedClients: []string{"nginx.internal"}})
	require.Error(t, err)
	_, err = mtls.New(mtls.Config{CertFile: "x", ClientCAFile: "y"})
	require.Error(t, err)
}
