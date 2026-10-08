// Package mtls serves the app over TLS and only accepts requests from proxies
// that present a client certificate issued by the internal CA.
//
// Certificates are short-lived and re-rendered by Nomad, so they are re-read
// on SIGHUP rather than only at start-up.
package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync/atomic"
	"syscall"

	"github.com/labstack/echo/v4"
)

type (
	// Config says where the certificates live and which clients may connect.
	Config struct {
		// CertFile holds the server certificate (and any intermediates).
		CertFile string
		// KeyFile holds the server's private key; empty means it is in CertFile.
		KeyFile string
		// ClientCAFile holds the CA certificates client certificates must chain to.
		ClientCAFile string
		// AllowedClients are the DNS names a client certificate must carry one of.
		AllowedClients []string
	}

	// Server holds the current certificates and checks incoming requests.
	Server struct {
		conf  Config
		state atomic.Pointer[state]
	}

	state struct {
		cert      *tls.Certificate
		clientCAs *x509.CertPool
	}
)

// Enabled reports whether mTLS has been configured at all.
func (c Config) Enabled() bool {
	return c.CertFile != ""
}

// New loads the certificates in conf.
func New(conf Config) (*Server, error) {
	if conf.KeyFile == "" {
		conf.KeyFile = conf.CertFile
	}
	if conf.ClientCAFile == "" {
		return nil, errors.New("mtls: a client CA file is required")
	}
	if len(conf.AllowedClients) == 0 {
		return nil, errors.New("mtls: at least one allowed client name is required")
	}
	s := &Server{conf: conf}
	if err := s.Reload(); err != nil {
		return nil, err
	}
	return s, nil
}

// Reload re-reads the certificates from disk. On error the previous ones stay in use.
func (s *Server) Reload() error {
	cert, err := tls.LoadX509KeyPair(s.conf.CertFile, s.conf.KeyFile)
	if err != nil {
		return fmt.Errorf("mtls: failed to load server certificate: %w", err)
	}
	caPEM, err := os.ReadFile(s.conf.ClientCAFile)
	if err != nil {
		return fmt.Errorf("mtls: failed to read client CA file: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return fmt.Errorf("mtls: no certificates found in %s", s.conf.ClientCAFile)
	}
	s.state.Store(&state{cert: &cert, clientCAs: pool})
	return nil
}

// TLSConfig returns a config that always uses the most recently loaded certificates.
//
// Client certificates are verified when given but not required at the
// handshake, so the health check can connect without one; Middleware enforces
// them for everything else.
func (s *Server) TLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			st := s.state.Load()
			return &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{*st.cert},
				ClientAuth:   tls.VerifyClientCertIfGiven,
				ClientCAs:    st.clientCAs,
			}, nil
		},
	}
}

// ReloadOnSIGHUP reloads the certificates whenever the process gets SIGHUP,
// which is what Nomad sends when it re-renders them. Call stop to stop listening.
func (s *Server) ReloadOnSIGHUP() (stop func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-signals:
				if err := s.Reload(); err != nil {
					slog.Error(fmt.Sprintf("failed to reload mTLS certificates, keeping the old ones: %+v", err))
				} else {
					slog.Info("reloaded mTLS certificates")
				}
			case <-done:
				return
			}
		}
	}()
	return func() {
		signal.Stop(signals)
		close(done)
	}
}

// Middleware rejects requests that did not present a verified client
// certificate for one of the allowed names. Requests for skipPaths (the health
// check) are let through so the orchestrator can probe without a certificate.
func (s *Server) Middleware(skipPaths ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			if slices.Contains(skipPaths, r.URL.Path) || s.allowed(r) {
				return next(c)
			}
			return echo.NewHTTPError(http.StatusForbidden, "client certificate required")
		}
	}
}

func (s *Server) allowed(r *http.Request) bool {
	if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 || len(r.TLS.VerifiedChains[0]) == 0 {
		return false
	}
	leaf := r.TLS.VerifiedChains[0][0]
	return slices.ContainsFunc(leaf.DNSNames, func(name string) bool {
		return slices.Contains(s.conf.AllowedClients, name)
	})
}
