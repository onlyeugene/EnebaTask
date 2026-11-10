package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name        string
		behindProxy bool
		remoteAddr  string
		headers     map[string]string
		want        string
	}{
		{
			name:       "local: spoofed headers are ignored",
			remoteAddr: "203.0.113.7:51234",
			headers: map[string]string{
				"X-Forwarded-For": "1.1.1.1",
				"X-Real-IP":       "2.2.2.2",
				"True-Client-IP":  "3.3.3.3",
			},
			want: "203.0.113.7",
		},
		{
			name:        "proxy: client-sent entries before the proxy's are ignored",
			behindProxy: true,
			remoteAddr:  "10.0.0.5:443",
			headers: map[string]string{
				"X-Forwarded-For": "1.1.1.1, 9.9.9.9, 198.51.100.20",
				"X-Real-IP":       "2.2.2.2",
				"True-Client-IP":  "3.3.3.3",
			},
			want: "198.51.100.20",
		},
		{
			name:        "proxy: no header falls back to the TCP address",
			behindProxy: true,
			remoteAddr:  "10.0.0.5:443",
			headers:     map[string]string{"X-Real-IP": "2.2.2.2"},
			want:        "10.0.0.5",
		},
		{
			name:        "proxy: garbage last entry falls back to the TCP address",
			behindProxy: true,
			remoteAddr:  "10.0.0.5:443",
			headers:     map[string]string{"X-Forwarded-For": "1.1.1.1, not-an-ip"},
			want:        "10.0.0.5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			h := ClientIP(tc.behindProxy)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.RemoteAddr
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			h.ServeHTTP(httptest.NewRecorder(), req)

			if got != tc.want {
				t.Fatalf("ip = %q, want %q", got, tc.want)
			}
		})
	}
}
