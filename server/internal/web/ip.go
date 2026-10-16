package web

import (
	"fmt"
	"net"
	"strings"

	"github.com/labstack/echo/v4"
)

// ParseCIDRs parses a comma-separated list of CIDR ranges, e.g. the
// TRUSTED_PROXIES setting "10.20.0.0/16".
func ParseCIDRs(raw string) ([]*net.IPNet, error) {
	var out []*net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(part)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", part, err)
		}
		out = append(out, ipNet)
	}
	return out, nil
}

// IPExtractor decides which address c.RealIP() reports. X-Forwarded-For is
// believed only when the request comes from one of the trusted proxy
// ranges; otherwise anyone could pick their own IP, dodging the per-IP
// sign-in limit and faking the address in logs. With no trusted ranges the
// connecting address is used and the header is ignored.
func IPExtractor(trusted []*net.IPNet) echo.IPExtractor {
	if len(trusted) == 0 {
		return echo.ExtractIPDirect()
	}
	opts := []echo.TrustOption{echo.TrustLoopback(false), echo.TrustLinkLocal(false), echo.TrustPrivateNet(false)}
	for _, n := range trusted {
		opts = append(opts, echo.TrustIPRange(n))
	}
	return echo.ExtractIPFromXFFHeader(opts...)
}
