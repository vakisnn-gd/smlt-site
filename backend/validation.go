package main

import (
	"net"
	"net/http"
	"net/url"
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

// normalizeYouTubeID accepts both a raw 11-character ID and the common
// YouTube URL forms used when sharing a video. Events are stored only with the
// canonical ID so the embed URL cannot become /embed/https://....
func normalizeYouTubeID(value string) (string, bool) {
	raw := strings.TrimSpace(value)
	if isYouTubeID(raw) {
		return raw, true
	}

	parseValue := raw
	if !strings.Contains(parseValue, "://") {
		parseValue = "https://" + parseValue
	}
	u, err := url.Parse(parseValue)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	var id string
	switch host {
	case "youtu.be":
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[0]
		}
	case "youtube.com", "m.youtube.com", "music.youtube.com", "youtube-nocookie.com":
		if u.Path == "/watch" {
			id = u.Query().Get("v")
		} else {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) >= 2 && (parts[0] == "embed" || parts[0] == "shorts" || parts[0] == "live" || parts[0] == "v") {
				id = parts[1]
			}
		}
	}
	return id, isYouTubeID(id)
}

func isYouTubeID(value string) bool {
	if len(value) != 11 {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
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
		if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(xri) != nil {
			return xri
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				if candidate := strings.TrimSpace(parts[i]); net.ParseIP(candidate) != nil {
					return candidate
				}
			}
		}
	}

	return host
}
