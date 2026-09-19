package main

import (
	"net"
	"net/http"
	"strings"
)

func isValidCountry(c string) bool {
	allowed := map[string]bool{
		"RU": true, "UA": true, "BY": true, "KZ": true, "KG": true,
		"UZ": true, "TM": true, "AZ": true, "AM": true, "GE": true,
		"MD": true, "LV": true, "LT": true, "EE": true, "DE": true,
		"US": true, "JP": true, "KR": true, "CN": true, "BR": true,
		"RS": true, "BG": true, "OTHER": true,
	}
	return allowed[c]
}

func sanitizeInput(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if r < 32 || r == '\u200b' || r == '\u200c' || r == '\u200d' || r == '\ufeff' {
			return -1
		}
		return r
	}, s)
}

// clientIP returns the real client IP.
//
// Forwarded headers (X-Forwarded-For / X-Real-IP) are only trusted when the TCP
// peer is a loopback/private address, i.e. our own reverse proxy. Direct
// clients cannot spoof these headers to rotate rate-limit buckets, and proxied
// clients are not all squeezed into a single shared bucket: we take the
// rightmost XFF entry, which was appended by our own trusted edge.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				if candidate := strings.TrimSpace(parts[i]); candidate != "" {
					return candidate
				}
			}
		}
		if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
			return xri
		}
	}

	return host
}
