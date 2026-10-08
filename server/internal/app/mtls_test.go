package app_test

import (
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/app"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mtls"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mtls/mtlstest"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func TestMTLSGuardsEverythingButHealth(t *testing.T) {
	const serverName = "afc-go-prod.nomad.internal"
	ca := mtlstest.NewCA(t, "internal")
	dir := t.TempDir()
	s, err := mtls.New(mtls.Config{
		CertFile:       mtlstest.WriteFile(t, dir, "server.pem", ca.Issue(t, serverName, x509.ExtKeyUsageServerAuth)),
		ClientCAFile:   mtlstest.WriteFile(t, dir, "ca.pem", ca.PEM),
		AllowedClients: []string{"nginx.internal"},
	})
	require.NoError(t, err)

	conf := testConfig()
	conf.MTLS = s
	a := app.Build(conf, app.Stores{}, uploadtest.New(), mail.NewMailer(mail.Config{}))
	t.Cleanup(a.Stop)

	srv := httptest.NewUnstartedServer(a.Echo)
	srv.TLS = s.TLSConfig()
	srv.StartTLS()
	t.Cleanup(srv.Close)

	anonymous := mtlstest.Client(t, ca, serverName, nil)
	proxy := mtlstest.Client(t, ca, serverName, ca.Issue(t, "nginx.internal", x509.ExtKeyUsageClientAuth))

	for _, path := range []string{"/api/health", "/api/v1/health", "/api/health/"} {
		status, err := mtlstest.Get(t, anonymous, srv.URL+path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, status, path)
	}
	for _, path := range []string{"/", "/api/v1/auth/me", "/does-not-exist"} {
		status, err := mtlstest.Get(t, anonymous, srv.URL+path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, status, path)

		status, err = mtlstest.Get(t, proxy, srv.URL+path)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusForbidden, status, path)
	}
}
