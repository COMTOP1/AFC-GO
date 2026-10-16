package web

import (
	"fmt"
	"net"
	"strings"

	"github.com/labstack/echo/v4"
)

// ParseCIDRs parses a comma-separated list of IP addresses and CIDR ranges,
// e.g. the TRUSTED_PROXIES setting "10.20.1.10,10.20.1.11" or
// "10.20.0.0/16". A bare address means just that one host.
func ParseCIDRs(raw string) ([]*net.IPNet, error) {
	var out []*net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if ip := net.ParseIP(part); ip != nil {
			bits := 128
			if ip.To4() != nil {
				ip, bits = ip.To4(), 32
			}
			out = append(out, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}
		_, ipNet, err := net.ParseCIDR(part)
		if err != nil {
			return nil, fmt.Errorf("invalid IP address or CIDR range %q: %w", part, err)
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
