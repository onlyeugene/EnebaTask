package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP sets r.RemoteAddr to the caller's IP, which the logger and the
// per-IP rate limiter read.
//
// Behind a proxy (production on Render), the proxy appends the IP it saw to
// X-Forwarded-For, so only the LAST entry is trustworthy; everything before it,
// and X-Real-IP / True-Client-IP, can be set by the client. Locally there is no
// proxy, so all of those headers are ignored and the TCP address is used.
func ClientIP(behindProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := hostOnly(r.RemoteAddr)
			if behindProxy {
				if last := lastForwarded(r.Header.Get("X-Forwarded-For")); last != "" {
					ip = last
				}
			}
			r.RemoteAddr = ip
			next.ServeHTTP(w, r)
		})
	}
}

func lastForwarded(xff string) string {
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")
	last := strings.TrimSpace(parts[len(parts)-1])
	if net.ParseIP(last) == nil {
		return ""
	}
	return last
}

func hostOnly(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
